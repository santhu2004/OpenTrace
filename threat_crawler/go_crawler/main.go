package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	// Command line flags
	var (
		inputFile  = flag.String("input", "", "Input JSON configuration file")
		outputFile = flag.String("output", "", "Output JSON results file (default: stdout)")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Set up logging
	if !*verbose {
		log.SetOutput(os.Stderr)
		log.SetFlags(0) // Disable timestamp for cleaner output
	}

	// Read configuration
	config, err := readConfig(*inputFile)
	if err != nil {
		log.Fatalf("Failed to read configuration: %v", err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create and run crawler
	crawler := NewCrawler(config)
	log.Printf("Starting crawl of %s (max depth: %d, max links: %d)",
		config.TargetURL, config.MaxDepth, config.MaxLinks)

	// Add a timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run crawler in a goroutine with timeout
	done := make(chan *CrawlOutput, 1)
	go func() {
		output := crawler.Crawl()
		done <- output
	}()

	var output *CrawlOutput
	select {
	case output = <-done:
		// Crawl completed successfully
	case <-ctx.Done():
		log.Fatalf("Crawl timed out after 30 seconds")
	}

	// Write results
	if err := writeOutput(output, *outputFile); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	log.Printf("Crawl completed: %d pages, %d successful, %d failed, duration: %v",
		output.Summary.TotalPages, output.Summary.Successful, output.Summary.Failed, output.Summary.Duration)
}

// readConfig reads configuration from file or stdin
func readConfig(inputFile string) (*CrawlConfig, error) {
	var reader *os.File
	var err error

	if inputFile == "" {
		// Read from stdin
		reader = os.Stdin
	} else {
		// Read from file
		reader, err = os.Open(inputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open input file: %v", err)
		}
		defer reader.Close()
	}

	// Parse JSON
	var config CrawlConfig
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Set defaults for missing values
	setDefaults(&config)

	return &config, nil
}

// setDefaults sets default values for missing configuration fields
func setDefaults(config *CrawlConfig) {
	if config.Timeout == "" {
		config.Timeout = "30s"
	}
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = 10
	}
	if config.UserAgent == "" {
		config.UserAgent = "ThreatCrawler/1.0"
	}
	if config.MaxLinks == 0 {
		config.MaxLinks = 100
	}
	if config.MaxDepth == 0 {
		config.MaxDepth = 3
	}
}

// validateConfig validates the configuration
func validateConfig(config *CrawlConfig) error {
	if config.TargetURL == "" {
		return fmt.Errorf("target_url is required")
	}
	if config.MaxDepth < 0 {
		return fmt.Errorf("max_depth must be >= 0")
	}
	if config.MaxLinks <= 0 {
		return fmt.Errorf("max_links must be > 0")
	}
	if config.MaxConcurrency <= 0 {
		return fmt.Errorf("max_concurrency must be > 0")
	}
	if config.Timeout == "" {
		return fmt.Errorf("timeout is required")
	}
	return nil
}

// writeOutput writes results to file or stdout
func writeOutput(output *CrawlOutput, outputFile string) error {
	var writer *os.File
	var err error

	if outputFile == "" {
		// Write to stdout
		writer = os.Stdout
	} else {
		// Write to file
		writer, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer writer.Close()
	}

	// Encode JSON with pretty formatting
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	return nil
}
