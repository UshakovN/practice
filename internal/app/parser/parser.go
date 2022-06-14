package parser

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

func FisherSciencific(brandTag string) {
	URL := fmt.Sprintf("https://www.fishersci.com/us/en/brands/%s.html", brandTag)
	resp, err := http.Get(URL)
	if err != nil {
		log.Fatalf("uncorrect ulr: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error status code: %d %s", resp.StatusCode, resp.Status)
	}
	_, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("unexpected error %s", err)
	}
	file, _ := os.Create("buffer.html")
	defer file.Close()
	io.Copy(file, resp.Body)
}
