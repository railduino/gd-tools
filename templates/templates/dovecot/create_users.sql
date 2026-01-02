CREATE DATABASE IF NOT EXISTS vmail CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'vmail'@'localhost' IDENTIFIED BY '{{.Password}}';
ALTER USER 'vmail'@'localhost' IDENTIFIED BY '{{.Password}}';
GRANT SELECT, INSERT, UPDATE, DELETE ON vmail.* TO 'vmail'@'localhost';
FLUSH PRIVILEGES;

CREATE TABLE IF NOT EXISTS vmail.virtual_users (
    email VARCHAR(255) PRIMARY KEY,
    initial_password VARCHAR(255),
    user_password VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS vmail.last_login (
    email       VARCHAR(255) NOT NULL,
    service     VARCHAR(10)  NOT NULL,
    last_access BIGINT       NOT NULL,
    last_ip     VARCHAR(45)  NOT NULL,
    PRIMARY KEY (email, service)
);

