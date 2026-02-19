package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	handler "github.com/kawai-org/kawai-backend/api"
)

func main() {
	// 1. Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// 2. Setup Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// 3. Register Handler
	http.HandleFunc("/", handler.Handler)

	fmt.Printf("Server running at http://localhost:%s\n", port)
	fmt.Printf("Swagger UI available at http://localhost:%s/swagger/index.html\n", port)

	// 4. Start Server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
