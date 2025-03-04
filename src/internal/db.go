package internal

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLClient struct {
	DB *sql.DB
}

func NewMySQLClient() *MySQLClient {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("MYSQL_USERNAME"),
		os.Getenv("MYSQL_PASS"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	return &MySQLClient{DB: db}
}

func (c *MySQLClient) InsertProductAndHistory(products []Product) error {
	// Start a transaction
	tx, err := c.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback in case of an error

	// Prepare the bulk insert query for ajio_products
	insertProductQuery := `
        INSERT INTO ajio_products (
            code, brandName, colorGroup, imageUrl, price, wasPrice, name, brandType, url, offerPrice, segment, vertical, brick
        ) VALUES %s
        ON DUPLICATE KEY UPDATE
            brandName = VALUES(brandName),
            colorGroup = VALUES(colorGroup),
            imageUrl = VALUES(imageUrl),
            price = VALUES(price),
            wasPrice = VALUES(wasPrice),
            name = VALUES(name),
            brandType = VALUES(brandType),
            code = VALUES(code),
            offerPrice = VALUES(offerPrice),
            segment = VALUES(segment),
            vertical = VALUES(vertical),
            brick = VALUES(brick)
    `

	// Build the bulk values for ajio_products
	productValues := ""
	productArgs := []interface{}{}
	for i, product := range products {
		if i > 0 {
			productValues += ","
		}
		productValues += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		productArgs = append(productArgs,
			product.Code, product.BrandName, product.ColorGroup, product.ImageURL,
			product.Price, product.WasPrice, product.Name, product.BrandType,
			product.URL, product.OfferPrice, product.Segment, product.Vertical, product.Brick,
		)
	}

	// Execute the bulk insert for ajio_products
	finalProductQuery := fmt.Sprintf(insertProductQuery, productValues)
	_, err = tx.Exec(finalProductQuery, productArgs...)
	if err != nil {
		log.Printf("Error inserting products into ajio_products: %v", err)
		return err
	}

	// Prepare the bulk insert query for ajio_product_history
	insertHistoryQuery := `
        INSERT INTO ajio_product_history (url, price, wasPrice, offerPrice)
        VALUES %s
    `

	// Build the bulk values for ajio_product_history
	historyValues := ""
	historyArgs := []interface{}{}
	for i, product := range products {
		if i > 0 {
			historyValues += ","
		}
		historyValues += "(?, ?, ?, ?)"
		historyArgs = append(historyArgs, product.URL, product.Price, product.WasPrice, product.OfferPrice)
	}

	// Execute the bulk insert for ajio_product_history
	finalHistoryQuery := fmt.Sprintf(insertHistoryQuery, historyValues)
	_, err = tx.Exec(finalHistoryQuery, historyArgs...)
	if err != nil {
		log.Printf("Error inserting products into ajio_product_history: %v", err)
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
