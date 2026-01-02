package agent

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	MySQL_Socket = "/run/mysqld/mysqld.sock"
	MySQL_DSN    = "root@unix(" + MySQL_Socket + ")/"
)

type MySQL struct {
	Stmts   []string `json:"stmts"`
	DbPath  string   `json:"db_path"`
	DbName  string   `json:"db_name"`
	Comment string   `json:"comment"`
}

// MySQLTest checks if there is work to be done
func MySQLTest(req *Request) bool {
	return req != nil && len(req.MySQLs) > 0
}

// MySQLHandler executes SQL statements sent in the Request
func MySQLHandler(req *Request, resp *Response) error {
	if req == nil || len(req.MySQLs) == 0 || resp == nil {
		return nil
	}

	db, err := sql.Open("mysql", MySQL_DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to MariaDB: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("MariaDB ping failed: %w", err)
	}

	for _, entry := range req.MySQLs {
		if len(entry.Stmts) > 0 {
			for _, line := range entry.Stmts {
				if err := ExecMySQL(db, line); err != nil {
					return err
				}
			}
			resp.Sayf("✅ SQL statement '%s'", entry.Comment)
			continue
		}
		if entry.DbName != "" && entry.DbPath != "" {
			if err := RunMySQL(resp, db, entry.DbName, entry.DbPath); err != nil {
				return err
			}
			continue
		}
		return fmt.Errorf("incomplete SQL entry '%s'", entry.Comment)
	}

	return nil
}

func ExecMySQL(db *sql.DB, stmt string) error {
	if ok := strings.HasSuffix(stmt, ";"); !ok {
		stmt += ";"
	}
	log.Printf("ExecMySQL: >%s<", stmt)

	if _, err := db.Exec(stmt); err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

func RunMySQL(resp *Response, db *sql.DB, dbName, dbPath string) error {
	query := `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ?`
	var count int

	if err := db.QueryRow(query, dbName).Scan(&count); err != nil {
		return fmt.Errorf("query table-count failed for %s: %w", dbName, err)
	}

	if count > 0 {
		resp.Sayf("✅ database %s initialized", dbName)
		return nil
	}
	resp.Sayf("running database script %s for %s", dbPath, dbName)

	cmd := exec.Command("mysql", dbName,
		"-S", MySQL_Socket,
		"-u", "root",
		"-e", fmt.Sprintf("source %s;", dbPath),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run mysql script: %w\noutput:\n%s", err, output)
	}

	return nil
}

func MakeDBName(name string) string {
	re := regexp.MustCompile(`[^a-z0-9]+`)

	name = strings.ToLower(name)
	name = re.ReplaceAllString(name, "_")

	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "db_" + name
	}

	if len(name) > 64 {
		name = name[:64]
	}

	return name
}
