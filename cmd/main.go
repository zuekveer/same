package main

import (
	"log"
	"os"

	"app/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Println("Server failed:", err)
		os.Exit(1)
	}
}
