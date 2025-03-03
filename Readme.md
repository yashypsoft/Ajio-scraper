
# Ajio-Scraper

Ajio-Scraper is a robust Go-based web scraper designed to extract product data from Ajio's API and store it in a MySQL database. The scraper supports concurrent processing, error handling, and periodic updates via Telegram notifications.

## Features

- **Concurrent Scraping**: Fetches data from multiple pages simultaneously using goroutines and a semaphore for concurrency control.
- **Retry Logic**: Automatically retries failed requests up to 3 times to handle transient errors.
- **Database Integration**: Stores scraped data in a MySQL database with support for batch inserts and history tracking.
- **Telegram Notifications**: Sends real-time updates about the scraping progress and completion status.
- **Graceful Shutdown**: Handles interruptions (e.g., Ctrl+C) gracefully by canceling ongoing operations.
- **Error Logging**: Logs failed pages and database errors for debugging and reprocessing.

## Table of Contents

- [Ajio-Scraper](#ajio-scraper)
  - [Features](#features)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
  - [Configuration](#configuration)
  - [Usage](#usage)
  - [Database Schema](#database-schema)
  - [Contributing](#contributing)
  - [License](#license)
  - [Acknowledgments](#acknowledgments)

## Installation

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/your-username/AjioScraper.git
   cd AjioScraper
   ```

2. **Install Dependencies**:
   Ensure you have Go installed (version 1.21 or higher). Then, install the required dependencies:
   ```bash
   go mod tidy
   ```

3. **Build the Scraper**:
   Compile the project into an executable binary:
   ```bash
   go build -o ajioscraper ./src
   ```

## Configuration

1. **Environment Variables**:
   Create a `.env` file in the root directory with the following variables:
   ```env
   TELEGRAM_BOT_TOKEN=your_telegram_bot_token
   TELEGRAM_CHAT_ID=your_telegram_chat_id
   MYSQL_USERNAME=your_mysql_username
   MYSQL_PASS=your_mysql_password
   MYSQL_HOST=your_mysql_host
   MYSQL_PORT=your_mysql_port
   MYSQL_DATABASE=your_mysql_database
   ```

2. **Database Setup**:
   Ensure your MySQL database has the following tables:
   - `ajio_products`: Stores product details.
   - `ajio_product_history`: Tracks price changes over time.

   Example schema:
   ```sql
   CREATE TABLE IF NOT EXISTS ajio_products (
   id INT AUTO_INCREMENT PRIMARY KEY,
   code VARCHAR(255), -- Ensure `code` is unique
   brandName VARCHAR(255),
   colorGroup VARCHAR(255),
   imageUrl TEXT,
   price DECIMAL(10, 2),
   wasPrice DECIMAL(10, 2),
   name VARCHAR(255),
   brandType VARCHAR(255),
   url VARCHAR(255) NOT NULL UNIQUE,
   offerPrice DECIMAL(10, 2),
   segment VARCHAR(255),
   vertical VARCHAR(255),
   brick VARCHAR(255),
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );


   CREATE TABLE IF NOT EXISTS ajio_product_history (
      id INT AUTO_INCREMENT PRIMARY KEY,
      url VARCHAR(255) NOT NULL, -- Product identifier
      price DECIMAL(10, 2), -- Current price
      wasPrice DECIMAL(10, 2), -- Original price
      offerPrice DECIMAL(10, 2), -- Offer price
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Date of record creation
      FOREIGN KEY (url) REFERENCES ajio_products(url) ON DELETE CASCADE
   );

   ```

## Usage

1. **Run the Scraper**:
   Start scraping products from page 0 to 23400:
   ```bash
   ./ajioscraper --start-page=0 --end-page=23400
   ```

2. **Custom Range**:
   Scrape a specific range of pages:
   ```bash
   ./ajioscraper --start-page=1000 --end-page=2000
   ```

3. **Monitor Progress**:
   Real-time updates are sent to your Telegram chat during the scraping process.

## Database Schema

The scraper interacts with two tables:

1. **`ajio_products`**:
   - Stores detailed information about each product.
   - Fields: `code`, `brandName`, `colorGroup`, `imageUrl`, `price`, `wasPrice`, `name`, `brandType`, `url`, `offerPrice`, `segment`, `vertical`, `brick`.

2. **`ajio_product_history`**:
   - Tracks historical price changes for each product.
   - Fields: `id`, `code`, `price`, `wasPrice`, `offerPrice`, `timestamp`.

## Contributing

We welcome contributions! To contribute:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature`).
3. Commit your changes (`git commit -m "Add your feature"`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Open a pull request.

Please ensure your code adheres to Go best practices and includes appropriate tests.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- Built using Go's concurrency model for efficient scraping.
- Inspired by open-source scraping projects and community contributions.

For questions or feedback, feel free to open an issue or contact the maintainer.
