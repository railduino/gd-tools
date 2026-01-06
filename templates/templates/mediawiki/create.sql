CREATE DATABASE IF NOT EXISTS `mw_{{.Name}}` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'mw_{{.Name}}'@'localhost' IDENTIFIED BY '{{.Password}}';
ALTER USER 'mw_{{.Name}}'@'localhost' IDENTIFIED BY '{{.Password}}';
GRANT ALL PRIVILEGES ON `mw_{{.Name}}`.* TO 'mw_{{.Name}}'@'localhost';
FLUSH PRIVILEGES;

