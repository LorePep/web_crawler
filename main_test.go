package main

import (
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

func TestIsValidContentType(t *testing.T) {
	ts := startMockServerOrFail(t)
	defer ts.Close()

	testcases := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{
			name:        "no_content_type",
			contentType: "",
			expected:    false,
		},
		{
			name:        "empty",
			contentType: "",
			expected:    false,
		},
		{
			name:        "html",
			contentType: "text/html;",
			expected:    true,
		},
		{
			name:        "pdf",
			contentType: "application/pdf;",
			expected:    false,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			header := http.Header{}
			if tt.name != "no_content_type" {
				header.Set("Content-Type", tt.contentType)
			}
			actual := isValidContentType(&header)
			if tt.expected != actual {
				t.Errorf("isValidContentType: expected %v, got: %v", tt.expected, actual)
			}
		})
	}
}

func TestParseContentType(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "only_type",
			input:    "foo/bar",
			expected: "foo/bar",
		},
		{
			name:     "only_type_semicolumn",
			input:    "foo/bar;",
			expected: "foo/bar",
		},
		{
			name:     "type_and_charset",
			input:    "foo/bar; foo",
			expected: "foo/bar",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual := parseContentType(tt.input)
			if tt.expected != actual {
				t.Errorf("parseContentType(%s): expected %s, got: %s", tt.input, tt.expected, actual)
			}
		})
	}

}

func TestIsExternalDomain(t *testing.T) {
	testcases := []struct {
		name     string
		link     string
		root     string
		expected bool
	}{
		{
			name:     "external_domain",
			link:     "https://www.foo.com/bar",
			root:     "https://www.bar.com/foo",
			expected: true,
		},
		{
			name:     "internal_domain",
			link:     "https://www.foo.com/bar",
			root:     "https://www.foo.com/something",
			expected: false,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := isExternalDomain(tt.link, tt.root)
			if err != nil {
				t.Errorf("isExternalDomain(%s): expected nil error, got: %s\n", tt.link, err)
			}
			if tt.expected != actual {
				t.Errorf("isExternalDomain(%s), expected %v, got: %v\n", tt.link, tt.expected, actual)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	testcases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "empty_url",
			url:      "",
			expected: "",
		},
		{
			name:     "no_trailing",
			url:      "http://foo/bar",
			expected: "http://foo/bar",
		},
		{
			name:     "trailing",
			url:      "http://foo/bar/",
			expected: "http://foo/bar",
		},
		{
			name:     "upper_cases_no_trailing",
			url:      "http://Foo/Bar",
			expected: "http://foo/bar",
		},
		{
			name:     "upper_cases_trailing",
			url:      "http://Foo/Bar/",
			expected: "http://foo/bar",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual := normalizeURL(tt.url)
			if tt.expected != actual {
				t.Errorf("normalizeURL(%s), expected %v, got: %v\n", tt.url, tt.expected, actual)
			}
		})
	}
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
				t.Errorf("getLinksFromURL(%s): expected nil error, got: %s\n", tt.url, err)
			}
			if !areLinksEqual(tt.expected, actual) {
				t.Errorf("getLinksFromURL(%s): expected %v, got: %v\n", tt.url, tt.expected, actual)
			}
		})
	}
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
