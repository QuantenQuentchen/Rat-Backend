package main

import (
	events "RatBackend/Events"
	"RatBackend/auth"
	"RatBackend/db"
	"RatBackend/handler"
	"RatBackend/logic"
	"fmt"
	"log"
	"net/http"
)

// TODO: Timeout for Emergency Votes (needs to be clearer defined)
// TODO: Add actual audit Trail
// TODO: Save Audit Trail
// TODO: Dynamic Role Creation (define Role Perview, force Timeout, etc.) maybe in V2
// TODO: Possible Ewigkeitsklausel (Eternal Clause) for Logical topography (vote kinds, timeouts, etc.)
func main() {
	err := db.InitDB()
	if err != nil {
		return
	}
	go auth.CleanupJTI()
	logic.ScheduledTasks()

	http.HandleFunc("/admin/", handler.AdminHandler)
	http.HandleFunc("/vote/", handler.VoteHandler)
	http.HandleFunc("/ws", events.WsHandler)
	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
