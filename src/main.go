package main

import (
	"ajio-scraper/src/internal"
	"context"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting Ajio Scraper...")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Load environment variables
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
	failedPages := make(chan int, 22550)

	go internal.FetchPages(ctx, &wg, results, failedPages)
	internal.ProcessResults(ctx, dbClient, telegramBot, results, failedPages, &wg)

	wg.Wait()
	telegramBot.SendMessage("Completed Fetching Record")
	log.Println("Scraping completed successfully.")
}
