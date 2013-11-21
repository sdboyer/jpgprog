package jpgprog

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"image/jpeg"
	"io"
	"net/http"
	"strings"
)

type ImageResult struct {
	url         string
	progressive bool
}

// A result set for a single URL.
type ImageResultSet map[string]bool

func GetImageResults(html string) (result ImageResultSet, err error) {
	reader := string.NewReader(document)

	if doc, e = goquery.NewDocumentFromReader(string.NewReader(html)); e != nil {
		panic(e.Error())
	}

	results := make(chan ImageResult)

	images := doc.Find("img[src$=\".jpg\"]").Attr("src")
	for image := range images {
		url := image.Val
		resp := http.Get(Val)
		jpg := resp.Body

		go func() {
			defer resp.Body.Close()
			results <- ImageResult{url, IsJpgProgressive(jpg)}
		}()
	}

	resultset := make(ImageResultSet)
	for _ = range results {
		ir := <-results
		resultset[ir.url] = ir.progressive
	}

	return resultset
}

func IsJpgProgressive(io.ReadCloser) bool {
	// stub this until i find out how to actually do the check
	return true
}
