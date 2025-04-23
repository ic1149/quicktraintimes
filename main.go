package main

import (
	"http"
	"log"
)

func main() {
	req, err := http.NewRequest("GET", "http://10.0.0.1/", nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("x-apikey", "")
}
