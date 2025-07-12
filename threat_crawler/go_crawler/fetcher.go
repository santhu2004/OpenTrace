package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
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
	req.Header.Set("Accept-Encoding", "gzip, deflate")
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

// extractLinks extracts all links from HTML content
func extractLinks(html, baseURL string) []string {
	var links []string
	seen := make(map[string]bool) // Avoid duplicates

	// Simple link extraction - looks for href attributes
	// This is a basic implementation; for production, consider using a proper HTML parser
	startIndex := 0
	for {
		// Find the next href attribute
		hrefIndex := strings.Index(html[startIndex:], "href=\"")
		if hrefIndex == -1 {
			break
		}

		// Adjust to absolute position in the string
		hrefIndex += startIndex + 6 // Length of "href=\""

		// Find the closing quote
		quoteEnd := strings.Index(html[hrefIndex:], "\"")
		if quoteEnd == -1 {
			break
		}

		// Extract the link
		link := html[hrefIndex : hrefIndex+quoteEnd]

		// Process the link
		if link != "" && !strings.HasPrefix(link, "#") && !strings.HasPrefix(link, "javascript:") {
			var fullLink string

			// Normalize the link
			if strings.HasPrefix(link, "http") {
				fullLink = link
			} else if strings.HasPrefix(link, "/") {
				// Relative to domain
				fullLink = baseURL + link
			} else {
				// Skip relative links without leading slash
				startIndex = hrefIndex + quoteEnd + 1
				continue
			}

			// Add to results if not seen before
			if !seen[fullLink] {
				links = append(links, fullLink)
				seen[fullLink] = true
			}
		}

		// Move to next position
		startIndex = hrefIndex + quoteEnd + 1

		// Safety check to prevent infinite loops
		if startIndex >= len(html) {
			break
		}
	}

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

// extractDomain extracts the domain from a URL
func extractDomain(url string) string {
	// Remove protocol
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}

	// Get domain part
	slashIndex := strings.Index(url, "/")
	if slashIndex != -1 {
		url = url[:slashIndex]
	}

	return url
}
