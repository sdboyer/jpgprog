package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/sdboyer/jpgprog/lib"
	"net/http"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "jpgprog"
	app.Usage = "Specify a url, and jpegprog will report on whether any jpegs contained therein are progressive."
	app.Action = func(c *cli.Context) {
		for _, url := range c.Args() {
			doFromUrl(url)
		}
	}

	app.Run(os.Args)
}

func doFromUrl(url string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	results, err := jpgprog.GetImageResults(resp.Body)
	if err != nil {
		panic(err)
	}

	for url, progressive := range results {
		fmt.Println(progressive, url)
	}
}
