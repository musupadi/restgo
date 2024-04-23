package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Response struct {
	Status bool   `json:"status"`
	Data   []Item `json:"data"`
}

func GetItems(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name FROM items")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if len(items) != 0 {
		response := Response{
			Status: true,
			Data:   items,
		}

		json.NewEncoder(w).Encode(response)
	} else {
		response := Response{
			Status: false,
			Data:   []Item{},
		}

		json.NewEncoder(w).Encode(response)
	}

}

func GetItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var item Item
	err := db.QueryRow("SELECT * FROM items WHERE id = ?", id).Scan(&item.ID, &item.Name)
	switch {
	case err == sql.ErrNoRows:
		response := Response{
			Status: false,
			Data:   []Item{},
		}
		json.NewEncoder(w).Encode(response)
		return
	case err != nil:
		http.Error(w, "Failed to get item", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	response := Response{
		Status: true,
		Data:   []Item{item},
	}
	json.NewEncoder(w).Encode(response)
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO items(name) VALUES(?)", name)
	if err != nil {
		http.Error(w, "Failed to insert item into database", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	item := Item{Name: name}
	json.NewEncoder(w).Encode(item)
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/test")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/items", GetItems).Methods("GET")
	router.HandleFunc("/items/{id}", GetItem).Methods("GET")
	router.HandleFunc("/items", CreateItem).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", router))
}
