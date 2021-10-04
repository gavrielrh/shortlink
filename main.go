package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func handler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	short := r.URL.Path[1:]
	switch r.Method {
	case "GET":
		if len(short) <= 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "not found")
			return
		}
		long := getLink(db, short)
		if len(long) <= 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "not found")
			return
		}
		log.Printf("%s -> %s", short, long)
		http.Redirect(w, r, long, http.StatusFound)
	case "POST":
		long := r.URL.Query().Get("long")
		if long != "" {
			insertLink(db, short, long)
			fmt.Fprintf(w, "New: %s -> %s", short, long)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func initializeLinksTable(db *sql.DB) {
	log.Println("Initializing links table...")
	sql := `
	CREATE TABLE IF NOT EXISTS links (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"short" TEXT,
		"long" TEXT
	);
	`
	stmt, err := db.Prepare(sql)
	if err != nil {
		log.Fatal(err)
	}
	stmt.Exec()
	log.Println("Links table ready")
}

func insertLink(db *sql.DB, short string, long string) {
	query := `INSERT INTO links (short, long) VALUES (?, ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(short, long)
	if err != nil {
		log.Fatal(err)
	}
}

func getLink(db *sql.DB, short string) string {
	var long string
	query := `SELECT long FROM links WHERE short = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	row := stmt.QueryRow(short)
	switch err := row.Scan(&long); err {
	case sql.ErrNoRows:
		return ""
	default:
		return long
	}
}

func main() {
	dbPath := "./links.db"
	if _, err := os.Stat(dbPath); err != nil {
		file, err := os.Create(dbPath)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	initializeLinksTable(db)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, db)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
