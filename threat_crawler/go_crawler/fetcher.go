package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Fetcher handles HTTP requests with proper timeouts and error handling
type Fetcher struct {
	client    *http.Client
	userAgent string
}

// NewFetcher creates a new fetcher with the given configuration
func NewFetcher(timeoutStr string, userAgent string) *Fetcher {
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		// Default to 30 seconds if parsing fails
		timeout = 30 * time.Second
	}

	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
		Jar: jar,
	}

	return &Fetcher{
		client:    client,
		userAgent: userAgent,
	}
}

// FetchPage retrieves a single page with proper error handling
func (f *Fetcher) FetchPage(ctx context.Context, url string) (*CrawlResult, error) {
	// Create request with context for cancellation
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &CrawlResult{
			URL:    url,
			Status: 0,
			Error:  fmt.Sprintf("Failed to create request: %v", err),
		}, nil
	}

	// Set user agent
	if f.userAgent != "" {
		req.Header.Set("User-Agent", f.userAgent)
	} else {
		req.Header.Set("User-Agent", "ThreatCrawler/1.0")
	}

	// Add common headers
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	// req.Header.Set("Accept-Encoding", "gzip, deflate") // Let Go handle decompression automatically
	req.Header.Set("Connection", "keep-alive")

	// Make the request
	resp, err := f.client.Do(req)
	if err != nil {
		return &CrawlResult{
			URL:    url,
			Status: 0,
			Error:  fmt.Sprintf("Request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &CrawlResult{
			URL:    url,
			Status: resp.StatusCode,
			Error:  fmt.Sprintf("Failed to read response: %v", err),
		}, nil
	}

	// Extract headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	// Extract title from HTML
	title := extractTitle(string(body))

	// Extract links from HTML
	links := extractLinks(string(body), url)
	internalLinks, externalLinks := categorizeLinks(links, url)

	// After extracting links
	if links == nil {
		links = []string{}
	}
	if internalLinks == nil {
		internalLinks = []string{}
	}
	if externalLinks == nil {
		externalLinks = []string{}
	}

	return &CrawlResult{
		URL:           url,
		Status:        resp.StatusCode,
		Title:         title,
		Discovered:    time.Now(),
		Headers:       headers,
		Links:         links,
		InternalLinks: internalLinks,
		ExternalLinks: externalLinks,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractTitle extracts the page title from HTML content
func extractTitle(html string) string {
	// Simple title extraction - looks for <title> tag
	titleStart := strings.Index(html, "<title>")
	if titleStart == -1 {
		return "No Title"
	}
	titleStart += 7 // Length of "<title>"

	titleEnd := strings.Index(html[titleStart:], "</title>")
	if titleEnd == -1 {
		return "No Title"
	}

	title := html[titleStart : titleStart+titleEnd]
	return strings.TrimSpace(title)
}

// extractLinks extracts all href, src, and action attributes from any tag in the HTML
func extractLinks(htmlContent, baseURL string) []string {
	var links []string
	seen := make(map[string]bool)

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return links
	}

	var base *url.URL
	base, _ = url.Parse(baseURL)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				key := strings.ToLower(attr.Key)
				if key == "href" || key == "src" || key == "action" {
					val := strings.TrimSpace(attr.Val)
					if val == "" || strings.HasPrefix(val, "#") || strings.HasPrefix(strings.ToLower(val), "javascript:") {
						continue
					}
					u, err := url.Parse(val)
					if err != nil {
						continue
					}
					var abs *url.URL
					if base != nil {
						abs = base.ResolveReference(u)
					} else {
						abs = u
					}
					absStr := abs.String()
					if !seen[absStr] {
						links = append(links, absStr)
						seen[absStr] = true
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return links
}

// categorizeLinks separates internal and external links
func categorizeLinks(links []string, baseURL string) (internal, external []string) {
	baseDomain := extractDomain(baseURL)

	for _, link := range links {
		linkDomain := extractDomain(link)
		if linkDomain == baseDomain {
			internal = append(internal, link)
		} else {
			external = append(external, link)
		}
	}

	return internal, external
}

// extractDomain extracts the domain from a URL and normalizes by removing 'www.'
func extractDomain(rawurl string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}
	host := u.Hostname()
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}
	return host
}
