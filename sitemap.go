package main

import (
    "fmt"
    "golang.org/x/net/html"
    "net/http"
    "net/url"
    "sort"
    "strings"
    "os"
    "strconv"
)

// This structure contains the URL of a page and its links to static assets, 
// other pages on the domain and external sites
type PageInfo struct {
    URL         *url.URL    // URL of page
    Scripts     []string    // Links to scripts on the page 
    Images      []string    // Links to images on the page
    Files       []string    // Links to to other files (e.g., CSS)
    Links       []string    // Links to other pages on the domain
    External    []string    // Links to external sites (e.g., social media and other sub-domains)
}

// Method to print a PageInfo object
func (u PageInfo) Print() {
    fmt.Printf("%s%s. (%d, %d, %d, %d)\n", u.URL.Host, u.URL.Path, len(u.Scripts), len(u.Files), len(u.Images), len(u.External))
}

// Method to print a PageInfo object with indentation
func (u PageInfo) PrintIndent(indent int) {
    for i := 1; i < indent; i++ {
        fmt.Printf("    ")
    }
    u.Print()
}

// This structure maintains information for the crawler
type Crawler struct {
    RestrictDomain bool             // If true, crawls only the root domain (i.e., no external or subdomains)
    MaxDepth  int                   // Maximum depth to follow from root
    URLInfo   map[string]PageInfo   // Cache of URLs reached and their associated links
    NumURLs   int                   // Number of uniqe URLs crawled
}

// This method starts the crawling process for a given URL and the maximum depth to follow
func (c *Crawler) Init(crawlURL string, restrictDomain bool, maxDepth int) (bool, string) {
    parsedURL, err := url.Parse(crawlURL)
    if err != nil {
        errorString := fmt.Sprintf("Failed to parse URL: %s.\n", crawlURL)
        return false, errorString
    }
    if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
        errorString := fmt.Sprintf("Unsupported scheme: %s. Supported schemes: http, https.\n", parsedURL.Scheme)
        return false, errorString
    }
    c.RestrictDomain = restrictDomain
    c.MaxDepth = maxDepth
    c.URLInfo  = make(map[string]PageInfo)
    return true, "Success"
}

func (c *Crawler) Crawl(crawlURL string) {
    c.crawl(crawlURL, 1)
}

// This method does a recursive Depth-First Search crawl 
func (c *Crawler) crawl(crawlURL string, depth int) {
    if (depth > c.MaxDepth) {
        return
    }
    parsedURL, err := url.Parse(crawlURL)
    if err != nil {
        // Bad url
        return
    }
    urlToCrawl := parsedURL.String()
    var pageInfo PageInfo
    var hasInfo bool
    if pageInfo, hasInfo = c.URLInfo[urlToCrawl]; hasInfo == false {
        pageInfo, hasInfo = c.fetchURL(urlToCrawl)
        if hasInfo == true {
            c.URLInfo[urlToCrawl] = pageInfo
        } else {
            pageInfo.URL.Path = pageInfo.URL.Path + " -> Failed to Fetch (!!)"
        }
    }
    if (hasInfo != true) {
        return;
    }
    pageInfo.PrintIndent(depth)
    if (depth != c.MaxDepth) {
        for _, link := range pageInfo.Links {
            c.crawl(link, depth+1)
        }
    }
}

// Crawler helper method to fetch the URL. Returns a PageInfo object and a boolean success/failure status
func (c *Crawler) fetchURL(urlString string) (PageInfo, bool) {
    c.NumURLs++
    var pageInfo PageInfo
    parsedURL, parseError := url.Parse(urlString)
    if parseError != nil {
        return pageInfo, false
    }
    pageInfo.URL = parsedURL
    host := parsedURL.Host
    scheme := parsedURL.Scheme
    response, err := http.Get(urlString)
    if err != nil {
        return pageInfo, false
    } else {
        defer response.Body.Close()
        document, err := html.Parse(response.Body)
        if err != nil {
            return pageInfo, false
        }
        links := make(map[string]string)        // Contains a unique list of links => tag/type in the page
        var parse func(*html.Node)              // Helper function to parse through a valid HTML page
        parse = func(n *html.Node) {
            tag := strings.ToLower(n.Data)
	        if n.Type == html.ElementNode {
                switch tag {
                case "script":  fallthrough
                case "link":    fallthrough
                case "img":     fallthrough
                case "area":    fallthrough
                case "a":
                    for _, attr := range n.Attr {
                        key := strings.ToLower(attr.Key)
                        if len(attr.Val) > 0 && (key == "href" || key == "src") {
                            schemeRelativeURL := false
                            if len(attr.Val) >= 2 {
                                if attr.Val[0] == '/' && attr.Val[1] == '/' {
                                    schemeRelativeURL = true
                                }
                            }
                            if attr.Val[0] == '/' && schemeRelativeURL == false {
                                // Relative URLs - those that begin with a single / 
                                // so construct URL with the same scheme and host 
                                // of parent path
                                var link url.URL
                                link.Scheme = scheme
                                link.Host   = host
                                link.Path   = attr.Val
                                links[link.String()] = tag
                            } else {
                                // Absolute URL, so include as is
                                var absoluteURL string
                                if schemeRelativeURL {
                                    absoluteURL = scheme + ":" + attr.Val
                                } else {
                                    absoluteURL = attr.Val
                                }
                                parsedLink, err := url.Parse(absoluteURL)
                                if err == nil {
                                    if c.RestrictDomain == false {
                                        links[absoluteURL] = tag
                                    } else {
                                        if parsedLink.Host == host {
                                            links[absoluteURL] = tag
                                        } else {
                                            links[absoluteURL] = "external"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
	        }
	        for child := n.FirstChild; child != nil; child = child.NextSibling {
		        parse(child)
	        }
        }
        parse(document)
        for link, linkType := range links {
            switch linkType {
            case "script":   pageInfo.Scripts   = append(pageInfo.Scripts, link)
            case "link":     pageInfo.Files     = append(pageInfo.Files, link)
            case "img":      pageInfo.Images    = append(pageInfo.Images, link)
            case "external": pageInfo.External  = append(pageInfo.External, link)
            case "area":     pageInfo.Links     = append(pageInfo.Links, link)
            case "a":        pageInfo.Links     = append(pageInfo.Links, link)
            }
        }
    }
    // Sort links to pages so that they can be displayed in alphabetical order
    sort.Strings(pageInfo.Links)
    return pageInfo, true
}

func InputCheck(domainInput string, depthInput string) (bool, int, string) {
    restrictDomain, errB := strconv.ParseBool(domainInput)
    depth, errI := strconv.Atoi(depthInput)
    if errI != nil {
        err := fmt.Sprintf("Error: max-depth must be an integer.")
        return false, 0, err
    }
    if depth < 1 {
        err := fmt.Sprintf("Error: max-depth must be greater than or equal to 1.")
        return false, 0, err
    }
    if errB != nil {
        err := fmt.Sprintf("Error: restrict-to-domain must be an boolean value (e.g., true, false, 1, 0).")
        return false, 0, err
    }
    return restrictDomain, depth, ""
}

func main() {
    argsCount := len(os.Args[1:])
    if argsCount != 3 {
        fmt.Printf("usage: %s site restrict-to-domain max-depth\n", os.Args[0])
        fmt.Printf("   Example: %s http://yathin.com false 2\n", os.Args[0])
        os.Exit(1)
    } else {
        site := os.Args[1]
        restrictDomain, depth, errString := InputCheck(os.Args[2], os.Args[3])
        if (errString != "") {
            fmt.Printf("%s\n", errString)
            os.Exit(1)
        }
        var crawler Crawler
        success, errString := crawler.Init(site, restrictDomain, depth);
        if success {
            fmt.Println("Output: Path (Number of Scripts, Number of Files (e.g, CSS), Number of Images, Number of External Links)")
            crawler.Crawl(site)
        } else {
            fmt.Printf("Error: %s\n", errString)
        }
    }
}
