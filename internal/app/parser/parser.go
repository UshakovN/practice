package parser

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"golang.org/x/net/html"
)

func findLink(child *html.Node, expr string) {
	reg, err := regexp.Compile(expr)
	if err != nil {
		log.Fatalf("invalid regex: %s", err)
	}
	if child.Type == html.ElementNode &&
		child.Data == "a" {
		for _, childAttr := range child.Attr {
			if childAttr.Key == "href" && reg.MatchString(childAttr.Val) {
				fmt.Println(childAttr.Val)
			}
		}
	}
}

func findCurrentLink(child *html.Node) {
	findLink(child, `/shop/.+[^=]$`)
	for c := child.FirstChild; c != nil; c = c.NextSibling {
		findCurrentLink(c)
	}
}

func getItemLinks(node *html.Node) {
	findLink(node, `\?page=\d+`)
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" &&
				attr.Val == "row search_result_item displayToggleControlled " {
				findCurrentLink(node)
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		getItemLinks(c)
	}
}

func FisherSciencific(brandTag string) {
	url := fmt.Sprintf("https://www.fishersci.com/us/en/brands/%s.html", brandTag)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("uncorrect ulr: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error status code: %d %s", resp.StatusCode, resp.Status)
	}
	/*
		_, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatalf("unexpected error %s", err)
		}
		file, err := os.Create("buffer.html")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		io.Copy(file, resp.Body)
	*/
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error %s", err)
	}
	getItemLinks(doc)
}
