package scraper

import (
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SeoData struct {
	URL             string
	Title           string
	H1              string
	MetaDescription string
	StatusCode      int
}

type DefaultParser struct{}

type Parser interface {
	GetSeoData(resp *http.Response) (SeoData, error)
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36 Edg/126.0.2592.81",
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func randUserAgent() string {
	return userAgents[rnd.Int()%len(userAgents)]
}

func isSiteMap(urls []string) ([]string, []string) {
	sitemaps := []string{}
	pages := []string{}
	for _, page := range urls {
		if strings.Contains(page, "xml") {
			sitemaps = append(sitemaps, page)
		} else {
			pages = append(pages, page)
		}
	}
	return sitemaps, pages
}

func extractSiteMapUrls(startUrl string, concurrency int) []string {
	// 1. Buffered channel solves deadlock risk by safely holding dynamic responses
	Worklist := make(chan []string, 100000)
	toCrawl := []string{}
	var n int
	n = 1
	var mu sync.Mutex

	// 2. Limiting processing tokens to prevent RAM explosion
	tokens := make(chan struct{}, concurrency)

	Worklist <- []string{startUrl}

	for ; n > 0; n-- {
		list := <-Worklist
		for _, link := range list {
			n++
			go func(link string) {
				// 3. Guarantee exact 1-to-1 Worklist message per goroutine created to perfectly decrement `n`
				var nextWork []string
				defer func() {
					Worklist <- nextWork
				}()

				// Hold semaphore over the ENTIRE job
				tokens <- struct{}{}
				defer func() { <-tokens }()

				res, err := makeRequest(link)
				if err != nil {
					log.Printf("Error retriveing url:%s", link)
					return
				}
				urls, err := extractUrls(res)
				if err != nil {
					log.Printf("Error extracting documents url:%s", link)
					return
				}

				siteMapFiles, pages := isSiteMap(urls)
				if len(siteMapFiles) > 0 {
					nextWork = siteMapFiles
				}

				mu.Lock()
				toCrawl = append(toCrawl, pages...)
				mu.Unlock()
			}(link)
		}
	}

	return toCrawl
}

func makeRequest(url string) (*http.Response, error) {
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", randUserAgent())
	return client.Do(req)
}

func scrapeUrls(urls []string, parser Parser, concurrency int) []SeoData {
	tokens := make(chan struct{}, concurrency)
	var n int
	n = 1
	var mu sync.Mutex
	Worklist := make(chan []string, 100000)
	results := []SeoData{}

	Worklist <- urls

	for ; n > 0; n-- {
		list := <-Worklist
		for _, url := range list {
			if url != "" {
				n++
				go func(url string) {
					// Enforce strictly one reply per goroutine to keep `n` safe
					defer func() { Worklist <- []string{} }()

					// Hold semaphore for the entire page load + scrape process
					tokens <- struct{}{}
					defer func() { <-tokens }()

					log.Printf("Requesting URL:%s", url)
					res, err := scrapePage(url, parser)
					if err != nil {
						log.Printf("Encountered error URL:%s", url)
					} else {
						mu.Lock()
						results = append(results, res)
						mu.Unlock()
					}
				}(url)
			}
		}
	}
	return results
}

func scrapePage(url string, parser Parser) (SeoData, error) {
	res, err := makeRequest(url)
	if err != nil {
		return SeoData{}, err
	}
	data, err := parser.GetSeoData(res)
	if err != nil {
		return SeoData{}, err
	}
	return data, nil
}

func (d DefaultParser) GetSeoData(resp *http.Response) (SeoData, error) {
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return SeoData{}, err
	}
	result := SeoData{}
	result.URL = resp.Request.URL.String()
	result.StatusCode = resp.StatusCode
	result.Title = doc.Find("title").First().Text()
	result.H1 = doc.Find("h1").First().Text()
	result.MetaDescription, _ = doc.Find("meta[name^=description]").Attr("content")
	return result, nil
}

func extractUrls(resp *http.Response) ([]string, error) {
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	results := []string{}
	sel := doc.Find("loc")
	for i := range sel.Nodes {
		loc := sel.Eq(i)
		results = append(results, loc.Text())
	}
	return results, nil
}

func ScrapeSiteMap(url string, parser Parser, concurrency int) []SeoData {
	results := extractSiteMapUrls(url, concurrency)
	res := scrapeUrls(results, parser, concurrency)
	return res
}
