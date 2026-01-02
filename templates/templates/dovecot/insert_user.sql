INSERT INTO virtual_users (email, initial_password)
    VALUES ('{{.Email}}', '{BLF-CRYPT}{{.Hash}}')
    ON CONFLICT(email) DO UPDATE SET initial_password = excluded.initial_password;

