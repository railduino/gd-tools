CREATE DATABASE IF NOT EXISTS rc_{{.Name}} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'rc_{{.Name}}'@'localhost' IDENTIFIED BY '{{.Password}}';
ALTER USER 'rc_{{.Name}}'@'localhost' IDENTIFIED BY '{{.Password}}';
GRANT SELECT, INSERT, UPDATE, DELETE ON rc_{{.Name}}.* TO 'rc_{{.Name}}'@'localhost';

GRANT SELECT, UPDATE ON vmail.* TO 'rc_{{.Name}}'@'localhost';
FLUSH PRIVILEGES;

