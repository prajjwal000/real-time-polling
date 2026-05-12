package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"strconv"

	"github.com/redis/go-redis/v9"
	"polling-app/backend/cache"
)

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return real
	}
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}
	return addr
}

type PollHandler struct {
	DB  *sql.DB
	RDB *redis.Client
}

type Poll struct {
	ID        int64     `json:"id"`
	Question  string    `json:"question"`
	CreatedAt string    `json:"created_at"`
	Options   []Option  `json:"options,omitempty"`
}

type Option struct {
	ID     int64  `json:"id"`
	PollID int64  `json:"poll_id"`
	Text   string `json:"text"`
}

type VoteCount struct {
	OptionID int64 `json:"option_id"`
	Count    int   `json:"count"`
}

type CreatePollRequest struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

type VoteRequest struct {
	OptionID int64 `json:"option_id"`
}

func (h *PollHandler) ListPolls(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, question, created_at FROM polls ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	polls := []Poll{}
	for rows.Next() {
		var p Poll
		if err := rows.Scan(&p.ID, &p.Question, &p.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		polls = append(polls, p)
	}

	writeJSON(w, http.StatusOK, polls)
}

func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	var req CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Question == "" {
		http.Error(w, "question is required", http.StatusBadRequest)
		return
	}
	if len(req.Options) < 2 {
		http.Error(w, "at least 2 options required", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var pollID int64
	err = tx.QueryRow("INSERT INTO polls (question) VALUES ($1) RETURNING id", req.Question).Scan(&pollID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, optText := range req.Options {
		_, err = tx.Exec("INSERT INTO options (poll_id, text) VALUES ($1, $2)", pollID, optText)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	poll := Poll{ID: pollID, Question: req.Question}
	writeJSON(w, http.StatusCreated, poll)
}

func (h *PollHandler) GetPoll(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid poll id", http.StatusBadRequest)
		return
	}

	var poll Poll
	err = h.DB.QueryRow("SELECT id, question, created_at FROM polls WHERE id = $1", id).
		Scan(&poll.ID, &poll.Question, &poll.CreatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "poll not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := h.DB.Query("SELECT id, poll_id, text FROM options WHERE poll_id = $1 ORDER BY id", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var opt Option
		if err := rows.Scan(&opt.ID, &opt.PollID, &opt.Text); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		poll.Options = append(poll.Options, opt)
	}

	writeJSON(w, http.StatusOK, poll)
}

func (h *PollHandler) Vote(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	pollID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid poll id", http.StatusBadRequest)
		return
	}

	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ip := clientIP(r)

	voterKey := "poll:" + idStr + ":voters"
	resultsKey := "poll:" + idStr + ":results"
	optStr := strconv.FormatInt(req.OptionID, 10)

	voted, err := h.RDB.SIsMember(cache.Ctx, voterKey, ip).Result()
	if err == nil && voted {
		http.Error(w, "already voted", http.StatusConflict)
		return
	}

	var exists bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM votes WHERE poll_id = $1 AND ip_address = $2)", pollID, ip).Scan(&exists)
	if exists {
		http.Error(w, "already voted", http.StatusConflict)
		return
	}

	pipe := h.RDB.Pipeline()
	pipe.HIncrBy(cache.Ctx, resultsKey, optStr, 1)
	pipe.SAdd(cache.Ctx, voterKey, ip)
	pipe.Expire(cache.Ctx, voterKey, 0)
	if _, err := pipe.Exec(cache.Ctx); err != nil {
		log.Printf("Redis pipeline error: %v", err)
	}

	_, dbErr := h.DB.Exec(
		"INSERT INTO votes (poll_id, option_id, ip_address) VALUES ($1, $2, $3) ON CONFLICT (poll_id, ip_address) DO NOTHING",
		pollID, req.OptionID, ip,
	)
	if dbErr != nil {
		log.Printf("DB vote insert error: %v", dbErr)
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *PollHandler) HasVoted(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	pollID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid poll id", http.StatusBadRequest)
		return
	}

	ip := clientIP(r)

	voterKey := "poll:" + idStr + ":voters"
	voted, err := h.RDB.SIsMember(cache.Ctx, voterKey, ip).Result()
	if err == nil && voted {
		writeJSON(w, http.StatusOK, map[string]bool{"voted": true})
		return
	}

	var exists bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM votes WHERE poll_id = $1 AND ip_address = $2)", pollID, ip).Scan(&exists)
	writeJSON(w, http.StatusOK, map[string]bool{"voted": exists})
}

func (h *PollHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	pollID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid poll id", http.StatusBadRequest)
		return
	}

	resultsKey := "poll:" + idStr + ":results"
	redisResults, err := h.RDB.HGetAll(cache.Ctx, resultsKey).Result()
	if err == nil && len(redisResults) > 0 {
		counts := []VoteCount{}
		for optID, countStr := range redisResults {
			oid, _ := strconv.ParseInt(optID, 10, 64)
			c, _ := strconv.Atoi(countStr)
			counts = append(counts, VoteCount{OptionID: oid, Count: c})
		}
		writeJSON(w, http.StatusOK, counts)
		return
	}

	rows, err := h.DB.Query(
		"SELECT option_id, COUNT(*) as count FROM votes WHERE poll_id = $1 GROUP BY option_id",
		pollID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	counts := []VoteCount{}
	for rows.Next() {
		var vc VoteCount
		if err := rows.Scan(&vc.OptionID, &vc.Count); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		counts = append(counts, vc)
	}

	writeJSON(w, http.StatusOK, counts)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
