CREATE DATABASE IF NOT EXISTS `nc_{{.Name}}` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

CREATE USER IF NOT EXISTS 'nc_{{.Name}}'@'localhost' IDENTIFIED BY '{{.Password}}';
ALTER USER 'nc_{{.Name}}'@'localhost' IDENTIFIED BY '{{.Password}}';
GRANT ALL ON `nc_{{.Name}}`.* TO 'nc_{{.Name}}'@'localhost';
FLUSH PRIVILEGES;

