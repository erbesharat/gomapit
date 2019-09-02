package main

import (
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/erbesharat/gomapit/fileio"
	httphandler "github.com/erbesharat/gomapit/handler"
	"github.com/erbesharat/gomapit/sitemap"
)

func main() {
	// Setup flags
	var parallel, depth int
	var output string
	flag.IntVar(&parallel, "parallel", 1, "number of parallel workers to navigate through site")
	flag.IntVar(&depth, "depth", 1, "max depth of url navigation recursion")
	flag.StringVar(&output, "output", "./sitemap.xml", "output file path")
	flag.Parse()

	// Check arguments length
	if len(os.Args) < 2 {
		log.Fatalf("[gomapit]: Please pass an url as an argument")
	}

	// Parse the start URL
	url, err := url.Parse(os.Args[len(os.Args)-1])
	if err != nil {
		log.Fatalf("[gomapit]: Can't parse the given url: %s", err.Error())
	}

	// Send a request to the start URL
	resp, err := http.Get(url.String())
	if err != nil {
		log.Fatalf("[gomapit]: Couldn't send the first request: %s", err.Error())
	}

	// Parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("[gomapit]: Couldn't parse the response body: %s", err.Error())
	}

	// Fetch the links from the response body
	links, err := httphandler.GetLinks(string(body), url.Host)
	if err != nil {
		log.Fatalf("[gomapit]: Couldn't get the links from response: %s", err.Error())
	}

	smap := sitemap.NewMap()

	if depth > 1 {
		handler := map[int][]string{
			0: links,
		}
		for i := 1; i <= depth; i++ {
			handler[i] = httphandler.FetchNestedURLs(handler[i-1], url.Host)
		}
		for _, urls := range handler {
			links = append(links, urls...)
		}
		links = httphandler.Deduplicate(links)
	}

	// Add the URLs to sitemap instance
	for _, link := range links {
		smap.AddURL(&sitemap.URL{Loc: link})
	}

	// Generate the final xml response
	xml, err := xml.Marshal(smap)
	if err != nil {
		log.Fatalf("[gomapit]: Could'nt generate the xml from sitemap struct: %s", err.Error())
	}

	if err := fileio.WriteXML(xml, output); err != nil {
		log.Fatalf("[gomapit]: Couldn't write to the file: %s", err.Error())
	}
}
