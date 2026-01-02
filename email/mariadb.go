package email

import (
	"database/sql"
	"fmt"
)

const (
	MySQL_Socket = "/run/mysqld/mysqld.sock"
	MySQL_DSN    = "root@unix(" + MySQL_Socket + ")/"
)

func OpenDB(dbName string) (*sql.DB, error) {
	db, err := sql.Open("mysql", MySQL_DSN+dbName)
	if err != nil {
		return nil, fmt.Errorf("open mysql (%s): %w", dbName, err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("MariaDB ping failed: %w", err)
	}

	return db, nil
}

// InsertUserMySQL inserts a new user and all related records (domain, mailbox, aliases).
// If the user already exists, it will be replaced.
func InsertUserMySQL(user *User) error {
	db, err := OpenDB("")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction to insert %s: %w", user.Email(), err)
	}
	defer tx.Rollback()

	// Ensure domain exists
	if _, err := tx.Exec(`INSERT IGNORE INTO postfix.virtual_domains (name) VALUES (?)`, user.Domain); err != nil {
		return fmt.Errorf("failed to insert postfix domain %s: %w", user.Domain, err)
	}

	// Mailbox
	if _, err := tx.Exec(`INSERT INTO postfix.virtual_mailboxes (email) VALUES (?) ON DUPLICATE KEY UPDATE email=email`, user.Email()); err != nil {
		return fmt.Errorf("failed to insert postfix mailbox %s: %w", user.Email(), err)
	}

	// Aliases
	for _, alias := range user.Aliases {
		if _, err := tx.Exec(`REPLACE INTO postfix.virtual_aliases (source, destination) VALUES (?, ?)`, alias, user.Email()); err != nil {
			return fmt.Errorf("failed to insert postfix alias %s: %w", alias, err)
		}
	}

	// User credentials
	if _, err := tx.Exec(`REPLACE INTO dovecot.virtual_users (email, initial_password) VALUES (?, ?)`, user.Email(), user.Password); err != nil {
		return fmt.Errorf("failed to insert dovecot user %s: %w", user.Email(), err)
	}

	return tx.Commit()
}

// UpdateUserMySQL updates the user's initial password and alias list.
func UpdateUserMySQL(user *User) error {
	db, err := OpenDB("")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction to update %s: %w", user.Email(), err)
	}
	defer tx.Rollback()

	// Update password
	if _, err := tx.Exec(`UPDATE dovecot.virtual_users SET initial_password = ? WHERE email = ?`, user.Password, user.Email()); err != nil {
		return fmt.Errorf("failed to update password for %s: %w", user.Email(), err)
	}

	// Clear existing aliases
	if _, err := tx.Exec(`DELETE FROM postfix.virtual_aliases WHERE destination = ?`, user.Email()); err != nil {
		return fmt.Errorf("failed to delete aliases for %s: %w", user.Email(), err)
	}

	// Insert new aliases
	for _, alias := range user.Aliases {
		if _, err := tx.Exec(`INSERT INTO postfix.virtual_aliases (source, destination) VALUES (?, ?)`, alias, user.Email()); err != nil {
			return fmt.Errorf("failed to insert alias for %s: %w", user.Email(), err)
		}
	}

	return tx.Commit()
}

// DeleteUserMySQL removes the user, mailbox and aliases, but leaves the domain intact.
func DeleteUserMySQL(user *User) error {
	db, err := OpenDB("")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction to delete %s: %w", user.Email(), err)
	}
	defer tx.Rollback()

	// Delete user credentials
	if _, err := tx.Exec(`DELETE FROM dovecot.virtual_users WHERE email = ?`, user.Email()); err != nil {
		return fmt.Errorf("failed to delete user %s: %w", user.Email(), err)
	}

	// Delete aliases
	if _, err := tx.Exec(`DELETE FROM postfix.virtual_aliases WHERE destination = ? OR source = ?`, user.Email(), user.Email()); err != nil {
		return fmt.Errorf("failed to delete aliases for %s: %w", user.Email(), err)
	}

	// Delete mailbox
	if _, err := tx.Exec(`DELETE FROM postfix.virtual_mailboxes WHERE email = ?`, user.Email()); err != nil {
		return fmt.Errorf("failed to delete mailbox %s: %w", user.Email(), err)
	}

	return tx.Commit()
}
