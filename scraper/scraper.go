package scraper

import (
	"fmt"
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

type DefaultParser struct {
}

type Parser interface {
	GetSeoData(resp *http.Response) (SeoData, error)
}

var userAgents = []string{
	// Chrome - Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",

	// Firefox - Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0",

	// Safari - macOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",

	// Edge - Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36 Edg/126.0.2592.81",
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func randUserAgent() string {
	// pick a random user agent from the pool
	randnum := rnd.Int() % len(userAgents)
	return userAgents[randnum]
}

// checks and splits a list of urls into sitemaps and standard pages
func isSiteMap(urls []string) ([]string, []string) {
	sitemapFiles := []string{}
	pages := []string{}
	for _, page := range urls {
		foundSitemap := strings.Contains(page, "xml")
		if foundSitemap == true {
			fmt.Println("Found sitemap,page")
			sitemapFiles = append(sitemapFiles, page)
		} else {
			pages = append(pages, page)
		}
	}
	return sitemapFiles, pages
}

func extractSiteMapUrls(startUrl string) []string {
	// unbuffered channel to orchestrate worklist tasks
	Worklist := make(chan []string)
	toCrawl := []string{}
	var n int
	n = 1
	var mu sync.Mutex
	go func() { Worklist <- []string{startUrl} }()

	for ; n > 0; n-- {
		list := <-Worklist
		for _, link := range list {
			n++
			go func(link string) {
				res, err := makeRequest(link)
				if err != nil {
					log.Printf("Error retriveing url:%s", link)
					Worklist <- []string{}
					return
				}
				urls, err := extractUrls(res)
				if err != nil {
					log.Printf("Error extracting documents from response url:%s", link)
					Worklist <- []string{}
					return
				}
				siteMapFiles, pages := isSiteMap(urls)
				if siteMapFiles != nil {
					Worklist <- siteMapFiles
				} else {
					Worklist <- []string{}
				}
				mu.Lock()
				for _, page := range pages {
					toCrawl = append(toCrawl, page)
				}
				mu.Unlock()
			}(link)
		}
	}

	return toCrawl
}

func makeRequest(url string) (*http.Response, error) {
	client := http.Client{
		//dont want to overload the website so adding timeout
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", randUserAgent())
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, err
}

func scrapeUrls(urls []string, parser Parser, concurrency int) []SeoData {
	// semaphore channel to limit the maximum number of concurrent requests
	tokens := make(chan struct{}, concurrency)
	var n int
	n = 1
	var mu sync.Mutex
	Worklist := make(chan []string)
	results := []SeoData{}
	go func() {
		Worklist <- urls
	}()

	for ; n > 0; n-- {
		list := <-Worklist
		for _, url := range list {
			if url != "" {
				n++
				go func(url string, token chan struct{}) {
					log.Printf("Requesting URL:%s", url)
					res, err := scrapePage(url, tokens, parser)
					if err != nil {
						log.Printf("Encountered error URL:%s", url)
					} else {
						mu.Lock()
						results = append(results, res)
						mu.Unlock()
					}
					Worklist <- []string{}
				}(url, tokens)
			}
		}
	}
	return results
}

func scrapePage(url string, tokens chan struct{}, parser Parser) (SeoData, error) {
	res, err := crawlPage(url, tokens)
	if err != nil {
		return SeoData{}, err
	}
	data, err := parser.GetSeoData(res)
	if err != nil {
		return SeoData{}, err
	}
	return data, nil
}

func crawlPage(url string, tokens chan struct{}) (*http.Response, error) {
	tokens <- struct{}{}
	resp, err := makeRequest(url)
	<-tokens
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (d DefaultParser) GetSeoData(resp *http.Response) (SeoData, error) {
	// close the response body when we are done
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
	// close the response body when we are done
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	results := []string{}
	// find all <loc> tags in sitemap xml
	sel := doc.Find("loc")
	for i := range sel.Nodes {
		loc := sel.Eq(i)
		result := loc.Text()
		results = append(results, result)
	}
	return results, nil
}

func ScrapeSiteMap(url string, parser Parser, concurrency int) []SeoData {
	results := extractSiteMapUrls(url)
	//get structured data
	res := scrapeUrls(results, parser, concurrency)
	return res
}