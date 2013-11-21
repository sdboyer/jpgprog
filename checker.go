package jpgprog

import (
	"github.com/PuerkitoBio/goquery"
	"image/jpeg"
	"io"
	"net/http"
	"strings"
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

// Given a string of HTML, searches it for jpg references and determines
// whether referenced images are progressive or not for all the images.
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

// Given an io reader (which we assume contains jpeg data), determine if the
// jpeg represented therein is progressive.
func IsJpgProgressive(r io.Reader) (bool, error){
  var cur [512]byte

  // process initial SOI marker
  _, err := io.ReadFull(r, cur[0:2])
  if err != nil {
    return nil, err
  }
  if marker[0] != 0xff || marker[1] != 0xd8 {
    return nil, jpeg.FormatError("missing SOI marker")
  }

  // now, scan through the body for a progressive marker
  for {
    _, err := io.ReadFull(r, cur[0:2])
    if err != nil {
      return nil, err
    }

    for cur[0] != 0xff {
      cur[0] = cur[1]
      cur[1], err = r.ReadByte()
      if err != nil {
        return nil, err
      }
    }
    marker := cur[1]

    if marker == 0x00 {
      // "\xff\x00" is extraneous data, according to JPEG spec
      continue;
    }

    for marker == 0xff {
      // spec allows as many fill bytes as desired with val "\xff"
      marker, err = r.ReadByte()
      if err != nil {
        return nil, err
      }
    }

    if 0xd0 <= marker && marker <= 0xd7 {
      // weird shit i don't grok about reset markers
      continue;
    }

    // we know we're on a marker that has a length payload - figure
    // out how long it is (less the 2 bytes for the marker)
    _, err = io.ReadFull(r, cur[0:2])
    if err != nil {
      return nil, err
    }

    n := int(cur[0])<<8 + int(cur[1]) - 2

    if marker == sof0Marker || marker == sof2Marker {
      return marker == sof2marker, nil
    }
  }
}
