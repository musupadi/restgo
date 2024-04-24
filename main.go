package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

//Variable

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Response struct {
	Status  bool   `json:"status"`
	Data    []Item `json:"data"`
	Message string `json:"message"`
}
type Response2 struct {
	Status  bool   `json:"status"`
	Data    []User `json:"data"`
	Message string `json:"message"`
}

type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	Department  string `json:"department"`
	PhoneNumber string `json:"phone_number"`
	IDRole      string `json:"id_role"`
	Photo       string `json:"photo"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
	UpdatedAt   string `json:"updated_at"`
	UpdatedBy   string `json:"updated_by"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	// Hash the password using MD5
	hashedPassword := md5.Sum([]byte(password))
	hashedPasswordStr := hex.EncodeToString(hashedPassword[:])

	var user User
	err := db.QueryRow("SELECT * FROM m_user WHERE username = ? AND password = ?", username, hashedPasswordStr).Scan(
		&user.ID, &user.Name, &user.Username, &user.Password, &user.Email, &user.Department,
		&user.PhoneNumber, &user.IDRole, &user.Photo, &user.CreatedAt,
		&user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
	)
	switch {
	case err == sql.ErrNoRows:
		response := Response2{
			Status:  false,
			Data:    []User{},
			Message: hashedPasswordStr,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Fatal(err)
		}
		return
	case err != nil:
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	response := Response2{
		Status:  true,
		Data:    []User{user},
		Message: "Successfully logged in",
	}

	// Marshal the response data with indentation
	jsonResponse, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response to the response writer
	if _, err := w.Write(jsonResponse); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
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
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/db_inventory")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/items", GetItems).Methods("GET")
	router.HandleFunc("/items/{id}", GetItem).Methods("GET")
	router.HandleFunc("/items", CreateItem).Methods("POST")
	router.HandleFunc("/login", Login).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", router))
}
