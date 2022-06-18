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
		return nil, errors.New("not found urls")
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

func getItemData(doc *goquery.Document) (*Data, []string, error) {
	// muiltiple item page
	multipage := doc.Find("div.products_list")
	if multipage.Nodes != nil {
		buffer := make([]string, 0)
		multipage.Find("tbody.itemRowContent").Each(
			func(index int, item *goquery.Selection) {
				url, exist := item.Find("a.chemical_fmly_glyph").Attr("href")
				if exist {
					buffer = append(buffer, url)
				}
			})
		if len(buffer) == 0 {
			return nil, nil, errors.New("not found urls")
		}
		return nil, buffer, nil
	}

	// default item page
	reg := regexp.MustCompile(`[[:^ascii:]]`)
	selMain := doc.Find("div.product_description_wrapper")

	label := strings.ReplaceAll(
		reg.ReplaceAllLiteralString(
			strings.TrimSpace(
				selMain.Find("h1").Contents().First().Text()), " "), "  ", " ")
	if label == "" {
		return nil, nil, errors.New("label not found")
	}

	selDescAndMan := selMain.Find("p")

	descript := reg.ReplaceAllLiteralString(
		strings.TrimSpace(
			selDescAndMan.First().Contents().Text()), "")
	if descript == "" {
		return nil, nil, errors.New("description not found")
	}

	manufact := strings.ReplaceAll(
		strings.TrimSpace(
			reg.ReplaceAllLiteralString(
				strings.ReplaceAll(
					selDescAndMan.Next().Contents().Text(), "Manufacturer:", ""), "")), "  ", " ")
	if manufact == "" {
		return nil, nil, errors.New("manufacturer not found")
	}

	selPriceAndArtc := selMain.Find("div.product_sku_options_block")

	priceStr, exist := selPriceAndArtc.Find("label.price>span>span").Attr("content")
	if !exist {
		return nil, nil, errors.New("price is not available")
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return nil, nil, errors.New("invalid price")
	}

	artc := strings.TrimSpace(
		selPriceAndArtc.Find("span.float_right").Contents().Text())
	if artc == "" {
		return nil, nil, errors.New("article not found")
	}

	data := Data{
		Price:        price,
		Label:        label,
		Article:      artc,
		Description:  descript,
		Manufacturer: manufact,
	}
	return &data, nil, nil
}

func printItemData(data *Data) {
	fmt.Printf("\nitem:\n%s\n%s\n%s\n%.2f\n\n",
		data.Article, data.Label, data.Manufacturer, data.Price)
}

func FisherSciencific(brand string) {
	currentPageDoc, err := getPageDocument(brand, 0)
	if err != nil {
		log.Fatal(err)
	}
	pageCount, err := getPagesCount(currentPageDoc)
	if err != nil {
		log.Fatal(err)
	}
	for i := 1; i <= pageCount; i++ {
		currentPageDoc, err = getPageDocument(brand, i)
		if err != nil {
			log.Fatal(err)
		}
		itemsUrl, err := getItemsUrl(currentPageDoc)
		if err != nil {
			log.Fatal(err)
		}
		for _, currentItemUrl := range itemsUrl {
			currentItemDoc, err := getItemDocument(currentItemUrl)
			if err != nil {
				log.Fatal(err)
			}
			data, multipleItemsUrl, err := getItemData(currentItemDoc)
			if err != nil {
				log.Fatal(err)
			}
			if len(multipleItemsUrl) != 0 {
				for _, internalItemUrl := range multipleItemsUrl {
					internalItemDoc, err := getItemDocument(internalItemUrl)
					if err != nil {
						log.Fatal(err)
					}
					data, _, err = getItemData(internalItemDoc)
					if err != nil {
						log.Printf("\n%s\n", internalItemUrl)
					}
					printItemData(data)
				}
			} else {
				printItemData(data)
			}
		}

	}
}
