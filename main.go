package main

import (
	events "RatBackend/Events"
	"RatBackend/auth"
	"RatBackend/db"
	"RatBackend/handler"
	"fmt"
	"log"
	"net/http"
)

func main() {
	db.InitDB()
	go auth.CleanupJTI()

	http.HandleFunc("/admin/", handler.AdminHandler)
	http.HandleFunc("/vote/", handler.VoteHandler)
	http.HandleFunc("/ws", events.WsHandler)
	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
