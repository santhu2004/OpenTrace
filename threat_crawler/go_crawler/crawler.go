package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Crawler manages the crawling process with concurrency and depth control
type Crawler struct {
	config     *CrawlConfig
	fetcher    *Fetcher
	visited    map[string]VisitedPage
	visitedMux sync.RWMutex
	results    []*CrawlResult
	resultsMux sync.Mutex
	queue      chan Page
	wg         sync.WaitGroup
	activeJobs int32
	startTime  time.Time
}

// NewCrawler creates a new crawler instance
func NewCrawler(config *CrawlConfig) *Crawler {
	fetcher := NewFetcher(config.Timeout, config.UserAgent)

	return &Crawler{
		config:    config,
		fetcher:   fetcher,
		visited:   make(map[string]VisitedPage),
		results:   make([]*CrawlResult, 0),
		queue:     make(chan Page, config.MaxLinks*2), // Buffer for queue
		startTime: time.Now(),
	}
}

// Crawl starts the crawling process
func (c *Crawler) Crawl() *CrawlOutput {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker goroutines
	c.wg.Add(c.config.MaxConcurrency)
	for i := 0; i < c.config.MaxConcurrency; i++ {
		go c.worker(ctx)
	}

	// Add initial page to queue
	c.queue <- Page{URL: c.config.TargetURL, Depth: 0}

	// Start a goroutine to monitor and close queue when done
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)

			// Check if we've reached max links
			c.resultsMux.Lock()
			currentCount := len(c.results)
			c.resultsMux.Unlock()

			if currentCount >= c.config.MaxLinks {
				close(c.queue)
				return
			}

			// Check if no active jobs and queue is empty
			if atomic.LoadInt32(&c.activeJobs) == 0 && len(c.queue) == 0 {
				close(c.queue)
				return
			}
		}
	}()

	// Wait for all workers to finish
	c.wg.Wait()

	// Generate summary
	summary := c.generateSummary()

	return &CrawlOutput{
		Config:    *c.config,
		Results:   c.getResults(),
		Summary:   summary,
		Timestamp: time.Now(),
	}
}

// worker processes pages from the queue
func (c *Crawler) worker(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case page, ok := <-c.queue:
			if !ok {
				// Queue is closed, exit
				return
			}
			c.processPage(ctx, page)
		}
	}
}

// processPage handles a single page crawl
func (c *Crawler) processPage(ctx context.Context, page Page) {
	// Increment active jobs counter
	atomic.AddInt32(&c.activeJobs, 1)
	defer atomic.AddInt32(&c.activeJobs, -1)

	// Check if already visited
	c.visitedMux.RLock()
	if _, exists := c.visited[page.URL]; exists {
		c.visitedMux.RUnlock()
		return
	}
	c.visitedMux.RUnlock()

	// Mark as visited
	c.visitedMux.Lock()
	c.visited[page.URL] = VisitedPage{
		URL:       page.URL,
		Depth:     page.Depth,
		Timestamp: time.Now(),
	}
	c.visitedMux.Unlock()

	// Check depth limit
	if page.Depth > c.config.MaxDepth {
		return
	}

	// Check if we've reached max links
	c.resultsMux.Lock()
	if len(c.results) >= c.config.MaxLinks {
		c.resultsMux.Unlock()
		return
	}
	c.resultsMux.Unlock()

	// Fetch the page
	log.Printf("Fetching page: %s", page.URL)
	result, err := c.fetcher.FetchPage(ctx, page.URL)
	if err != nil {
		log.Printf("Error fetching %s: %v", page.URL, err)
		result = &CrawlResult{
			URL:    page.URL,
			Status: 0,
			Error:  err.Error(),
		}
	} else {
		log.Printf("Successfully fetched %s (status: %d)", page.URL, result.Status)
	}

	// Set depth
	result.Depth = page.Depth

	// Add to results
	c.resultsMux.Lock()
	c.results = append(c.results, result)
	currentCount := len(c.results)
	c.resultsMux.Unlock()

	// Log progress
	log.Printf("Crawled %s (depth: %d, status: %d) - Total: %d/%d",
		page.URL, page.Depth, result.Status, currentCount, c.config.MaxLinks)

	// Add internal links to queue if we haven't reached limits
	if page.Depth < c.config.MaxDepth && currentCount < c.config.MaxLinks {
		c.addLinksToQueue(result.InternalLinks, page.Depth+1)
	}
}

// addLinksToQueue adds new links to the crawl queue
func (c *Crawler) addLinksToQueue(links []string, depth int) {
	for _, link := range links {
		// Check if already visited
		c.visitedMux.RLock()
		if _, exists := c.visited[link]; exists {
			c.visitedMux.RUnlock()
			continue
		}
		c.visitedMux.RUnlock()

		// Check if queue is full
		select {
		case c.queue <- Page{URL: link, Depth: depth}:
			// Successfully added to queue
		default:
			// Queue is full, skip this link
			log.Printf("Queue full, skipping %s", link)
		}
	}
}

// getResults returns a copy of all results
func (c *Crawler) getResults() []CrawlResult {
	c.resultsMux.Lock()
	defer c.resultsMux.Unlock()

	results := make([]CrawlResult, len(c.results))
	for i, result := range c.results {
		results[i] = *result
	}
	return results
}

// generateSummary creates crawl statistics
func (c *Crawler) generateSummary() CrawlSummary {
	c.resultsMux.Lock()
	defer c.resultsMux.Unlock()

	summary := CrawlSummary{
		TotalPages: len(c.results),
		Duration:   time.Since(c.startTime),
	}

	maxDepth := 0
	totalInternal := 0
	totalExternal := 0

	for _, result := range c.results {
		if result.Status == 200 {
			summary.Successful++
		} else {
			summary.Failed++
		}

		if result.Depth > maxDepth {
			maxDepth = result.Depth
		}

		totalInternal += len(result.InternalLinks)
		totalExternal += len(result.ExternalLinks)
	}

	summary.MaxDepthReached = maxDepth
	summary.InternalLinks = totalInternal
	summary.ExternalLinks = totalExternal

	return summary
}

// GetVisitedCount returns the number of visited pages
func (c *Crawler) GetVisitedCount() int {
	c.visitedMux.RLock()
	defer c.visitedMux.RUnlock()
	return len(c.visited)
}
