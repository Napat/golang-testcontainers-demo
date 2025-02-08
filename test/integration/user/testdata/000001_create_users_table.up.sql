DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    version INT NOT NULL DEFAULT 1,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status)
);

INSERT INTO users (
    username,
    email,
    password,
    first_name,
    last_name,
    status,
    created_at,
    updated_at
) VALUES (
    'Napat',
    'napat@example.com',
    '$2a$10$6EwITJ7Zdo6b8pE6L9X8NOyQ1av7YOKxhHJoBWCqpGe9OHgGQYdce', -- hashed "password123"
    'Napat',
    'Rungruangbangchan',
    'active',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);
