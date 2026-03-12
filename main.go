package main

import (
	"final_project/pkg/db"
	"final_project/pkg/server"
	"log"
)

func main() {
	// godotenv.Load()
	// cfg := config.Load()
	port := "7540"
	dbFile := "scheduler.db"

	err := db.Init(dbFile)
	if err != nil {
		log.Print(err)
	}
	defer db.DB.Close()

	log.Printf("Сервер запущен на порту: %s", port)
	server.Serv(":" + port)
}
