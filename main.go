package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Poll struct {
	Options   map[string]int
	VotesLock sync.RWMutex
}

type PollStore struct {
	Polls     map[string]*Poll
	StoreLock sync.RWMutex
}

var store = PollStore{
	Polls: make(map[string]*Poll),
}

func createPoll(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PollID  string   `json:"pollId"`
		Options []string `json:"options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	store.StoreLock.Lock()
	defer store.StoreLock.Unlock()

	if _, exists := store.Polls[req.PollID]; exists {
		http.Error(w, "Polls already exists", http.StatusConflict)
		return
	}

	poll := &Poll{
		Options: make(map[string]int),
	}

	for _, option := range req.Options {
		poll.Options[option] = 0
	}

	store.Polls[req.PollID] = poll
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Polls created successfully")
}

func castVote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PollID      string `json:"pollId"`
		OptionVoted string `json:"optionVoted"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	store.StoreLock.RLock()
	poll, exists := store.Polls[req.PollID]
	store.StoreLock.RUnlock()
	if !exists {
		http.Error(w, "Poll not found", http.StatusNotFound)
		return
	}

	poll.VotesLock.Lock()
	defer poll.VotesLock.Unlock()
	if _, ok := poll.Options[req.OptionVoted]; ok {
		http.Error(w, "Option not found", http.StatusBadRequest)
		return
	}

	poll.Options[req.OptionVoted]++
	fmt.Fprintf(w, "Vote cast successfully")
}

func getPollResult(w http.ResponseWriter, r *http.Request) {
	pollId := r.URL.Query().Get("polllId")
	if pollId == "" {
		http.Error(w, "pollId required", http.StatusBadRequest)
		return
	}

	store.StoreLock.RLock()
	poll, exists := store.Polls[pollId]
	store.StoreLock.RUnlock()

	if !exists {
		http.Error(w, "Poll not Found", http.StatusNotFound)
		return
	}
	poll.VotesLock.RLock()
	defer poll.VotesLock.RUnlock()
	json.NewEncoder(w).Encode(poll.Options)
}

func main() {
	http.HandleFunc("/createPoll", createPoll)
	http.HandleFunc("castVote", castVote)
	http.HandleFunc("getsPollResults", getPollResult)

	fmt.Println("Server started at: 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
