package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type World struct {
	ID          string                 `json:"id"`
	Prompt      string                 `json:"prompt"`
	Refinements []string               `json:"refinements"`
	State       map[string]interface{} `json:"state"`
	Log         []string               `json:"log"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

var (
	store = make(map[string]*World)
	mu    sync.RWMutex
)

func main() {
	// ✅ Moved inside main()
	dataDir := filepath.Join("..", "data") // relative to cmd/voidspark/

	os.MkdirAll(filepath.Join(dataDir, "worlds"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "web", "preview"), 0755)

	mux := http.NewServeMux()
	mux.HandleFunc("/generate", generateHandler)
	mux.HandleFunc("/refine", refineHandler)
	mux.HandleFunc("/party", partyHandler)
	mux.HandleFunc("/explore", exploreHandler)
	mux.HandleFunc("/state", stateHandler)
	mux.HandleFunc("/api/latest-world", latestWorldHandler)
	mux.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir(dataDir)))) // ✅ Use dataDir

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})

	log.Println("✅ Void Spark — pure user-defined world engine")
	log.Println("🌀 Preview: http://localhost:8080/data/web/preview/world_preview.html")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct{ Prompt string `json:"prompt"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	id := time.Now().UnixNano()
	wld := &World{
		ID:          strconv.FormatInt(id, 10), // ✅ FIXED
		Prompt:      req.Prompt,
		Refinements: []string{req.Prompt},
		State:       make(map[string]interface{}),
		Log:         []string{"World created from prompt"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mu.Lock()
	store[wld.ID] = wld // ✅ FIXED
	mu.Unlock()

	saveWorld(wld)
	writeJSON(w, wld)
}

func refineHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string `json:"id"`
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	mu.Lock()
	wld, ok := store[req.ID]
	mu.Unlock()
	if !ok {
		http.Error(w, `{"error":"world not found"}`, http.StatusNotFound)
		return
	}

	wld.Refinements = append(wld.Refinements, req.Prompt)
	wld.Log = append(wld.Log, "Refined: "+req.Prompt)
	wld.UpdatedAt = time.Now()

	mu.Lock()
	store[req.ID] = wld
	mu.Unlock()

	saveWorld(wld)
	writeJSON(w, wld)
}

func partyHandler(w http.ResponseWriter, r *http.Request) {
	var req struct{ ID string `json:"id"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	mu.Lock()
	wld, ok := store[req.ID]
	mu.Unlock()
	if !ok {
		http.Error(w, `{"error":"world not found"}`, http.StatusNotFound)
		return
	}

	if wld.State == nil {
		wld.State = make(map[string]interface{})
	}
	if _, exists := wld.State["agents"]; !exists {
		wld.State["agents"] = []map[string]interface{}{
			{"id": "agent-1", "name": "Drone-Taxi", "type": "drone", "hp": 100},
			{"id": "agent-2", "name": "Hacker", "type": "human", "hp": 80},
		}
		wld.Log = append(wld.Log, "Agents added from prompt context")
		wld.UpdatedAt = time.Now()
	}

	writeJSON(w, wld)
}

func exploreHandler(w http.ResponseWriter, r *http.Request) {
	var req struct{ ID string `json:"id"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	mu.Lock()
	wld, ok := store[req.ID]
	mu.Unlock()
	if !ok {
		http.Error(w, `{"error":"world not found"}`, http.StatusNotFound)
		return
	}

	wld.Log = append(wld.Log, "Explored world per user action")
	wld.UpdatedAt = time.Now()

	writeJSON(w, wld)
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	mu.RLock()
	wld := store[id]
	mu.RUnlock()
	if wld == nil {
		http.Error(w, `{"error":"world not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, wld)
}

func latestWorldHandler(w http.ResponseWriter, r *http.Request) {
	files, _ := filepath.Glob("../data/worlds/*.json")
	if len(files) == 0 {
		http.Error(w, `{"error":"no worlds"}`, http.StatusNotFound)
		return
	}

	var latest string
	for _, f := range files {
		if f > latest {
			latest = f
		}
	}

	name := filepath.Base(latest)
	writeJSON(w, map[string]string{"latest": name})
}

func saveWorld(wld *World) {
	data, err := json.MarshalIndent(wld, "", "  ")
	if err != nil {
		log.Printf("⚠️ JSON marshal failed for world %s: %v", wld.ID, err)
		return
	}
	path := filepath.Join("../data", "worlds", "world_"+wld.ID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("⚠️ Save failed for world %s: %v", wld.ID, err)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}