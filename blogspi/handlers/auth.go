package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"blogspi/utils"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("my_secret_key")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	db := utils.ConnectDB()
	defer db.Close()

	var storedCreds Credentials
	var role string

	err = db.QueryRow("SELECT password, role FROM users WHERE username=$1", creds.Username).Scan(&storedCreds.Password, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Unauthorized: Invalid Credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Unauthorized: Invalid Credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(10000000 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Internal Server Error: Token Creation Failed", http.StatusInternalServerError)
		return
	}

	jsonResponse := map[string]string{"message": "Login successful", "status": "success", "token": tokenString}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonResponse)
}

func Register(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		http.Error(w, "Internal Server Error: Unable to Hash Password", http.StatusInternalServerError)
		return
	}

	db := utils.ConnectDB()
	defer db.Close()

	_, err = db.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", creds.Username, hashedPassword, creds.Role)
	if err != nil {
		log.Println("Error inserting user into database:", err)
		http.Error(w, "Internal Server Error: Unable to Register User", http.StatusInternalServerError)
		return
	}

	jsonResponse := map[string]string{"message": "Account registered successfully", "status": "success"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonResponse)
}
