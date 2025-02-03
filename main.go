package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type URL struct {
	Id           int       `json:"id"`
	OriginalURL  string    `json:"original_url"`
	ShortURL     string    `json:"short_url"`
	CreationDate time.Time `json:"creation_date"`
}

var globalIDCounter = 0
var urlDB = make(map[string]URL)

//-------------------------------------------------------------------------//
// CREATE AND RETURN THE SHORT URL

func generateShortURL(OriginalURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(OriginalURL))
	data := hasher.Sum(nil)
	hash := hex.EncodeToString(data)
	return hash[:8]
}

func createURL(originalURL string) string {
	shortURL := generateShortURL(originalURL)
	id := shortURL
	globalIDCounter++
	urlDB[id] = URL{
		Id:           globalIDCounter,
		OriginalURL:  originalURL,
		ShortURL:     shortURL,
		CreationDate: time.Now(),
	}
	return shortURL
}

func ShortURLHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		URL string `json:"url"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shortURL := createURL(data.URL)

	response := struct {
		ShortURL string `json:"short_url"`
	}{
		ShortURL: "localhost:3000/redirect/" + shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

// -------------------------------------------------------------------------//
//
//	REDIRECT TO THE ORIGINAL URL
func getURL(id string) (URL, error) {
	url, ok := urlDB[id]
	if !ok {
		return URL{}, errors.New("URL NOT FOUND")
	}
	return url, nil
}

func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/redirect/"):]
	url, err := getURL(id)
	if err != nil {
		http.Error(w, "URL NOT FOUND!!", http.StatusNotFound)
	}
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

// -------------------------------------------------------------------------//
// ROOT PAGE

func RootPageURL(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("GET Method")
	fmt.Fprintf(w, "Hello! Welcome to the URL Shortner !!!")
}

// -------------------------------------------------------------------------//
// DB HANDLER PAGE

func DBHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("GET Method")
	for key, value := range urlDB {
		output := "key : " + key + " value : "
		fmt.Fprintf(w, output)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(value)
	}
}

// -------------------------------------------------------------------------//
// MAIN FUNCTION
func main() {
	fmt.Println("URL Shortner running....")

	// Register the handler function to handle all the requests to the root URL("/")
	http.HandleFunc("/", RootPageURL)
	http.HandleFunc("/shorten", ShortURLHandler)
	http.HandleFunc("/redirect/", redirectURLHandler)
	http.HandleFunc("/db", DBHandler)

	// starting a server on port 3000
	fmt.Println("Starting the server on port 3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println("Error while running the server:", err)
	}
}
