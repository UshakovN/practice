package parser

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/UshakovN/practice/internal/app/common"
	"github.com/UshakovN/practice/internal/app/store"
	json "github.com/buger/jsonparser"
	"golang.org/x/sync/semaphore"
)

const (
	RESOURCE_URL       = "https://www.fishersci.com"
	MAX_GOROUTINES_PGS = 5
)

type Parser struct {
	Brand Brand
}

type Brand struct {
	Name string
	Code string
}

func NewParser(brand Brand) *Parser {
	return &Parser{
		Brand: Brand{
			Name: brand.Name,
			Code: brand.Code,
		},
	}
}

func (parser *Parser) getPagesCount(doc *goquery.Document) (int, error) {

	pagination := doc.Find("div.node-pagination")
	if pagination.Nodes == nil { // onepages / tables
		return 1, nil
	}

	pages, exist := pagination.Find("li").Last().Prev().Find("a").Attr("href")
	// doc.Find("li.hide_in_mobile").Last().Find("a").Attr("href")
	if !exist {
		return 0, errors.New("not found pages")
	}
	return strconv.Atoi(strings.Trim(strings.TrimSpace(pages), `?page=`))
}

func (parser *Parser) getItemsUrl(doc *goquery.Document) ([]string, error) {
	buffer := make([]string, 0)

	// brands with tables (brand navigator form)
	brandNavigator := doc.Find("nav.brand-navigation")
	itemsTableUrl := ""
	if brandNavigator.Nodes != nil {
		brandNavigator.Find("ul>li").EachWithBreak(
			func(index int, item *goquery.Selection) bool {
				aContent := item.Find("a")
				if aContent.Contents().Text() == "Products" {
					itemsTableUrl, _ = aContent.Attr("href")
					return false
				}
				return true
			})

		if itemsTableUrl != "" {
			itemsTable, err := parser.getItemDocument(itemsTableUrl)
			if err != nil {
				return nil, err
			}
			itemsTable.Find("div.tabs_wrap").Find("div.tab").Each(
				func(index int, item *goquery.Selection) {
					item.Find("table.general_table.product_table").Find("tbody").Each(
						func(i int, s *goquery.Selection) {
							itemUrl, exist := s.Find("div>a").Attr("href")
							if exist {
								buffer = append(buffer, strings.TrimSpace(itemUrl))
							}
						})
				})
		}
	}

	doc.Find("div.search_results>div.search_results_listing>div.row.search_result_item").Each(
		func(index int, item *goquery.Selection) {
			url, exist := item.Find("div.columns>div.block>a").Attr("href")
			if exist {
				buffer = append(buffer, strings.TrimSpace(url))
			}
		})

	if len(buffer) == 0 {
		return nil, errors.New("not found item urls") // DEBUG
	}
	return buffer, nil
}

func (parser *Parser) getPageDocument(brand Brand, page int) (*goquery.Document, error) {
	url := fmt.Sprintf("https://www.fishersci.com/us/en/brands/%s/%s.html?page=%d",
		brand.Code, brand.Name, page)
	return parser.GetHtmlDocument(url)
}

func (parser *Parser) getItemDocument(item string) (*goquery.Document, error) {
	url := fmt.Sprintf("%s%s", RESOURCE_URL, item)
	return parser.GetHtmlDocument(url)
}

func (parser *Parser) GetHtmlDocument(url string) (*goquery.Document, error) {
	/*
		zyte := proxy.NewZyteProxy(os.Getenv("PROXY_CERT_PATH"))
		client := &http.Client{
			Transport: zyte.GetHttpTransport(),
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

func getCurrentTimeUTC() string {
	return time.Now().UTC().String()
}

func (parser *Parser) getItemPriceFromAPI(itemArtc string) (float64, error) {
	url := fmt.Sprintf("%s/shop/products/service/pricing", RESOURCE_URL)
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
func (parser *Parser) getSingleItemData(doc *goquery.Document) (*store.ItemData, error) {
	available := true
	selMain := doc.Find("div.singleProductPage")

	reg := regexp.MustCompile(`[[:^ascii:]]`)

	label := strings.ReplaceAll(
		reg.ReplaceAllLiteralString(
			strings.TrimSpace(
				selMain.Find("div.productSelectors>h1").Contents().First().Text()), " "), "  ", " ")
	if label == "" {
		return nil, errors.New("label not found")
	}
	descript := "None" // secondary items without a description

	// artc, exist := selMain.Find("div.glyphs_html_container").Attr("data-partnumber")
	artc, exist := selMain.Find("input").Last().Attr("data-partnumbers")
	if !exist {
		return nil, errors.New("article not found")
	}
	artc = strings.Split(artc, ",")[0] // temporarily

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
		manufact = "None" // temporarily
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
	data := &store.ItemData{
		Brand:        strings.Title(parser.Brand.Name),
		Article:      artc,
		Label:        label,
		Description:  descript,
		Manufacturer: manufact,
		Price:        price,
		Available:    available,
		CreatedAt:    created,
	}
	return data, nil
}

func (parser *Parser) getItemData(doc *goquery.Document) (*store.ItemData, []string, error) {
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
			return nil, nil, errors.New("not found internal urls")
		}
		return nil, buffer, nil
	}

	// single (secondary) item
	singlepage := doc.Find("div.productSelectors")
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
		descript = "None" // some items without a description
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
	data := &store.ItemData{
		Brand:        strings.Title(parser.Brand.Name),
		Article:      artc,
		Label:        label,
		Description:  descript,
		Manufacturer: manufact,
		Price:        price,
		Available:    available,
		CreatedAt:    created,
	}
	return data, nil, nil
}

func (parser *Parser) FisherSciencific(client *store.Client) {
	MAX_GOROUTINES_ITMS := MAX_GOROUTINES_PGS

	currentPageDoc, err := parser.getPageDocument(parser.Brand, 0)
	if err != nil {
		log.Fatal(err)
	}
	pageCount, err := parser.getPagesCount(currentPageDoc)
	if err != nil {
		log.Fatal(err)
	}
	// sync running goroutines count
	if pageCount < MAX_GOROUTINES_PGS {
		MAX_GOROUTINES_ITMS = pageCount
	}

	// error logging
	chanError := make(chan error, 30)

	chanPagesDoc := make(chan *goquery.Document, pageCount)

	semPages := semaphore.NewWeighted(int64(MAX_GOROUTINES_PGS))
	// semPages := make(chan struct{}, MAX_GOROUTINES_PGS) // fixed running goroutines
	var wgPagesDoc sync.WaitGroup
	go func(wg *sync.WaitGroup) {
		// pageCount
		for i := 1; i <= pageCount; i++ {
			// semPages <- struct{}{} // add
			if err := semPages.Acquire(context.TODO(), 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)
			}
			wg.Add(1)
			go func(num int, wg *sync.WaitGroup) {
				currentPageDoc, err := parser.getPageDocument(parser.Brand, num)
				if err != nil {
					chanError <- err
				}
				chanPagesDoc <- currentPageDoc
				// <-semPages // done
				semPages.Release(1)
				wg.Done()
			}(i, wg)
		}
		go func(wg *sync.WaitGroup) {
			wg.Wait()
			close(chanPagesDoc)
		}(wg)
	}(&wgPagesDoc)

	chanItemsUrl := make(chan []string, pageCount)

	var wgItemsUrl sync.WaitGroup
	go func(wg *sync.WaitGroup) {
		for pagesDoc := range chanPagesDoc {
			wg.Add(1)
			go func(doc *goquery.Document, wg *sync.WaitGroup) {
				itemsUrl, err := parser.getItemsUrl(doc)
				if err != nil {
					chanError <- err
				}
				chanItemsUrl <- itemsUrl
				wg.Done()
			}(pagesDoc, wg)
		}
		go func(wg *sync.WaitGroup) {
			wg.Wait()
			close(chanItemsUrl)
		}(wg)
	}(&wgItemsUrl)
	chanItemsDoc := make(chan *goquery.Document, 30)

	semItems := semaphore.NewWeighted(int64(MAX_GOROUTINES_ITMS))
	// semItems := make(chan struct{}, MAX_GOROUTINES_PGS) // fixed running goroutines
	var wgItemsDoc sync.WaitGroup
	go func(wg *sync.WaitGroup) {
		for itemsUrl := range chanItemsUrl {
			if err := semItems.Acquire(context.TODO(), 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)
			}
			// semItems <- struct{}{} // add
			wg.Add(1)
			go func(urls []string, wg *sync.WaitGroup) {
				for _, currentItemUrl := range urls {
					currentItemDoc, err := parser.getItemDocument(currentItemUrl)
					if err != nil {
						chanError <- err
					}
					chanItemsDoc <- currentItemDoc
				}
				// <-semItems // done
				semItems.Release(1)
				wg.Done()
			}(itemsUrl, wg)
		}
		go func(wg *sync.WaitGroup) {
			wg.Wait()
			close(chanItemsDoc)
		}(wg)
	}(&wgItemsDoc)

	chanItemsData := make(chan *store.ItemData, 30)
	chanInternalUrls := make(chan []string, 30)

	var wgItemsData, wgInternalUrls sync.WaitGroup
	go func(wgItem, wgInternal *sync.WaitGroup) {
		for itemDoc := range chanItemsDoc {
			wgItem.Add(1)
			wgInternal.Add(1)
			go func(doc *goquery.Document, wgItem *sync.WaitGroup, wgInternal *sync.WaitGroup) {
				data, multipleItemsUrl, err := parser.getItemData(doc)
				if err != nil {
					chanError <- err
				}
				if len(multipleItemsUrl) != 0 {
					chanInternalUrls <- multipleItemsUrl
				} else {
					chanItemsData <- data
				}
				wgItem.Done()
				wgInternal.Done()
			}(itemDoc, wgItem, wgInternal)
		}
		go func(wgInternal *sync.WaitGroup) {
			wgInternal.Wait()
			close(chanInternalUrls)
		}(wgInternal)
	}(&wgItemsData, &wgInternalUrls)

	chanInternalDocs := make(chan *goquery.Document, 30)

	var wgInternalDocs sync.WaitGroup
	go func(wg *sync.WaitGroup) {
		for internalUrls := range chanInternalUrls {
			wg.Add(1)
			go func(urls []string, wg *sync.WaitGroup) {
				for _, internalItemUrl := range urls {
					internalItemDoc, err := parser.getItemDocument(internalItemUrl)
					if err != nil {
						chanError <- err
					}
					chanInternalDocs <- internalItemDoc
				}
				wg.Done()
			}(internalUrls, wg)
		}
		go func(wg *sync.WaitGroup) {
			wg.Wait()
			close(chanInternalDocs)
		}(wg)
	}(&wgInternalDocs)

	var wgItem sync.WaitGroup
	go func(wg *sync.WaitGroup) {
		for internalDoc := range chanInternalDocs {
			wg.Add(1)
			go func(doc *goquery.Document) {
				data, _, err := parser.getItemData(doc)
				if err != nil {
					chanError <- err
				}
				chanItemsData <- data
				wg.Done()
			}(internalDoc)
		}
		go func(wg *sync.WaitGroup) {
			wg.Wait()
			close(chanItemsData)
		}(wg)
	}(&wgItem)

	// form items batch
	itemsBatch := make([]*store.ItemData, 0)
	batchSize := 0

loop:
	for {
		select {
		case itemData, ok := <-chanItemsData:
			if !ok {
				close(chanError)
				break loop
			}
			// common.PrettyPrint(itemData)

			if !common.BatchContains(itemsBatch, itemData) {
				batchSize++
				itemsBatch = append(itemsBatch, itemData)
			}
			if batchSize == 25 {
				common.PrettyPrint(itemsBatch)
				if err := client.WriteBatch(itemsBatch); err != nil {
					chanError <- err
				}
				log.Println("batch sent")
				itemsBatch = itemsBatch[:0]
				batchSize = 0
			}

		case err := <-chanError:
			log.Println(err)
		}
	}
	if batchSize > 0 {
		common.PrettyPrint(itemsBatch)
	}
	log.Println("parse complete")
	/*
		for itemData := range chanItemsData {
			if !common.BatchContains(itemsBatch, itemData) {
				batchSize++
				itemsBatch = append(itemsBatch, itemData)
			}
			if batchSize == 25 {
				common.PrettyPrint(itemsBatch)
				if err := client.WriteBatch(itemsBatch); err != nil {
					chanError <- err
				}
				itemsBatch = itemsBatch[:0]
				batchSize = 0
			}
		}
		if batchSize > 0 {
			common.PrettyPrint(itemsBatch)
		}
		log.Println("parse complete")
	*/
}
