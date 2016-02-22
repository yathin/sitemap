package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
    "testing"
)

func TestUserInput(t *testing.T) {
    cases := []struct {
        inDomain string
        inDepth string
        outDomain bool
        outDepth int
        outErr string
    }{
        {"true","1",true,1,""},
        {"FALSE","2",false,2,""},
    }
    for _, c := range cases {
        outDomain, outDepth, outErr := InputCheck(c.inDomain, c.inDepth)
        if outErr != c.outErr {
            t.Errorf("InputCheck(%q, %q) == (%q %q %q), needed (%q %q %q)", c.inDomain, c.inDepth,  outDomain, outDepth, outErr, c.outDomain, c.outDepth, c.outErr)
        }
    }
}

func TestCrawlerInit(t *testing.T) {
    cases := []struct {
        url string
        out bool
    }{
        {"ftp://random.com", false},
        {"http://random.com", true},
        {"https://random.com", true},
    }
    for _, c := range cases {
        var crawler Crawler
        out, _ := crawler.Init(c.url, true, 1)
        if out != c.out {
            t.Errorf("Init(%q, true, 1) == %q, needed %q", c.url, out, c.out)
        }
    }
}

func TestFetchURL(t *testing.T) {
    cases := []struct {
        content string
        restrictDomain bool
        maxDepth int
        outStatus bool
        outLinks int
        outJS int
        outFiles int
        outImages int
        outExternal int
    }{
        {`{"json":"false"}<a href="http://127.0.0.1/"`, false, 1, true, 0, 0, 0, 0, 0},
        {`<html><body><a href="http://abc.com/abc">abc</a><script src="//1.js"></script></body></html>`, false, 1, true, 1, 1, 0, 0, 0},
        {`<html><head><link href="//abc.css"></head><body><script src="//1.js"></script></body></html>`, false, 1, true, 0, 1, 1, 0, 0},
        {`<html><body><img src="/abc.jpg">foo</a><script src="//1.js"></script></body></html>`, false, 1, true, 0, 1, 0, 1, 0},
        {`<html><body><a href="http://127.0.0.1/">foo</a><script src="//1.js"></script></body></html>`, true, 1, true, 0, 0, 0, 0, 2},
    }
    for _, c := range cases {
        ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "text/html")
            fmt.Fprintln(w, c.content)
        }))
        defer ts.Close()

        var crawler Crawler
        out, _ := crawler.Init(ts.URL, c.restrictDomain, c.maxDepth)
        if (out == false) {
            t.Errorf("Init(%q, 1) == %q, expected %q", ts.URL, true, out)
        }
        pageInfo, status := crawler.fetchURL(ts.URL)
        if status != c.outStatus {
            fmt.Printf("Content: %s\n", c.content)
            t.Errorf("fetchURL(%q) == %q, expected %q", ts.URL, true, status)
        } else {
            if len(pageInfo.Links) != c.outLinks {
                t.Errorf("Links count for fetchURL(%q) == %q, expected %q", ts.URL, len(pageInfo.Links), c.outLinks)
            }
            if len(pageInfo.Scripts) != c.outJS {
                t.Errorf("JS count for fetchURL(%q) == %q, expected %q", ts.URL, len(pageInfo.Scripts), c.outJS)
            }
            if len(pageInfo.Files) != c.outFiles {
                t.Errorf("File count for fetchURL(%q) == %q, expected %q", ts.URL, len(pageInfo.Files), c.outFiles)
            }
            if len(pageInfo.Images) != c.outImages {
                t.Errorf("JS count for fetchURL(%q) == %q, expected %q", ts.URL, len(pageInfo.Images), c.outImages)
            }
            if len(pageInfo.External) != c.outExternal {
                t.Errorf("File count for fetchURL(%q) == %q, expected %q", ts.URL, len(pageInfo.External), c.outExternal)
            }
        }
    }
}

func TestCrawler(t *testing.T) {
    cases := []struct {
        content string
        restrictDomain bool
        maxDepth int
        numUrls int
    }{
        {`<html><body><a href="http://abc.com/abc">abc</a><script src="//1.js"></script></body></html>`, false, 1, 1},
        {`<html><body><a href="http://abc.com/abc">abc</a><script src="//1.js"></script></body></html>`, false, 2, 2},
        {`<html><body><a href="http://abc.com/abc">abc</a><a href="http://123">123</a></body></html>`, false, 1, 1},
        {`<html><body><a href="http://abc.com/abc">abc</a><a href="http://123">123</a></body></html>`, false, 2, 3},
    }
    for _, c := range cases {
        ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "text/html")
            fmt.Fprintln(w, c.content)
        }))
        defer ts.Close()

        var crawler Crawler
        out, _ := crawler.Init(ts.URL, c.restrictDomain, c.maxDepth)
        if (out == false) {
            t.Errorf("Init(%q, 1) == %q, expected %q", ts.URL, true, out)
        }
        crawler.Crawl(ts.URL)
        if crawler.NumURLs != c.numUrls {
            t.Errorf("URLs crawled for Crawl(%q) == %q, expected %q", ts.URL, crawler.NumURLs, c.numUrls)
        }
    }
}
