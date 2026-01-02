CREATE DATABASE IF NOT EXISTS `wbce_{{.Name}}` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

CREATE USER IF NOT EXISTS 'wbce_{{.Name}}'@'localhost' IDENTIFIED BY '{{.Password}}';
ALTER USER 'wbce_{{.Name}}'@'localhost' IDENTIFIED BY '{{.Password}}';
GRANT ALL ON `wbce_{{.Name}}`.* TO 'wbce_{{.Name}}'@'localhost';
FLUSH PRIVILEGES;

