package parser

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getPagesCount(doc *goquery.Document) (int, error) {
	pages, exist := doc.Find("li.hide_in_mobile").Last().Find("a").Attr("href")
	if !exist {
		return 0, errors.New("not found pages")
	}
	return strconv.Atoi(strings.Trim(strings.TrimSpace(pages), `?page=`))
}

func getItemsUrl(doc *goquery.Document) ([]string, error) {
	buffer := make([]string, 0)
	doc.Find("div.search_results>div.search_results_listing>div.row").Each(
		func(index int, item *goquery.Selection) {
			url, exist := item.Find("div.columns>div.block>a").Attr("href")
			if exist {
				buffer = append(buffer, strings.TrimSpace(url))
			}
		})
	if len(buffer) == 0 {
		return nil, errors.New("not found url")
	}
	return buffer, nil
}

func getPageDocument(brand string, page int) (*goquery.Document, error) {
	url := fmt.Sprintf("https://www.fishersci.com/us/en/brands/%s.html?page=%d", brand, page)
	return getHtmlDocument(url)
}

func getItemDocument(item string) (*goquery.Document, error) {
	url := fmt.Sprintf("https://www.fishersci.com%s", item)
	return getHtmlDocument(url)
}

func getHtmlDocument(url string) (*goquery.Document, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("uncorrect ulr: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error status code: %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func AllocateEmptySlice() [][]string {
	slice := make([][]string, 0)
	for i := range slice {
		slice[i] = make([]string, 0)
	}
	return slice
}

func getItemData(doc *goquery.Document) (*Data, error) {
	reg := regexp.MustCompile(`[[:^ascii:]]`)
	selMain := doc.Find("div.product_description_wrapper>div.row>div.columns")

	label := reg.ReplaceAllLiteralString(
		strings.TrimSpace(
			selMain.Find("h1").Contents().First().Text()), "")
	if label == "" {
		return nil, errors.New("label not found")
	}

	selDescAndMan := selMain.Find("div.subhead>p")

	descript := reg.ReplaceAllLiteralString(
		strings.TrimSpace(
			selDescAndMan.First().Contents().Text()), "")
	if descript == "" {
		return nil, errors.New("description not found")
	}

	// need fix (manufacturer ?)
	manufact := strings.ReplaceAll(
		strings.TrimSpace(
			reg.ReplaceAllLiteralString(
				strings.ReplaceAll(
					selDescAndMan.Next().Contents().Text(), "Manufacturer:", ""), "")), "  ", " ")
	if manufact == "" {
		return nil, errors.New("manufacturer not found")
	}

	return nil, nil
}

func FisherSciencific(brand string) {
	doc, err := getPageDocument(brand, 0)
	if err != nil {
		log.Fatal(err)
	}
	pages, err := getPagesCount(doc)
	if err != nil {
		log.Fatal(err)
	}
	for i := 1; i <= pages; i++ {
		page, err := getPageDocument(brand, i)
		if err != nil {
			log.Fatal(err)
		}
		urls, err := getItemsUrl(page)
		if err != nil {
			log.Fatal(err)
		}
		for _, u := range urls {
			fmt.Println(u)
		}
		item, err := getItemDocument(urls[0])
		if err != nil {
			log.Fatal(err)
		}
		getItemData(item)
	}
}
