package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/RudraPratapDev/SiteMapr/exporter"
	"io"
	"github.com/RudraPratapDev/SiteMapr/scraper"
)

func getDomainName(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return "unknown"
	}
	parts := strings.Split(parsed.Hostname(), ".")
	if len(parts) > 1 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return parsed.Hostname()
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter sitemap URL (default: https://www.quicksprout.com/sitemap.xml): ")
	urlInput, _ := reader.ReadString('\n')
	urlInput = strings.TrimSpace(urlInput)
	if urlInput == "" {
		urlInput = "https://www.quicksprout.com/sitemap.xml"
	}

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

	fmt.Print("Enter export format [json, csv] (default: json): ")
	formatInput, _ := reader.ReadString('\n')
	formatInput = strings.TrimSpace(strings.ToLower(formatInput))
	if formatInput != "csv" {
		formatInput = "json"
	}

	// Create Output Directory specific to website and date
	domain := getDomainName(urlInput)
	timestamp := time.Now().Format("20060102-150405")
	outDir := filepath.Join("results", fmt.Sprintf("%s-%s", domain, timestamp))
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v\n", err)
	}

	// Setup logging to a file in the directory
	logFilePath := filepath.Join(outDir, "logs.txt")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	fmt.Printf("\nStarting crawler on %s with %d workers...\n", urlInput, concurrency)
	log.Printf("Starting crawler on %s with %d workers...", urlInput, concurrency)

	p := scraper.DefaultParser{}
	results := scraper.ScrapeSiteMap(urlInput, p, concurrency)

	fmt.Printf("\nScraping complete! Found %d records.\n", len(results))
	log.Printf("Scraping complete! Found %d records.", len(results))

	filename := filepath.Join(outDir, "data."+formatInput)
	if formatInput == "csv" {
		err = exporter.ExportCSV(results, filename)
	} else {
		err = exporter.ExportJSON(results, filename)
	}

	if err != nil {
		fmt.Printf("Error exporting data: %v\n", err)
		log.Printf("Error exporting data: %v", err)
	} else {
		fmt.Printf("Successfully exported structured data to %s\n", filename)
		fmt.Printf("Successfully exported logs to %s\n", logFilePath)
		log.Printf("Successfully exported structured data to %s", filename)
	}
}
