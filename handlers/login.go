package handlers

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
)

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

type Response2 struct {
	Status  bool   `json:"status"`
	Data    []User `json:"data"`
	Message string `json:"message"`
}

func Login(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
