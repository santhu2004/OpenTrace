package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var ()

// Crawler manages the crawling process with concurrency and depth control
type Crawler struct {
	config       *CrawlConfig
	fetcher      *Fetcher
	visited      map[string]VisitedPage
	visitedMux   sync.RWMutex
	results      []*CrawlResult
	resultsMux   sync.Mutex
	queue        chan Page
	wg           sync.WaitGroup
	activeJobs   int32
	startTime    time.Time
	resultsCount int32 // atomic counter for results
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

	// Strict atomic check for max links
	if atomic.LoadInt32(&c.resultsCount) >= int32(c.config.MaxLinks) {
		return
	}

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

	// Atomically increment and check if we can add this result
	if atomic.AddInt32(&c.resultsCount, 1) > int32(c.config.MaxLinks) {
		// Exceeded limit, do not add this result
		return
	}

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

// CrawlStream performs the crawl and streams each result to the provided channel as soon as it's available.
func (c *Crawler) CrawlStream(ctx context.Context, out chan<- *CrawlResult) {
	type QueueItem struct {
		URL    string
		Depth  int
		Parent string
	}

	visited := make(map[string]struct{})
	var visitedMux sync.RWMutex
	queue := make(chan QueueItem, c.config.MaxLinks*2)
	var wg sync.WaitGroup       // For worker goroutines
	var inFlight sync.WaitGroup // For in-flight work
	done := make(chan struct{}) // Signal for forced shutdown
	var closeOnce sync.Once     // Guard for closing out

	normalize := func(url string) string {
		return strings.TrimRight(url, "/")
	}

	// Worker function (with defer recover to catch panics)
	worker := func(id int) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[WORKER %d] Panic recovered: %v", id, r)
			}
		}()
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Printf("[WORKER %d] Context cancelled, exiting", id)
				return
			case <-done:
				log.Printf("[WORKER %d] Forced shutdown, exiting", id)
				return
			case item := <-queue:
				func() {
					defer inFlight.Done() // Always call Done for this item

					log.Printf("[WORKER %d] Processing: %s (depth: %d)", id, item.URL, item.Depth)

					if item.Depth > c.config.MaxDepth {
						return
					}

					visitedMux.RLock()
					_, seen := visited[normalize(item.URL)]
					visitedMux.RUnlock()
					if seen {
						return
					}

					visitedMux.Lock()
					visited[normalize(item.URL)] = struct{}{}
					visitedMux.Unlock()

					// Fetch the page using the fetcher
					result, _ := c.fetcher.FetchPage(ctx, item.URL)
					if result != nil {
						result.Depth = item.Depth
						if result.Headers == nil {
							result.Headers = map[string]string{}
						}
						result.Headers["Parent-URL"] = item.Parent
						// Atomically increment resultsCount. If it exceeds maxLinks, return immediately.
						count := atomic.AddInt32(&c.resultsCount, 1)
						if count > int32(c.config.MaxLinks) {
							return
						}
						// When sending to out, always guard with select
						select {
						case <-ctx.Done():
							return
						case <-done:
							return
						case out <- result:
						}
						// Enqueue discovered links
						if atomic.LoadInt32(&c.resultsCount) < int32(c.config.MaxLinks) && item.Depth < c.config.MaxDepth {
							for _, link := range result.Links {
								visitedMux.RLock()
								_, already := visited[normalize(link)]
								visitedMux.RUnlock()
								if !already {
									log.Printf("[WORKER %d] Enqueueing: %s (depth: %d)", id, link, item.Depth+1)
									inFlight.Add(1)
									select {
									case queue <- QueueItem{URL: link, Depth: item.Depth + 1, Parent: item.URL}:
										// enqueued
									case <-ctx.Done():
										inFlight.Done() // undo the Add, since we won't process this link
										return
									case <-done:
										inFlight.Done() // undo the Add, since we won't process this link
										return
									}
								}
							}
						}
					}
				}()
			}
		}
	}

	// Seed the crawl
	inFlight.Add(1)
	queue <- QueueItem{URL: c.config.TargetURL, Depth: 0, Parent: ""}

	// Start workers
	for i := 0; i < c.config.MaxConcurrency; i++ {
		wg.Add(1)
		go worker(i)
	}

	// Monitor goroutine: closes out only after all work is done
	go func() {
		wg.Wait() // Wait for all workers to exit
		log.Printf("[MONITOR] All workers exited. Draining queue...")
		// Drain any remaining items in the queue and mark them done
		queueDrained := 0
		for {
			select {
			case <-ctx.Done():
				for {
					select {
					case <-queue:
						inFlight.Done()
						queueDrained++
					default:
						goto drained
					}
				}
			drained:
				break
			default:
				break
			}
			break
		}
		log.Printf("[MONITOR] Queue drained. Items drained: %d", queueDrained)
		// Log inFlight count before waiting
		log.Printf("[MONITOR] inFlight before Wait: (should be 0) ...")
		// Start a hard shutdown timer
		doneCh := make(chan struct{})
		go func() {
			inFlight.Wait()
			close(doneCh)
		}()
		select {
		case <-doneCh:
			log.Printf("[MONITOR] inFlight.Wait() returned. Proceeding to close output.")
		case <-time.After(10 * time.Second):
			log.Printf("[MONITOR] FATAL: inFlight.Wait() did not return after 10s. Forcing exit.")
			os.Exit(2)
		}
		closeOnce.Do(func() {
			close(out)
		})
	}()

	// Forced shutdown monitor
	go func() {
		select {
		case <-ctx.Done():
			log.Println("[MONITOR] Context cancelled, forcing shutdown...")
			close(done)
		}
	}()
}
