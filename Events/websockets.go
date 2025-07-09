package events

import (
	"RatBackend/auth"
	"RatBackend/db"
	"RatBackend/models"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type EventType int

const (
	VoteUpdate EventType = iota
	VoteCreation
	VoteStateChange
	RoleUpdate
)

type RoleUpdateType int

const (
	RoleAdded RoleUpdateType = iota
	RoleRemoved
)

// Thread-safe set of connected clients
var (
	clients   = make(map[*websocket.Conn]struct{})
	clientsMu sync.Mutex
)

// Proper upgrader with correct function signature
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Optionally add stricter checks here
		return true
	},
}

// Broadcast a message to all connected clients
func broadcastEvent(message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for conn := range clients {
		// WriteMessage is already safe for concurrent use per connection
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			conn.Close()
			delete(clients, conn)
		}
	}
}

// WebSocket handler (REST endpoint to upgrade connection)
func WsHandler(w http.ResponseWriter, r *http.Request) {
	// Optional: Authenticate here using your authenticate(r) function
	authStruct, err := auth.Authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade to websocket", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// Register client
	clientsMu.Lock()
	clients[conn] = struct{}{}
	clientsMu.Unlock()

	// Optionally send a welcome message
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Welcome, %s!", authStruct.Subject)))

	// Listen for client messages (optional, can be omitted if server only pushes)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	// Unregister client on disconnect
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
}

type EventMessage struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

type RoleUpdateMessage struct {
	RoleID      string                  `json:"role_id"`
	UserID      string                  `json:"user_id"`
	UpdateType  RoleUpdateType          `json:"update_type"`
	Reason      models.RoleChangeReason `json:"reason,omitempty"`
	IsCascading bool                    `json:"is_cascading,omitempty"`
	Details     interface{}             `json:"details,omitempty"`
}

type RemovedByID struct {
	RemovedByID string `json:"removed_by_id"`
}

type VoteCreationMessage struct {
	VoteID uint64 `json:"vote_id"`
}

type VoteStateChangeMessage struct {
	VoteID    uint64      `json:"vote_id"`
	Specifics interface{} `json:"specifics"`
}

type VetoDetails struct {
	VetoedBy string `json:"vetoed_by"`
	Reason   string `json:"reason"`
}

type NewVoteDetails struct {
	NewVoteID uint64 `json:"new_vote"`
}

type SimpleVoteDetails struct {
	Suceeded bool `json:"suceeded"`
}

type OngoingDetails struct {
	IsOngoing bool `json:"is_ongoing"`
}

type VoteUpdatePayload struct {
	VoteID uint64      `json:"vote_id"`
	Votes  interface{} `json:"payload"`
}

type VoteUpdateAnonymous struct {
	For     uint64 `json:"for_num"`
	Against uint64 `json:"against_num"`
	Abstain uint64 `json:"abstain_num"`
}

type VoteUpdatePersonal struct {
	For     []string `json:"for_arr"`
	Against []string `json:"against_arr"`
	Abstain []string `json:"abstain_arr"`
}

func BroadcastRoleUpdate(roleID string, userID string, updateType RoleUpdateType, ctx *models.RoleChangeContext) {
	var details interface{}
	if ctx.RemovedByID != "" {
		details = RemovedByID{
			RemovedByID: ctx.RemovedByID,
		}
	}
	message := EventMessage{
		Type: RoleUpdate,
		Payload: RoleUpdateMessage{
			RoleID:      roleID,
			UserID:      userID,
			UpdateType:  updateType,
			Reason:      ctx.Reason,
			IsCascading: ctx.IsCascading,
			Details:     details,
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Failed to marshal event:", err)
		return
	}
	broadcastEvent(string(data))
}

func BroadcastVoteStateChange(vote_id uint64, state models.VoteState, ctx *models.VoteStateChangeContext) {
	var specifics interface{}

	switch state {
	case models.FailedVeto:
		specifics = VetoDetails{
			VetoedBy: ctx.VetoedBy,
			Reason:   ctx.Reason,
		}
	case models.FailedInconclusive:
		specifics = NewVoteDetails{
			NewVoteID: ctx.NewVoteID,
		}
	case models.Succeeded:
		specifics = SimpleVoteDetails{
			Suceeded: true,
		}
	case models.FailedVotes:
		specifics = SimpleVoteDetails{
			Suceeded: false,
		}
	case models.Ongoing:
		specifics = OngoingDetails{
			IsOngoing: true,
		}
	}

	message := EventMessage{
		Type: VoteStateChange,
		Payload: VoteStateChangeMessage{
			VoteID:    vote_id,
			Specifics: specifics,
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Failed to marshal event:", err)
		return
	}
	broadcastEvent(string(data))
}

func BroadcastVoteCreation(Vote models.Vote) {
	message := EventMessage{
		Type: VoteCreation,
		Payload: VoteCreationMessage{
			VoteID: Vote.ID,
		},
	}
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Failed to marshal event:", err)
		return
	}
	broadcastEvent(string(data))
}

func BroadcastVoteUpdate(vote_id uint64, position models.Position) {
	var specifics interface{}
	if db.IsPrivate(vote_id) {
		maps, err := db.TallyVotes(vote_id)
		specifics = VoteUpdateAnonymous{
			For:     uint64(maps[models.For]),
			Against: uint64(maps[models.Against]),
			Abstain: uint64(maps[models.Abstain]),
		}
		if err != nil {
			fmt.Errorf("Error Occured Broadcasting: %w", err)
		}
	} else {
		maps, err := db.TallyVotesPersonal(vote_id)
		specifics = VoteUpdatePersonal{
			For:     maps[models.For],
			Against: maps[models.Against],
			Abstain: maps[models.Abstain],
		}
		if err != nil {
			fmt.Errorf("Error Occured Broadcasting: %w", err)
		}
	}
	message := EventMessage{
		Type: VoteUpdate,
		Payload: VoteUpdatePayload{
			VoteID: vote_id,
			Votes:  specifics,
		},
	}
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Failed to marshal event:", err)
		return
	}
	broadcastEvent(string(data))

}
