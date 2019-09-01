// Package handler is the package responsible for sending http requsts
// and extracting the links from response bodies.
package handler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var linkTagRegex = regexp.MustCompile("<a[^>]+href=\"(.*?)\"[^>]*>")

// GetLinks extracts the URLs from the given body and finds the right URLs
// using the given host. It returns a slice of URLs and if something goes wrong,
// slice would be empty and an error would be returend
func GetLinks(body string, host string) ([]string, error) {
	var links []string
	result := linkTagRegex.FindAllStringSubmatch(body, -1)

	for _, tag := range result {
		url, err := ValidateURL(tag[1], host)
		if err != nil {
			return nil, err
		}
		if url != nil {
			links = append(links, url.String())
		}
	}

	links = Deduplicate(links)

	return links, nil
}

// ValidateURL validates the given URL using the given host and returns the correct URL.
// If something goes wrong, it would return an empty URL and an error.
func ValidateURL(uri, host string) (*url.URL, error) {
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

// FetchNestedURLs finds the URLs inside the nested pages and returns a slice of URLs.
func FetchNestedURLs(links []string, host string) []string {
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
			urls, err := GetLinks(string(body), host)
			if err != nil {
				log.Fatalf("[gomapit]: Couldn't get the links from response: %s", err.Error())
			}
			result = append(result, urls...)
		}
	}
	return Deduplicate(result)
}

// Deduplicate checks the given slice of URLs for a duplicate URLs
// and removes them from the returend result.
func Deduplicate(input []string) []string {
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
