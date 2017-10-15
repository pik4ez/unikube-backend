package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"

	"github.com/gorilla/handlers"
)

// playerIDs holds a list of available player IDs
var availablePlayerIDs = map[int]string{
	1: "uno",
	2: "dos",
}

// lastAssignedUser is an index in availablePlayerIDs list which was returned
// from init handler the last time
var lastAssignedUser int

// db holds database connection
var db *sql.DB

// local hp start
var localHP = make(map[string]int)

// handlerState prints player health
func handlerState(w http.ResponseWriter, r *http.Request) {
	playerID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/state/"), "/")
	fmt.Fprintf(
		w,
		`{"result": "ok", "player": "%s", "hp": %d}`,
		playerID,
		localHP[playerID],
	)
}

// handlerInit creates player and prints its ID.
// if player ID provided as argument, restores health of provided player.
func handlerInit(w http.ResponseWriter, r *http.Request) {
	argPlayerID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/init/"), "/")
	if argPlayerID == "" {
		curUserID := lastAssignedUser + 1
		lastAssignedUser = curUserID
		var playerID string
		if lastAssignedUser%2 == 1 {
			playerID = availablePlayerIDs[1]
		} else {
			playerID = availablePlayerIDs[2]
		}
		localHP[playerID] = 100
		fmt.Fprintf(w, `{"result": "ok", "player_id": "%s"}`, playerID)
		return
	}
	localHP[argPlayerID] = 100
	fmt.Fprintf(w, `{"result": "ok", "player_id": "%s"}`, argPlayerID)
}

// handlerDamage commits damage received by user
func handlerDamage(w http.ResponseWriter, r *http.Request) {
	playerID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/damage/"), "/")
	if localHP[playerID] > 0 {
		localHP[playerID] -= 20
	}
	fmt.Fprintf(w, `{"result": "ok", "player_id": "%s"}`, playerID)
}

func main() {
	dbConnParams := fmt.Sprintf(
		"host=%s port=%d user=%s sslmode=disable binary_parameters=yes connect_timeout=%d",
		"192.168.64.12",
		31893,
		"postgres",
		10,
	)
	var err error
	db, err = sql.Open("postgres", dbConnParams)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	http.HandleFunc("/state/", handlerState)
	http.HandleFunc("/init/", handlerInit)
	http.HandleFunc("/damage/", handlerDamage)
	http.ListenAndServe(":8893", handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
}
