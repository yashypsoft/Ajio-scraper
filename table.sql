CREATE TABLE IF NOT EXISTS ajio_products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE, -- Ensure `code` is unique
    brandName VARCHAR(255),
    colorGroup VARCHAR(255),
    imageUrl TEXT,
    price DECIMAL(10, 2),
    wasPrice DECIMAL(10, 2),
    name VARCHAR(255),
    brandType VARCHAR(255),
    url TEXT,
    offerPrice DECIMAL(10, 2),
    segment VARCHAR(255),
    vertical VARCHAR(255),
    brick VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS ajio_product_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(255) NOT NULL, -- Product identifier
    price DECIMAL(10, 2), -- Current price
    wasPrice DECIMAL(10, 2), -- Original price
    offerPrice DECIMAL(10, 2), -- Offer price
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Date of record creation
    FOREIGN KEY (code) REFERENCES ajio_products(code) ON DELETE CASCADE
);