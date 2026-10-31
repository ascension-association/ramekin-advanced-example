// todomvc serves a TodoMVC app: the page, a JSON API, and SQLite storage.
package main

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

//go:embed index.html
var page []byte

type todo struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func main() {
	addr := flag.String("addr", ":80", "listen address")
	dbPath := flag.String("db", "/data/todos.db", "SQLite database; keep it in /data so it survives updates")
	flag.Parse()

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY, title TEXT NOT NULL, completed INTEGER NOT NULL DEFAULT 0)`); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(page)
	})
	http.HandleFunc("GET /api/todos", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT id, title, completed FROM todos ORDER BY id`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		todos := []todo{}
		for rows.Next() {
			var t todo
			if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			todos = append(todos, t)
		}
		json.NewEncoder(w).Encode(todos)
	})
	http.HandleFunc("POST /api/todos", func(w http.ResponseWriter, r *http.Request) {
		var t todo
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil || t.Title == "" {
			http.Error(w, "a title is required", http.StatusBadRequest)
			return
		}
		res, err := db.Exec(`INSERT INTO todos (title) VALUES (?)`, t.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.ID, _ = res.LastInsertId()
		json.NewEncoder(w).Encode(t)
	})
	http.HandleFunc("PATCH /api/todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		var t todo
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exec(w, `UPDATE todos SET completed = ? WHERE id = ?`, db, t.Completed, r.PathValue("id"))
	})
	http.HandleFunc("DELETE /api/todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		exec(w, `DELETE FROM todos WHERE id = ?`, db, r.PathValue("id"))
	})
	http.HandleFunc("DELETE /api/todos", func(w http.ResponseWriter, r *http.Request) {
		exec(w, `DELETE FROM todos WHERE completed`, db) // "Clear completed"
	})

	log.Printf("todomvc on %s, data in %s", *addr, *dbPath)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func exec(w http.ResponseWriter, query string, db *sql.DB, args ...any) {
	if _, err := db.Exec(query, args...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
