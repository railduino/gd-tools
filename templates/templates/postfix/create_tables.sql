CREATE DATABASE IF NOT EXISTS vmail CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'vmail'@'localhost' IDENTIFIED BY '{{.Password}}';
ALTER USER 'vmail'@'localhost' IDENTIFIED BY '{{.Password}}';
GRANT SELECT, INSERT, UPDATE, DELETE ON vmail.* TO 'vmail'@'localhost';
FLUSH PRIVILEGES;

CREATE TABLE IF NOT EXISTS vmail.domain (
    domain VARCHAR(255) PRIMARY KEY,
    active BOOL DEFAULT 1
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS vmail.mailbox (
    email VARCHAR(255) PRIMARY KEY,
    domain VARCHAR(255),
    active BOOL DEFAULT 1,
    FOREIGN KEY (domain) REFERENCES vmail.domain(domain) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS vmail.alias (
    source VARCHAR(255) PRIMARY KEY,
    destination VARCHAR(255)
) ENGINE=InnoDB;

