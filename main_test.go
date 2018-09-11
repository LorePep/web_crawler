package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
)

var templates *template.Template

func init() {
	templates = template.Must(template.ParseGlob("templates/*"))
}

func TestGetLinksFromURL(t *testing.T) {
	ts := startMockServerOrFail(t)
	defer ts.Close()

	testcases := []struct {
		name     string
		url      string
		expected []string
	}{
		{
			name:     "no_links",
			url:      ts.URL + "/nolinks",
			expected: []string{},
		},
		{
			name:     "one_link",
			url:      ts.URL,
			expected: []string{"/twolinks"},
		},
		{
			name:     "two_links",
			url:      ts.URL + "/twolinks",
			expected: []string{"/", "/nolinks"},
		},
		// {
		// 	// TODO how to treat redirects??
		// 	name:     "temporary_redirect",
		// 	url:      ts.URL + "/redirect",
		// 	expected: []string{"/"},
		// },
		// {
		// 	// TODO how to treat redirects??
		// 	name:     "moved_permanently",
		// 	url:      ts.URL + "/movedpermanently",
		// 	expected: []string{"/"},
		// },
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := getLinksFromURL(tt.url)
			if err != nil {
				fmt.Printf("expected nil error, got: %s\n", err)
				t.Fail()
			}
			if !areLinksEqual(tt.expected, actual) {
				fmt.Printf("expected %v, got: %v\n", tt.expected, actual)
				t.Fail()
			}
		})
	}
}

func TestIsValidContentType(t *testing.T) {
	resp, _ := http.Get("www.monzo.com")
	fmt.Println(isValidContentType(resp))
}

func areLinksEqual(ls1, ls2 []string) bool {
	if len(ls1) != len(ls2) {
		return false
	}
	sort.Strings(ls1)
	sort.Strings(ls2)

	for i := range ls1 {
		if ls1[i] != ls2[i] {
			return false
		}
	}

	return true
}

func startMockServerOrFail(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.Handle("/favicon.ico", http.NotFoundHandler())
	mux.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			templates.ExecuteTemplate(w, "index.gohtml", nil)
		})
	mux.HandleFunc("/twolinks",
		func(w http.ResponseWriter, req *http.Request) {
			templates.ExecuteTemplate(w, "twolinks.gohtml", nil)
		})
	mux.HandleFunc("/outofdomain",
		func(w http.ResponseWriter, req *http.Request) {
			templates.ExecuteTemplate(w, "outofdomain.gohtml", nil)
		})
	mux.HandleFunc("/redirect",
		func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req, "/", http.StatusSeeOther)
		})
	mux.HandleFunc("/relative",
		func(w http.ResponseWriter, req *http.Request) {
			templates.ExecuteTemplate(w, "relativelinks.gohtml", nil)
		})
	mux.HandleFunc("/movedpermanently",
		func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req, "/nolinks", http.StatusMovedPermanently)
		})
	mux.HandleFunc("/nolinks",
		func(w http.ResponseWriter, req *http.Request) {
			templates.ExecuteTemplate(w, "nolinks.gohtml", nil)
		})
	mux.HandleFunc("/forbidden",
		func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "thou shall not pass", http.StatusForbidden)
		})

	return httptest.NewServer(mux)
}
