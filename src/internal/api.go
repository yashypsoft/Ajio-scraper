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

func FetchPages(ctx context.Context, wg *sync.WaitGroup, results chan<- Product, failedPages chan<- int, telegramBot *TelegramBot, startPage, endPage int) {
	concurrencyLimit := 1000
	semaphore := make(chan struct{}, concurrencyLimit)

	// Launch goroutines for each page
	for i := startPage; i <= endPage; i++ {
		select {
		case <-ctx.Done():
			close(results)
			close(failedPages)
			return
		default:
			wg.Add(1)
			semaphore <- struct{}{}
			go func(page int) {
				defer wg.Done()
				defer func() { <-semaphore }()

				// Retry logic: Try fetching the page up to 3 times
				maxRetries := 3
				var data []Product
				var err error
				for attempt := 1; attempt <= maxRetries; attempt++ {
					log.Printf("Fetching page %d (Attempt %d/%d)", page, attempt, maxRetries)
					data, err = getAjioData(page)
					if err == nil {
						break // Success, exit retry loop
					}
					log.Printf("Failed to fetch page %d (Attempt %d/%d): %v", page, attempt, maxRetries, err)
					if attempt == maxRetries {
						// Mark the page as failed after all retries
						log.Printf("Marking page %d as failed after %d attempts", page, maxRetries)
						failedPages <- page
						return
					}
				}

				log.Printf("Processing %d products from page %d", len(data), page)
				for _, product := range data {
					results <- product
				}
			}(i)
		}
	}

	// Wait for all goroutines to finish before closing channels
	go func() {
		wg.Wait()
		close(results)
		close(failedPages)
	}()
}
func getAjioData(page int) ([]Product, error) {
	url := fmt.Sprintf("https://www.ajio.com/api/category/83?fields=SITE&currentPage=%d&pageSize=99&format=json&query=%%3Arelevance&gridColumns=3&advfilter=true&platform=Desktop&showAdsOnNextPage=false&is_ads_enable_plp=true&displayRatings=true&segmentIds=", page)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set User-Agent header
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Send the request using a client
	client := &http.Client{}
	resp, err := client.Do(req)
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

	// Check if "products" key exists and is not nil
	productsRaw, ok := result["products"]
	if !ok || productsRaw == nil {
		log.Printf("No products found for page %d", page)
		return nil, nil // Return an empty slice if no products are found
	}

	// Ensure "products" is a slice of interfaces
	products, ok := productsRaw.([]interface{})
	if !ok {
		log.Printf("Unexpected type for 'products' key on page %d: %T", page, productsRaw)
		return nil, fmt.Errorf("unexpected type for 'products' key")
	}

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
