package main

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	updateToken = os.Getenv("UPDATE_AGENT_TOKEN")
	imageRepo   = os.Getenv("UPDATE_IMAGE_REPOSITORY")
	mu          sync.Mutex
	updating    bool
)

func main() {
	if updateToken == "" || imageRepo == "" {
		log.Fatal("UPDATE_AGENT_TOKEN and UPDATE_IMAGE_REPOSITORY are required")
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
	var input struct { Version string `json:"version"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || !isVersion(input.Version) {
		http.Error(w, "a semantic version is required", http.StatusBadRequest)
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
			command.Env = append(os.Environ(), "APP_IMAGE="+imageRepo+":"+input.Version)
			if output, err := command.CombinedOutput(); err != nil {
				log.Printf("update command failed: docker %v: %v: %s", args, err, output)
				return
			}
		}
		log.Print("application update completed")
	}()
}

func isVersion(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) != 3 { return false }
	for _, part := range parts {
		if part == "" { return false }
		for _, char := range part {
			if char < '0' || char > '9' { return false }
		}
	}
	return true
}
