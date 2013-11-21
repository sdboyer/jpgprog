package jpgprog

import (
	"bufio"
	"github.com/PuerkitoBio/goquery"
	"image/jpeg"
	"io"
	"net/http"
)

// A struct representing a result for a single image, and its source URL.
type ImageResult struct {
	url         string
	progressive bool
}

// A result set for a single URL.
type ImageResultSet map[string]bool

const (
	sof0Marker = 0xc0
	sof2Marker = 0xc2
)

type Reader interface {
	io.Reader
	ReadByte() (c byte, err error)
}

// Given a string of HTML, searches it for jpg references and determines
// whether referenced images are progressive or not for all the images.
func GetImageResults(body io.ReadCloser) (resultset ImageResultSet, err error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	results := make(chan ImageResult)
	resultset = make(ImageResultSet)

	urls := doc.Find("img[src$=\".jpg\"]").Map(func(i int, s *goquery.Selection) string {
		url, _ := s.Attr("src")
		return url
	})

	for _, url := range urls {
      // YAY GO WEIRDNESS
      url := url
		resp, err := http.Get(url)
		if err != nil {
			return
		}
		jpg := resp.Body

		go func() {
			defer resp.Body.Close()
			prog, err := IsJpgProgressive(jpg)
			if err != nil {
				return
			}
			results <- ImageResult{url, prog}
		}()
	}

	for _ = range results {
		ir := <-results
		resultset[ir.url] = ir.progressive
	}

	return resultset, nil
}

// Given an io reader, determine if the jpeg represented therein is progressive.
func IsJpgProgressive(r io.Reader) (bool, error) {
	var rr Reader
	if rdr, ok := r.(Reader); ok {
		rr = rdr
	} else {
		rr = bufio.NewReader(r)
	}

	var cur [2]byte

	// process initial SOI marker
	_, err := io.ReadFull(rr, cur[0:2])
	if err != nil {
		return false, err
	}
	if cur[0] != 0xff || cur[1] != 0xd8 {
		return false, jpeg.FormatError("missing SOI marker")
	}

	// now, scan through the body for a progressive marker
	for {
		_, err := io.ReadFull(rr, cur[0:2])
		if err != nil {
			return false, err
		}

		for cur[0] != 0xff {
			cur[0] = cur[1]
			cur[1], err = rr.ReadByte()
			if err != nil {
				return false, err
			}
		}
		marker := cur[1]

		if marker == 0x00 {
			// "\xff\x00" is extraneous data, according to JPEG spec
			continue
		}

		if marker == 0xd9 {
			return false, jpeg.UnsupportedError("shouldn't reach EOF.")
		}

		for marker == 0xff {
			// spec allows as many fill bytes as desired with val "\xff"
			marker, err = rr.ReadByte()
			if err != nil {
				return false, err
			}
		}

		if 0xd0 <= marker && marker <= 0xd7 {
			// weird shit i don't grok about reset markers
			continue
		}

		// we've now covered all the types of markers that do not have a
		// two-byte length segment, skip that length segment.
		_, err = io.ReadFull(rr, cur[0:2])
		if err != nil {
			return false, err
		}

		if marker == sof0Marker || marker == sof2Marker {
			return marker == sof2Marker
		}
	}
}
