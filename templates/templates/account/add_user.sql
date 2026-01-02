INSERT IGNORE INTO vmail.domain (domain) VALUES ('{{.Domain}}');
INSERT INTO vmail.mailbox (email, domain) VALUES ('{{.Email}}', '{{.Domain}}') ON DUPLICATE KEY UPDATE email=email;
{{range .Aliases}}
INSERT INTO vmail.alias (source, destination)
  VALUES ('{{.}}', '{{$.Email}}')
  ON DUPLICATE KEY UPDATE destination = VALUES(destination);
{{end}}

INSERT INTO vmail.virtual_users (email, initial_password)
  VALUES ('{{.Email}}', '{BLF-CRYPT}{{.Password}}')
  ON DUPLICATE KEY UPDATE initial_password = VALUES(initial_password);

