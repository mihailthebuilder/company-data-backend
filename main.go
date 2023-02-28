package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
)

func main() {
	http.HandleFunc("/companies/sic_code/", handler)
	log.Println("Starting server on localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	sic := r.URL.Path[len("/companies/sic_code/"):]

	log.Printf("%s %s", r.Method, "/companies/sic_code/")

	valid, err := isValidSicFormat(sic)
	if err != nil {
		log.Println("Valid SIC format check failed:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	if !valid {
		http.Error(w, fmt.Sprintf("Invalid SIC code: %s", sic), http.StatusBadRequest)
	}

	fmt.Fprintf(w, "SIC Code: %s", sic)
}

func isValidSicFormat(sic string) (bool, error) {
	pattern := "^[0-9]+$"
	match, err := regexp.MatchString(pattern, sic)
	return match, err
}
