package main

import (
	"final_project/pkg/config"
	"final_project/pkg/db"
	"final_project/pkg/server"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	err := db.Init(cfg.TODO_DBFILE)
	if err != nil {
		log.Print(err)
	}
	defer db.DB.Close()

	log.Printf("Сервер запущен на порту: %s", cfg.TODO_PORT)
	server.Serv(":" + cfg.TODO_PORT)
}
