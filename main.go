package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/RudraPratapDev/SiteMapr/exporter"
	"github.com/RudraPratapDev/SiteMapr/scraper"
)

// Main entry point
func main() {
	reader := bufio.NewReader(os.Stdin)

	// Ask for URL
	fmt.Print("Enter sitemap URL (default: https://www.quicksprout.com/sitemap.xml): ")
	urlInput, _ := reader.ReadString('\n')
	urlInput = strings.TrimSpace(urlInput)
	if urlInput == "" {
		urlInput = "https://www.quicksprout.com/sitemap.xml"
	}

	// Ask for Concurrency
	fmt.Print("Enter concurrent workers (default: 10): ")
	concurrencyInput, _ := reader.ReadString('\n')
	concurrencyInput = strings.TrimSpace(concurrencyInput)
	concurrency := 10
	if concurrencyInput != "" {
		if c, err := strconv.Atoi(concurrencyInput); err == nil && c > 0 {
			concurrency = c
		} else {
			fmt.Println("Invalid input, using default 10.")
		}
	}

	// Ask for export format
	fmt.Print("Enter export format [json, csv] (default: json): ")
	formatInput, _ := reader.ReadString('\n')
	formatInput = strings.TrimSpace(strings.ToLower(formatInput))
	if formatInput != "csv" {
		formatInput = "json"
	}

	fmt.Printf("\nStarting crawler on %s with %d workers...\n", urlInput, concurrency)

	p := scraper.DefaultParser{}
	results := scraper.ScrapeSiteMap(urlInput, p, concurrency)

	fmt.Printf("\nScraping complete! Found %d records.\n", len(results))

	// Export structured data
	filename := "results." + formatInput
	var err error
	if formatInput == "csv" {
		err = exporter.ExportCSV(results, filename)
	} else {
		err = exporter.ExportJSON(results, filename)
	}

	if err != nil {
		fmt.Printf("Error exporting data: %v\n", err)
	} else {
		fmt.Printf("Successfully exported structured data to %s\n", filename)
	}
}
