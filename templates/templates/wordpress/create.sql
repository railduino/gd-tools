CREATE DATABASE IF NOT EXISTS `wp_{{.Name}}` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

CREATE USER IF NOT EXISTS 'wp_{{.Name}}'@'localhost' IDENTIFIED BY '{{.Password}}';
ALTER USER 'wp_{{.Name}}'@'localhost' IDENTIFIED BY '{{.Password}}';
GRANT ALL PRIVILEGES ON `wp_{{.Name}}`.* TO 'wp_{{.Name}}'@'localhost';
FLUSH PRIVILEGES;

