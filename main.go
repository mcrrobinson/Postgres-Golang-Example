package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Person struct contains the name and nickname of a person.
type Person struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
}

const (
	host     = "127.0.0.1"
	port     = 5555
	user     = "postgres"
	password = "password"
	dbname   = "postgres"
)

// OpenConnection starts the connection with the PostGres database.
func OpenConnection() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Attempt to connect to the database.
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("Error - couldn't connect to postgres")
	}

	// Ping the database to test the connection.
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Error - couldn't ping database")
	}

	return db, nil
}

// GETHandler opens the database connection and attempts to query.
func GETHandler(w http.ResponseWriter, r *http.Request) {
	db, err := OpenConnection()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	rows, err := db.Query("SELECT * FROM person")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var people []Person
	for rows.Next() {
		var person Person
		rows.Scan(&person.Name, &person.Nickname)
		people = append(people, person)
	}

	peopleBytes, _ := json.MarshalIndent(people, "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.Write(peopleBytes)

	defer rows.Close()
	defer db.Close()
}

// POSTHandler posts to the database.
func POSTHandler(w http.ResponseWriter, r *http.Request) {
	db, err := OpenConnection()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var p Person
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `INSERT INTO person (name, nickname) VALUES ($1, $2)`
	_, err = db.Exec(sqlStatement, p.Name, p.Nickname)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	defer db.Close()
}

func main() {
	http.HandleFunc("/", GETHandler)
	http.HandleFunc("/insert", POSTHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
