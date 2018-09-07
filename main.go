package main

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/html"
)

const defaultMaxConcurrency = 4

var siteMap = make(map[string]struct{})

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	startingURL := os.Args[1]
	urlsToCrawl := make(chan []string)
	tokens := make(chan struct{}, defaultMaxConcurrency)

	go func() { urlsToCrawl <- []string{startingURL} }()
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
	// Implement retry
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var links []string
	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tt := tokenizer.Next()

		switch tt {
		case html.ErrorToken:
			return links, nil
		case html.StartTagToken, html.EndTagToken:
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

func printUsage() {
	fmt.Print("Usage:\n webCrawler startingSite\n")
}
