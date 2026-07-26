package main

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

var (
	updateToken = os.Getenv("UPDATE_AGENT_TOKEN")
	mu          sync.Mutex
	updating    bool
)

func main() {
	if updateToken == "" {
		log.Fatal("UPDATE_AGENT_TOKEN is required")
	}

	http.HandleFunc("/update", update)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	log.Fatal(http.ListenAndServe(":8090", nil))
}

func update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Update-Token")), []byte(updateToken)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	if updating {
		mu.Unlock()
		http.Error(w, "update already running", http.StatusConflict)
		return
	}
	updating = true
	mu.Unlock()

	json.NewEncoder(w).Encode(map[string]string{"status": "scheduled"})
	go func() {
		// Return to the administrator before the app container is recreated.
		time.Sleep(time.Second)
		defer func() { mu.Lock(); updating = false; mu.Unlock() }()
		for _, args := range [][]string{
			{"compose", "-f", "/deploy/docker-compose.yml", "pull", "app"},
			{"compose", "-f", "/deploy/docker-compose.yml", "up", "-d", "--no-deps", "--no-build", "app"},
		} {
			command := exec.Command("docker", args...)
			command.Dir = "/deploy"
			if output, err := command.CombinedOutput(); err != nil {
				log.Printf("update command failed: docker %v: %v: %s", args, err, output)
				return
			}
		}
		log.Print("application update completed")
	}()
}
