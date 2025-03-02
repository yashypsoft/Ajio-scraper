package main

import (
	"ajio-scraper/src/internal"
	"context"
	"flag"
	"log"
	"os"
	"sync"
)

func main() {
	log.Println("Starting Ajio Scraper...")

	// Load environment variables
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }

	// Parse command-line arguments
	startPage := flag.Int("start-page", 0, "Start page number")
	endPage := flag.Int("end-page", 23400, "End page number")
	flag.Parse()

	log.Printf("Scraping pages from %d to %d...", *startPage, *endPage)

	// Load required environment variables
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

	// Initialize components
	dbClient := internal.NewMySQLClient()

	telegramBot := internal.NewTelegramBot(telegramToken, telegramChatID)
	telegramBot.SendMessage("Started Fetching Record")

	// Start scraping
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan internal.Product, 1000)
	failedPages := make(chan int, *endPage-*startPage+1)

	go internal.FetchPages(ctx, &wg, results, failedPages, telegramBot, *startPage, *endPage)
	internal.ProcessResults(ctx, dbClient, telegramBot, results, failedPages, &wg)

	wg.Wait()
	telegramBot.SendMessage("Completed Fetching Record")
	log.Println("Scraping completed successfully.")
}
