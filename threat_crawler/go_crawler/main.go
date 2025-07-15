package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("[GO-DEBUG] Go main() started")
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	log.Println("[DEBUG] Go crawler started")

	// --- CLI flags for v3.0 ---
	startURL := flag.String("start_url", "", "Starting URL to crawl (required)")
	maxDepth := flag.Int("max_depth", 3, "Maximum crawl depth")
	maxPages := flag.Int("max_pages", 100, "Maximum number of pages to crawl")
	timeout := flag.String("timeout", "30s", "Request timeout (e.g., 30s)")
	userAgent := flag.String("user_agent", "ThreatCrawler/3.0", "User agent string")
	workers := flag.Int("workers", 10, "Number of concurrent workers")
	flag.Parse()

	if *startURL == "" {
		fmt.Fprintln(os.Stderr, "Error: --start_url is required")
		os.Exit(1)
	}

	// Handle interrupts gracefully with context cancellation and global timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("[INFO] Interrupt received, shutting down...")
		cancel()
	}()

	// --- Prepare config ---
	config := &CrawlConfig{
		TargetURL:      *startURL,
		MaxDepth:       *maxDepth,
		MaxLinks:       *maxPages,
		Timeout:        *timeout,
		MaxConcurrency: *workers,
		UserAgent:      *userAgent,
		RespectRobots:  false, // Not implemented
	}

	// --- Start crawl ---
	crawler := NewCrawler(config)
	resultsCh := make(chan *CrawlResult, *workers)

	// Inactivity timer: cancel context if no results for 30s
	inactivity := 30 * time.Second
	lastActivity := time.Now()
	activityCh := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case <-activityCh:
				lastActivity = time.Now()
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			if time.Since(lastActivity) > inactivity {
				log.Println("[MONITOR] Inactivity timeout reached. Cancelling context.")
				cancel()
				return
			}
		}
	}()

	go func() {
		crawler.CrawlStream(ctx, resultsCh)
		// Do NOT close(resultsCh) here; CrawlStream handles it
	}()

	// Output each result as a single JSON object per line
	encoder := json.NewEncoder(os.Stdout)
	for result := range resultsCh {
		select {
		case activityCh <- struct{}{}:
		default:
		}
		encoder.Encode(result)
	}

	log.Println("[DEBUG] Go crawler finished. Exiting gracefully.")
	os.Exit(0)
}
