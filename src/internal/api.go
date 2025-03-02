package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Product struct {
	Code       string  `json:"code"`
	BrandName  string  `json:"brandName"`
	ColorGroup string  `json:"colorGroup"`
	ImageURL   string  `json:"imageUrl"`
	Price      float64 `json:"price"`
	WasPrice   float64 `json:"wasPrice"`
	Name       string  `json:"name"`
	BrandType  string  `json:"brandType"`
	URL        string  `json:"url"`
	OfferPrice float64 `json:"offerPrice"`
	Segment    string  `json:"segment"`
	Vertical   string  `json:"vertical"`
	Brick      string  `json:"brick"`
}

func FetchPages(ctx context.Context, wg *sync.WaitGroup, results chan<- Product, failedPages chan<- int, telegramBot *TelegramBot) {
	startPage := 17000
	totalPages := 23400
	concurrencyLimit := 1000
	semaphore := make(chan struct{}, concurrencyLimit)

	// Track the total number of pages processed
	var pageCount int
	const batchSize = 1000 // Send a Telegram message every 1000 pages

	// Launch goroutines for each page
	for i := startPage; i <= totalPages; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			wg.Add(1)
			semaphore <- struct{}{}
			go func(page int) {
				defer wg.Done()
				defer func() { <-semaphore }()

				// log.Printf("Fetching page %d", page)
				data, err := getAjioData(page)
				if err != nil {
					log.Printf("Failed to fetch page %d: %v", page, err)
					failedPages <- page
					return
				}

				log.Printf("Processing %d products from page %d", len(data), page)
				for _, product := range data {
					results <- product
				}

				// Increment the page count
				pageCount++

				// Send a Telegram message when 1000 pages are processed
				if pageCount%batchSize == 0 {
					go func(count int) {
						message := fmt.Sprintf("Processed %d pages", count)
						err := telegramBot.SendMessage(message)
						if err != nil {
							log.Printf("Failed to send Telegram message: %v", err)
						}
					}(pageCount)
				}
			}(i)
		}
	}

	// Wait for all goroutines to finish before closing channels
	go func() {
		wg.Wait()
		close(results)
		close(failedPages)

		// Send a final Telegram message with the total page count
		go func(count int) {
			message := fmt.Sprintf("Scraping completed. Total pages processed: %d", count)
			err := telegramBot.SendMessage(message)
			if err != nil {
				log.Printf("Failed to send final Telegram message: %v", err)
			}
		}(pageCount)
	}()
}
func getAjioData(page int) ([]Product, error) {
	url := fmt.Sprintf("https://www.ajio.com/api/category/83?fields=SITE&currentPage=%d&pageSize=100&format=json&query=:newn&sortBy=newn", page)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	products := result["products"].([]interface{})
	var productList []Product
	for _, p := range products {
		product := parseProduct(p.(map[string]interface{}))
		productList = append(productList, product)
	}
	return productList, nil
}

func parseProduct(data map[string]interface{}) Product {
	return Product{
		Code:       getString(data, "code"),
		BrandName:  getString(data, "fnlColorVariantData.brandName"),
		ColorGroup: getString(data, "fnlColorVariantData.colorGroup"),
		ImageURL:   getString(data, "images.0.url"),
		Price:      getFloat(data, "price.value"),
		WasPrice:   getFloat(data, "wasPriceData.value"),
		Name:       getString(data, "name"),
		BrandType:  getString(data, "brandTypeName"),
		URL:        getString(data, "url"),
		OfferPrice: getFloat(data, "offerPrice.value"),
		Segment:    getString(data, "segmentNameText"),
		Vertical:   getString(data, "verticalNameText"),
		Brick:      getString(data, "brickNameText"),
	}
}

func getString(data map[string]interface{}, key string) string {
	keys := strings.Split(key, ".")
	var value interface{} = data

	for _, k := range keys {
		switch v := value.(type) {
		case map[string]interface{}:
			value = v[k] // Access map key
		case []interface{}:
			index, err := strconv.Atoi(k) // Convert key to integer index
			if err != nil || index < 0 || index >= len(v) {
				return "" // Invalid index
			}
			value = v[index] // Access array element
		default:
			return "" // Invalid access
		}
	}

	// Ensure the final value is a string
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

func getFloat(data map[string]interface{}, key string) float64 {
	keys := strings.Split(key, ".")
	var value interface{} = data
	for _, k := range keys {
		if m, ok := value.(map[string]interface{}); ok {
			value = m[k]
		} else {
			return 0
		}
	}
	if num, ok := value.(float64); ok {
		return num
	}
	return 0
}
