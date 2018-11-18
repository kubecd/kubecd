package main

import (
	"fmt"
	"net/http"
)

const Version = "1.0"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		message := "Hello, I am demo-app version " + Version
		w.Write([]byte(message))
	})
	fmt.Printf("Starting demo-app version %s\n", Version)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
