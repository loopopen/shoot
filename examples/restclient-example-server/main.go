package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	store = make(map[string]string)
	mu    sync.RWMutex
)

func main() {
	http.HandleFunc("/set", setHandler)
	http.HandleFunc("/get", getHandler)
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	mu.Lock()
	store[req.Key] = req.Value
	mu.Unlock()
	time.Sleep(time.Second)
	w.WriteHeader(http.StatusOK)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	mu.RLock()
	value, ok := store[key]
	mu.RUnlock()
	if !ok {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	resp := map[string]string{"key": key, "value": value}
	w.Header().Set("Content-Type", "application/json")
	time.Sleep(time.Second)
	json.NewEncoder(w).Encode(resp)
}
