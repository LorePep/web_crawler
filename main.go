package main

import (
	"fmt"
	"io"
	"net/http"
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
				tokens <- struct{}{}

				go func(link string) {
					foundLinks, _ := getLinksFromURL(link)
					if foundLinks != nil {
						urlsToCrawl <- foundLinks
					}
					<-tokens
				}(link)
			}
		}
	}

	fmt.Println(siteMap)
}

func isLinkValid(l string) bool {
	if len(l) > 0 {
		if l[0] == '/' {
			return true
		}
	}

	return false
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
						if isLinkValid(attr.Val) {
							links = append(links, attr.Val)
						}
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
