package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB

type User struct {
	ID     string
	Name   string
	Email  string
	Salary int
}

func initDB() {
	var err error
	db, err = sql.Open(
		"postgres",
		"postgresql://neondb_owner:npg_R8TL4WNhnHav@ep-steep-wind-ahtciy2l-pooler.c-3.us-east-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require",
	)
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(
		"INSERT INTO users (id, name, email, salary) VALUES ($1, $2, $3, $4)",
		user.ID,
		user.Name,
		user.Email,
		user.Salary,
	)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func getUser(w http.ResponseWriter, r *http.Request, id string) {
	var user User
	err := db.QueryRow(
		"SELECT * FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Salary)

	if err == sql.ErrNoRows {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func updateUser(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Salary int    `json:"salary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	res, err := db.Exec(
		"UPDATE users SET name = $1, email = $2, salary = $3 WHERE id = $4",
		body.Name,
		body.Email,
		body.Salary,
		id,
	)

	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "no user found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func deleteUser(w http.ResponseWriter, r *http.Request, id string) {
	res, err := db.Exec(
		"DELETE FROM users WHERE id = $1",
		id,
	)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/users/")

	switch r.Method {
	case http.MethodGet:
		getUser(w, r, id)
	case http.MethodPut:
		updateUser(w, r, id)
	case http.MethodDelete:
		deleteUser(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	initDB()

	http.HandleFunc("/users", createUser)
	http.HandleFunc("/users/", userHandler)

	fmt.Println("server running on port 8080")
	http.ListenAndServe(":8080", nil)
}
