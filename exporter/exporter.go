package exporter

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"github.com/RudraPratapDev/SiteMapr/scraper"
)

// ExportJSON saves the results as JSON
func ExportJSON(data []scraper.SeoData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// ExportCSV saves the results as CSV
func ExportCSV(data []scraper.SeoData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"URL", "Title", "H1", "MetaDescription", "StatusCode"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, d := range data {
		record := []string{
			d.URL,
			d.Title,
			d.H1,
			d.MetaDescription,
			strconv.Itoa(d.StatusCode),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}
