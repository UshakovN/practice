package parser

/*
import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const (
	itemExpr = `/shop/.+[^=]$`
	pageExpr = `\?page=\d+`
)

var (
	count  int    = 0
	link   string = ""
	buffer        = make([]string, 0)
)

func getPagesCount(node *html.Node) {
	reg, err := regexp.Compile(pageExpr)
	if err != nil {
		log.Fatalf("invalid regex: %s", err)
	}
	getLink(node, pageExpr)
	if reg.MatchString(link) {
		buffer, err := strconv.Atoi(strings.ReplaceAll(link, `?page=`, ""))
		if err != nil {
			log.Fatalf("conversion error: %s", err)
		}
		if buffer > count {
			count = buffer
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		getPagesCount(c)
	}
}

func getLink(child *html.Node, expr string) {
	reg, err := regexp.Compile(expr)
	if err != nil {
		log.Fatalf("invalid regex: %s", err)
	}
	if child.Type == html.ElementNode &&
		child.Data == "a" {
		for _, childAttr := range child.Attr {
			if childAttr.Key == "href" && reg.MatchString(childAttr.Val) {
				link = childAttr.Val
			}
		}
	}
}

func findItemLink(child *html.Node) {
	getLink(child, itemExpr)
	for c := child.FirstChild; c != nil; c = c.NextSibling {
		findItemLink(c)
	}
}

func getItemLinks(node *html.Node) {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" &&
				attr.Val == "row search_result_item displayToggleControlled " {
				findItemLink(node)
				if link != "" {
					buffer = append(buffer, link)
				}
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		getItemLinks(c)
	}
}

func getHtmlPage(brandTag string, pageNumber int) (*html.Node, error) {
	url := fmt.Sprintf("https://www.fishersci.com/us/en/brands/%s.html?page=%d", brandTag, pageNumber)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("uncorrect ulr: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error status code: %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func FisherSciencific(brandTag string) {
	doc, err := getHtmlPage(brandTag, 0)
	if err != nil {
		log.Fatal(err)
	}
	getPagesCount(doc)
	fmt.Println(count)
	for i := 1; i <= count; i++ {
		doc, err = getHtmlPage(brandTag, i)
		if err != nil {
			log.Fatal(err)
		}
		getItemLinks(doc)
	}
	for _, l := range buffer {
		fmt.Println(l)
	}
	fmt.Println(len(buffer))
}
*/
