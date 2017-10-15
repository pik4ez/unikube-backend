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

// handlerState prints player health
func handlerState(w http.ResponseWriter, r *http.Request) {
	playerID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/state/"), "/")
	query := fmt.Sprintf(`
		select hp from players where player_id = '%s'
	`, playerID)
	var hp int
	err := db.QueryRow(query).Scan(&hp)
	if err != nil {
		log.Printf(`failed to get user %s hp, query %s, err: %s`, playerID, query, err)
		fmt.Fprint(w, `{"result": "err", "player_id": ""}`)
		return
	}
	fmt.Fprintf(
		w,
		`{"result": "ok", "player": "%s", "hp": %d}`,
		playerID,
		hp,
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
		query := fmt.Sprintf(`
			delete from players where player_id = '%[1]s';
			insert into players (player_id, hp) values ('%[1]s', 100);
		`, playerID)
		_, err := db.Exec(query)
		if err != nil {
			log.Printf(`failed to init player %s, query %s, err: %s`, playerID, query, err)
			fmt.Fprint(w, `{"result": "err", "player_id": ""}`)
			return
		}
		fmt.Fprintf(w, `{"result": "ok", "player_id": "%s"}`, playerID)
		return
	}
	query := fmt.Sprintf(`
		update players set hp = 100 where player_id = '%s'
	`, argPlayerID)
	_, err := db.Exec(query)
	if err != nil {
		log.Printf(`failed to restore hp for player %s, query %s, err: %s`, argPlayerID, query, err)
		fmt.Fprint(w, `{"result": "err", "player_id": ""}`)
		return
	}
	fmt.Fprintf(w, `{"result": "ok", "player_id": "%s"}`, argPlayerID)
}

// handlerDamage commits damage received by user
func handlerDamage(w http.ResponseWriter, r *http.Request) {
	playerID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/damage/"), "/")
	query := fmt.Sprintf(`
		update players set hp = hp - 20 where hp > 0 and player_id = '%s'
	`, playerID)
	_, err := db.Exec(query)
	if err != nil {
		log.Printf(`failed to update hp for player %s, query %s, err: %s`, playerID, query, err)
		fmt.Fprint(w, `{"result": "err", "player_id": ""}`)
		return
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
