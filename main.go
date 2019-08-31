package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// ChangeFreq specifies change frequency of a sitemap entry. It is just a string.
type ChangeFreq string

// Feel free to use these constants for ChangeFreq (or you can just supply
// a string directly).
const (
	Always  ChangeFreq = "always"
	Hourly  ChangeFreq = "hourly"
	Daily   ChangeFreq = "daily"
	Weekly  ChangeFreq = "weekly"
	Monthly ChangeFreq = "monthly"
	Yearly  ChangeFreq = "yearly"
	Never   ChangeFreq = "never"
)

// URL entry in sitemap or sitemap index. LastMod is a pointer
// to time.Time because omitempty does not work otherwise. Loc is the
// only mandatory item. ChangeFreq and Priority must be left empty when
// using with a sitemap index.
type URL struct {
	Loc        string     `xml:"loc"`
	LastMod    *time.Time `xml:"lastmod,omitempty"`
	ChangeFreq ChangeFreq `xml:"changefreq,omitempty"`
	Priority   float32    `xml:"priority,omitempty"`
}

// Sitemap represents a complete sitemap which can be marshaled to XML.
// New instances must be created with New() in order to set the xmlns
// attribute correctly. Minify can be set to make the output less human
// readable.
type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`

	URLs []*URL `xml:"url"`

	Minify bool `xml:"-"`
}

var linkTagRegex = regexp.MustCompile("<a[^>]+href=\"(.*?)\"[^>]*>")

func main() {
	// Setup flags
	var parallel, depth int
	var output string
	flag.IntVar(&parallel, "parallel", 1, "number of parallel workers to navigate through site")
	flag.IntVar(&depth, "depth", 2, "max depth of url navigation recursion")
	flag.StringVar(&output, "output", "./sitemap.xml", "output file path")
	flag.Parse()

	// Check arguments length
	if len(os.Args) < 2 {
		log.Fatalf("[gomapit]: Please pass an url as an argument")
	}

	// Parse the start URL
	url, err := url.Parse(os.Args[1])
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
	links, err := getLinks(string(body), url.Host)
	if err != nil {
		log.Fatalf("[gomapit]: Couldn't get the links from response: %s", err.Error())
	}

	sitemap := NewMap()

	if depth > 1 {
		handler := map[int][]string{
			0: links,
		}
		for i := 1; i <= depth; i++ {
			handler[i] = fetchNestedURLs(handler[i-1], url.Host)
		}
		for _, urls := range handler {
			links = append(links, urls...)
		}
		links = deduplicate(links)
	}

	// Add the URLs to sitemap instance
	for _, link := range links {
		sitemap.AddURL(&URL{Loc: link})
	}

	// Generate the final xml response
	xml, err := xml.Marshal(sitemap)
	if err != nil {
		log.Fatalf("[gomapit]: Could'nt generate the xml from sitemap struct: %s", err.Error())
	}
	fmt.Println(string(xml))
}

func getLinks(body string, host string) ([]string, error) {
	var links []string
	result := linkTagRegex.FindAllStringSubmatch(body, -1)

	for _, tag := range result {
		url, err := validateURL(tag[1], host)
		if err != nil {
			return nil, err
		}
		if url != nil {
			links = append(links, url.String())
		}
	}

	links = deduplicate(links)

	return links, nil
}

func validateURL(uri, host string) (*url.URL, error) {
	var finalURI *url.URL
	var err error
	if strings.Contains(uri, host) {
		if finalURI, err = url.Parse(uri); err != nil {
			return nil, fmt.Errorf("Couldn't parse the url inside href argument: %s", err.Error())
		}
		return finalURI, nil
	}
	if string(uri[0]) == "/" {
		if finalURI, err = url.Parse(fmt.Sprintf("https://%s%s", host, uri)); err != nil {
			return nil, fmt.Errorf("Couldn't parse the url inside href argument: %s", err.Error())
		}
		return finalURI, nil
	}
	return nil, nil
}

func fetchNestedURLs(links []string, host string) []string {
	var result []string
	for _, link := range links {
		if strings.Contains(link, host) {
			// Send a request to the start URL
			resp, err := http.Get(link)
			if err != nil {
				log.Fatalf("[gomapit]: Couldn't send the first request: %s", err.Error())
			}

			// Parse the response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("[gomapit]: Couldn't parse the response body: %s", err.Error())
			}

			// Fetch the links from the response body
			urls, err := getLinks(string(body), host)
			if err != nil {
				log.Fatalf("[gomapit]: Couldn't get the links from response: %s", err.Error())
			}
			result = append(result, urls...)
		}
	}
	return deduplicate(result)
}

// NewMap returns a fresh instance of Sitemap struct
func NewMap() *Sitemap {
	return &Sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]*URL, 0),
	}
}

// AddURL adds an URL to a Sitemap.
func (s *Sitemap) AddURL(u *URL) {
	s.URLs = append(s.URLs, u)
}

func deduplicate(input []string) []string {
	result := []string{}
	seen := make(map[string]struct{})
	for _, val := range input {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}
