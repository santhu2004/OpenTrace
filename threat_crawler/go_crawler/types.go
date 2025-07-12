package main

import (
	"time"
)

// CrawlConfig represents the input configuration for the crawler
type CrawlConfig struct {
	TargetURL      string `json:"target_url"`      // Starting URL to crawl
	MaxDepth       int    `json:"max_depth"`       // Maximum crawl depth
	MaxLinks       int    `json:"max_links"`       // Maximum links to crawl
	Timeout        string `json:"timeout"`         // Request timeout (e.g., "30s")
	MaxConcurrency int    `json:"max_concurrency"` // Maximum concurrent requests
	UserAgent      string `json:"user_agent"`      // User agent string
	RespectRobots  bool   `json:"respect_robots"`  // Whether to respect robots.txt
}

// CrawlResult represents a single crawled page result
type CrawlResult struct {
	URL           string            `json:"url"`             // The crawled URL
	Status        int               `json:"status"`          // HTTP status code
	Title         string            `json:"title"`           // Page title
	Depth         int               `json:"depth"`           // Crawl depth where found
	Discovered    time.Time         `json:"discovered"`      // When this page was discovered
	Headers       map[string]string `json:"headers"`         // Response headers
	Links         []string          `json:"links"`           // All links found on page
	InternalLinks []string          `json:"internal_links"`  // Same-domain links
	ExternalLinks []string          `json:"external_links"`  // Cross-domain links
	Error         string            `json:"error,omitempty"` // Error message if any
}

// CrawlOutput represents the complete output from the crawler
type CrawlOutput struct {
	Config    CrawlConfig   `json:"config"`    // Input configuration used
	Results   []CrawlResult `json:"results"`   // All crawled pages
	Summary   CrawlSummary  `json:"summary"`   // Crawl statistics
	Timestamp time.Time     `json:"timestamp"` // When crawl completed
}

// CrawlSummary contains statistics about the crawl
type CrawlSummary struct {
	TotalPages      int           `json:"total_pages"`       // Total pages crawled
	Successful      int           `json:"successful"`        // Pages with 200 status
	Failed          int           `json:"failed"`            // Pages that failed
	InternalLinks   int           `json:"internal_links"`    // Total internal links found
	ExternalLinks   int           `json:"external_links"`    // Total external links found
	MaxDepthReached int           `json:"max_depth_reached"` // Maximum depth actually reached
	Duration        time.Duration `json:"duration"`          // Total crawl duration
}

// Page represents a page in the crawl queue
type Page struct {
	URL   string
	Depth int
}

// VisitedPage tracks visited pages to avoid duplicates
type VisitedPage struct {
	URL       string
	Depth     int
	Timestamp time.Time
}
