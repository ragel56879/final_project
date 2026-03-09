package server

import (
	"final_project/pkg/api"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Serv(port string) {
	r := chi.NewRouter()

	r.Get("/api/task", api.TaskByIDHandler)
	r.Post("/api/task", api.AddTaskHandler)
	r.Put("/api/task", api.TaskByIDHandler)
	r.Delete("/api/task", api.TaskByIDHandler)
	r.Post("/api/task/done", api.TaskDoneHandler)
	r.Get("/api/tasks", api.TasksHandler)
	r.HandleFunc("/api/nextdate", api.NextDayHandler)
	r.Handle("/*", http.FileServer(http.Dir("web")))

	err := http.ListenAndServe(port, r)
	if err != nil {
		log.Fatal(err)
	}
}
