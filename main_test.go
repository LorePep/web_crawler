package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

var fixtureTemplate = `<a href="{{.}}">Something</a>`

func TestGetLinksFromURL(t *testing.T) {
	ts := startMockServerOrFail(t)
	defer ts.Close()

	links, err := getLinksFromURL(ts.URL)
	if err != nil {
		t.Fail()
	}
	if len(links) != 1 {
		fmt.Printf("expected 1 link, got %d", len(links))
		t.Fail()
	}
	if links[0] != "oneLink.com" {
		fmt.Printf("expected oneLink.com, got %s", links[0])
		t.Fail()
	}
}

func startMockServerOrFail(t *testing.T) *httptest.Server {
	tmpl, err := template.New("page").Parse(fixtureTemplate)
	if err != nil {
		t.Fatal()
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, "oneLink.com")
	}))

	return ts
}
