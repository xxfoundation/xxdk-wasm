package main

import (
	"fmt"
	"net/http"
)

func main() {
	port := "9090"
	root := "../assets"
	fmt.Printf("Starting server on port %s from %s\n", port, root)

	err := http.ListenAndServe(":"+port, http.FileServer(http.Dir(root)))
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
