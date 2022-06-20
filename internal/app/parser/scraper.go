package parser

import (
	// "encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	json "github.com/buger/jsonparser"
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

func getItemPriceFromAPI(itemArtc string) (float64, error) {
	url := "https://www.fishersci.com/shop/products/service/pricing"
	resp, err := http.PostForm(url, neturl.Values{"partNumber": {itemArtc}})
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	priceStr, err := json.GetString(body, "priceAndAvailability", itemArtc, "[0]", "price")
	if err != nil {
		return 0, err
	}
	price, err := strconv.ParseFloat(strings.ReplaceAll(priceStr, `$`, ""), 64)
	if err != nil {
		return 0, err
	}
	/*
		priceStruct := Price{}
		err = json.Unmarshal(body, &priceStruct)
		if err != nil {
			return "", err
		}
			price := priceStruct.PriceAndAvailability.PartNumber[0].Price
			if price == "" {
				return "", errors.New("internal error")
			}
	*/
	return price, nil
}

// (secondary)
func getSingleItemData(doc *goquery.Document) (*Data, error) {
	available := true
	selMain := doc.Find("div.productSelectors")

	reg := regexp.MustCompile(`[[:^ascii:]]`)

	label := strings.ReplaceAll(
		reg.ReplaceAllLiteralString(
			strings.TrimSpace(
				selMain.Find("h1").Contents().First().Text()), " "), "  ", " ")
	if label == "" {
		return nil, errors.New("label not found")
	}

	descript := "none"
	artc, exist := selMain.Find("div.glyphs_html_container").Attr("data-partnumber")
	if !exist {
		return nil, errors.New("internal error")
	}

	manufact := ""
	doc.Find(".spec_table").Find("tr").EachWithBreak(
		func(index int, item *goquery.Selection) bool {
			td := item.Find("td.bold")
			if td.Contents().Text() == "Product Line" {
				manufact = strings.TrimSpace(
					reg.ReplaceAllLiteralString(
						td.Parent().Find("td").Last().Contents().Text(), ""))
				return false
			}
			return true
		})
	if manufact == "" {
		return nil, errors.New("manufacturer not found")
	}
	manufact = strings.Join([]string{manufact, artc}, " ")

	price, err := getItemPriceFromAPI(artc)
	if err != nil {
		price = 0
		available = false
	}
	/*
		selManAndArt := doc.Find("div.block_head>p")

		manufact := reg.ReplaceAllLiteralString(
			selManAndArt.Last().Contents().Text(), "")

		fmt.Println(doc.Html())

		// not found - idk
		fmt.Println(doc.Find("div.block_head").Nodes)

		if manufact == "" {
			return nil, errors.New("manufacturer not found")
		}

		priceStr := reg.ReplaceAllLiteralString(
			doc.Find("div.block_body>span.qa_single_price>b").Contents().Text(), "")
		if priceStr == "" {
			priceStr = "0"
			available = false
		}

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, errors.New("invalid price")
		}

		artc := reg.ReplaceAllLiteralString(
			selManAndArt.First().Contents().Text(), "")
		if artc == "" {
			return nil, errors.New("article not found")
		}

	*/
	data := Data{
		Price:        price,
		Label:        label,
		Article:      artc,
		Description:  descript,
		Manufacturer: manufact,
		Available:    available,
	}
	return &data, nil
}

func getItemData(doc *goquery.Document) (*Data, []string, error) {
	// in stock
	available := true

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

	// single (secondary) item
	singlepage := doc.Find("div.productSelectors") // ?
	if singlepage.Nodes != nil {
		data, err := getSingleItemData(doc)
		if err != nil {
			return nil, nil, err
		}
		return data, nil, nil
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

	selDescAndMan := selMain.Find("div.subhead")

	// default item
	descript := reg.ReplaceAllLiteralString(
		strings.TrimSpace(
			selDescAndMan.Find("p").First().Contents().Text()), "")

	// kit item
	if descript == "" {
		descript = reg.ReplaceAllLiteralString(
			strings.TrimSpace(
				selDescAndMan.Find("div").First().Contents().Text()), "")
	}
	if descript == "" {
		descript = "none" // some items without a description
	}

	// default & kit item
	manufact := strings.ReplaceAll(
		strings.TrimSpace(
			reg.ReplaceAllLiteralString(
				strings.ReplaceAll(
					selDescAndMan.Find("p:nth-of-type(2), p:nth-of-type(3)").
						Contents().Text(), "Manufacturer:", ""), "")), "  ", " ")

	if manufact == "" {
		return nil, nil, errors.New("manufacturer not found")
	}

	selPriceAndArtc := selMain.Find("div.product_sku_options_block")

	priceStr, exist := selPriceAndArtc.Find("label.price>span>span").Attr("content")
	if !exist {
		// not stock
		priceStr = "0"
		available = false
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
		Available:    available,
	}
	return &data, nil, nil
}

func printItemData(data *Data) {
	fmt.Printf("\nitem:\n%s\n%s\n%s\n%.2f\n\n",
		data.Article, data.Label, data.Manufacturer, data.Price)
}

func FisherSciencific(brand string) {
	// "/shop/products/gibco-bottle-weight/A1098801#?keyword=gibco%20bottle"
	// "/shop/products/algimatrix-96-well-3d-culture-system-flat-bottom-microplate-1/12684031"
	debugUrl := "/shop/products/gibco-bottle-weight/A1098801#?keyword=gibco%20bottle"
	debugItemDoc, err := getItemDocument(debugUrl)
	if err != nil {
		log.Fatal(err)
	}
	data, _, err := getItemData(debugItemDoc)
	if err != nil {
		log.Fatal(err)
	}
	printItemData(data)

	/*
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
				fmt.Println(currentItemUrl) // debug
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
						fmt.Println(internalItemUrl) // debug
						data, _, err = getItemData(internalItemDoc)
						if err != nil {
							log.Fatal(err)
						}
						printItemData(data) // out
					}
				} else {
					printItemData(data) // out
				}
			}
		}
	*/
}
