package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

const defaultMaxConcurrency = 4

var siteMap = make(map[string]struct{})

var defaultHTMLContentType = "text/html"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	rootURL := os.Args[1]

	urlsToCrawl := make(chan []string)
	tokens := make(chan struct{}, defaultMaxConcurrency)

	go func() { urlsToCrawl <- []string{rootURL} }()
	toCrawlCount := 1

	for ; toCrawlCount > 0; toCrawlCount-- {
		list := <-urlsToCrawl
		for _, link := range list {
			if _, ok := siteMap[link]; !ok {
				siteMap[link] = struct{}{}
				toCrawlCount++

				go func(link string) {
					tokens <- struct{}{}
					foundLinks, _ := getLinksFromURL(link)
					sanitized := sanitizeLinks(foundLinks, link)
					if len(sanitized) > 0 {
						urlsToCrawl <- sanitized
					}
					<-tokens
				}(link)
			}
		}
	}

	fmt.Println(siteMap)
}

func sanitizeLinks(links []string, base string) []string {
	sanitized := make([]string, 0)

	for _, link := range links {
		url, err := url.Parse(link)
		if err != nil {
			continue
		}
		baseURL, err := url.Parse(base)
		if err != nil {
			continue
		}
		url = baseURL.ResolveReference(url)
		sanitized = append(sanitized, normalizeURL(url.String()))
	}

	return sanitized
}

func normalizeURL(url string) string {
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}

	return strings.ToLower(url)

}

func getLinksFromURL(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !isValidContentType(&resp.Header) {
		return nil, nil
	}

	links := getLinksFromBody(resp.Body)

	return links, nil
}

func getLinksFromBody(body io.ReadCloser) []string {
	links := make([]string, 0)

	tokenizer := html.NewTokenizer(body)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			return links
		}
		if tt == html.StartTagToken || tt == html.EndTagToken {
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			}
		}
	}
}

func isValidContentType(header *http.Header) bool {
	contentType := header.Get("Content-type")
	parsed := parseContentType(contentType)
	return parsed == defaultHTMLContentType
}

func parseContentType(ct string) string {
	split := strings.Split(ct, ";")
	return split[0]
}

func printUsage() {
	fmt.Print("Usage:\n webCrawler startingSite\n")
}
