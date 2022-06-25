package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/buger/jsonparser"
)

func PrettyPrint(data interface{}) {
	pb, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("%s\n\n", pb)
}

type Parser struct {
	Brand string
}

func NewParser(brand string) *Parser {
	return &Parser{
		Brand: brand,
	}
}

func (parser *Parser) getPagesCount(doc *goquery.Document) (int, error) {
	pages, exist := doc.Find("li.hide_in_mobile").Last().Find("a").Attr("href")
	if !exist {
		return 0, errors.New("not found pages")
	}
	return strconv.Atoi(strings.Trim(strings.TrimSpace(pages), `?page=`))
}

func (parser *Parser) getItemsUrl(doc *goquery.Document) ([]string, error) {
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

func (parser *Parser) getPageDocument(brand string, page int) (*goquery.Document, error) {
	url := fmt.Sprintf("https://www.fishersci.com/us/en/brands/%s.html?page=%d", brand, page)
	return parser.getHtmlDocument(url)
}

func (parser *Parser) getItemDocument(item string) (*goquery.Document, error) {
	url := fmt.Sprintf("https://www.fishersci.com%s", item)
	return parser.getHtmlDocument(url)
}

func (parser *Parser) getHtmlDocument(url string) (*goquery.Document, error) {
	/*
		proxy, err := neturl.Parse("http://user:pass@ip:port")
		if err != nil {
			return nil, err
		}
		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
			},
		}
	*/
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error status code: %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func getCurrentTimeUTC() time.Time {
	return time.Now().UTC()
}

func (parser *Parser) getItemPriceFromAPI(itemArtc string) (float64, error) {
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
	priceStr, err := jsonparser.GetString(body, "priceAndAvailability", itemArtc, "[0]", "price")
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
func (parser *Parser) getSingleItemData(doc *goquery.Document) (*ItemData, error) {
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

	price, err := parser.getItemPriceFromAPI(artc)
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
	created := getCurrentTimeUTC()
	data := &ItemData{
		Brand:   parser.Brand,
		Article: artc,
		Info: Info{
			Label:        label,
			Description:  descript,
			Manufacturer: manufact,
			Price:        price,
			Available:    available,
			CreatedAt:    created,
		},
	}
	return data, nil
}

func (parser *Parser) getItemData(doc *goquery.Document) (*ItemData, []string, error) {
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
		data, err := parser.getSingleItemData(doc)
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

	created := getCurrentTimeUTC()
	data := &ItemData{
		Brand:   parser.Brand,
		Article: artc,
		Info: Info{
			Label:        label,
			Description:  descript,
			Manufacturer: manufact,
			Price:        price,
			Available:    available,
			CreatedAt:    created,
		},
	}
	return data, nil, nil
}

func (parser *Parser) FisherSciencific() {
	currentPageDoc, err := parser.getPageDocument(parser.Brand, 0)
	if err != nil {
		log.Fatal(err)
	}
	pageCount, err := parser.getPagesCount(currentPageDoc)
	if err != nil {
		log.Fatal(err)
	}
	chanPagesDoc := make(chan *goquery.Document, pageCount)
	chanErr := make(chan error)
	defer close(chanPagesDoc)
	go func() {
		for i := 1; i <= pageCount; i++ {
			go func(num int) {
				currentPageDoc, err := parser.getPageDocument(parser.Brand, num)
				if err != nil {
					chanErr <- err
					return
				}
				chanPagesDoc <- currentPageDoc
			}(i)
		}
	}()
	chanItemsUrl := make(chan []string, pageCount)
	defer close(chanItemsUrl)
	go func() {
		for pagesDoc := range chanPagesDoc {
			go func(doc *goquery.Document) {
				itemsUrl, err := parser.getItemsUrl(doc)
				if err != nil {
					chanErr <- err
					return
				}
				chanItemsUrl <- itemsUrl
			}(pagesDoc)
		}
	}()
	chanItemsDoc := make(chan *goquery.Document, 30)
	go func() {
		for itemsUrl := range chanItemsUrl {
			go func(urls []string) {
				for _, currentItemUrl := range urls {
					currentItemDoc, err := parser.getItemDocument(currentItemUrl)
					if err != nil {
						chanErr <- err
						return
					}
					chanItemsDoc <- currentItemDoc
				}
			}(itemsUrl)
		}
	}()
	chanItemsData := make(chan *ItemData, 30)
	chanInternalUrls := make(chan []string, 30)
	go func() {
		for itemDoc := range chanItemsDoc {
			go func(doc *goquery.Document) {
				data, multipleItemsUrl, err := parser.getItemData(doc)
				if err != nil {
					chanErr <- err
					return
				}
				if len(multipleItemsUrl) != 0 {
					chanInternalUrls <- multipleItemsUrl
				} else {
					chanItemsData <- data
				}
			}(itemDoc)
		}
	}()
	chanInternalDocs := make(chan *goquery.Document, 30)
	go func() {
		for internalUrls := range chanInternalUrls {
			go func(urls []string) {
				for _, internalItemUrl := range urls {
					internalItemDoc, err := parser.getItemDocument(internalItemUrl)
					if err != nil {
						chanErr <- err
						return
					}
					chanInternalDocs <- internalItemDoc
				}
			}(internalUrls)
		}
	}()
	go func() {
		for internalDoc := range chanInternalDocs {
			go func(doc *goquery.Document) {
				data, _, err := parser.getItemData(doc)
				if err != nil {
					chanErr <- err
					return
				}
				chanItemsData <- data
			}(internalDoc)
		}
	}()
	defer close(chanErr)
	for itemData := range chanItemsData {
		PrettyPrint(itemData)
	}
}
