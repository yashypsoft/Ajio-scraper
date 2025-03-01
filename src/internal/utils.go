package internal

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

func ProcessResults(ctx context.Context, dbClient *MySQLClient, telegramBot *TelegramBot, results <-chan Product, failedPages <-chan int, wg *sync.WaitGroup) {
	var processed int
	var failed int
	buffer := make([]Product, 0, 1000)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Context canceled, exit the loop
			return
		case product, ok := <-results:
			if !ok {
				// Channel closed, disable this case
				results = nil
				continue
			}
			processed++
			log.Printf("Processed product: %d", processed)
			// Add product to buffer for batch insertion into ajio_products
			buffer = append(buffer, product)

			// Insert batch if buffer is full
			if len(buffer) >= 1000 {
				err := dbClient.InsertProductAndHistory(buffer)
				if err != nil {
					log.Printf("Error inserting batch: %v", err)
				}
				buffer = buffer[:0] // Clear the buffer
			}
		case page, ok := <-failedPages:
			if !ok {
				// Channel closed, disable this case
				failedPages = nil
				continue
			}
			failed++
			log.Printf("Failed page: %d", page)
		case <-ticker.C:
			if processed%1000 == 0 { // Send updates every 1000 products
				telegramBot.SendMessage(fmt.Sprintf("Processed: %d, Failed: %d", processed, failed))
			}
		}

		// Exit the loop if both channels are closed
		if results == nil && failedPages == nil {
			break
		}
	}

	// Insert any remaining products in the buffer
	if len(buffer) > 0 {
		err := dbClient.InsertProductAndHistory(buffer)
		if err != nil {
			log.Printf("Error inserting final batch: %v", err)
		}
	}
}
