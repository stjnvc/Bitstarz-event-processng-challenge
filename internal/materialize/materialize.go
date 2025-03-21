package materialize

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

type Materializer struct {
	eventsTotal                  int
	eventsPerMinute              float64
	eventsPerSecondMovingAverage float64
	topPlayerBets                PlayerStats
	topPlayerWins                PlayerStats
	topPlayerDeposits            PlayerStats
	lastMinuteCounts             []int
	lastMinuteIndex              int
	mu                           sync.Mutex
	eventTimestamps              []time.Time
	playerBets                   map[int]int
	playerWins                   map[int]int
	playerDeposits               map[int]int
}

type PlayerStats struct {
	ID    int `json:"id"`
	Count int `json:"count"`
}

func NewMaterializer() *Materializer {
	return &Materializer{
		lastMinuteCounts: make([]int, 60),
		eventTimestamps:  make([]time.Time, 0),
		playerBets:       make(map[int]int),
		playerWins:       make(map[int]int),
		playerDeposits:   make(map[int]int),
	}
}

func (m *Materializer) AggregateEvents(event casino.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.eventsTotal++

	m.eventTimestamps = append(m.eventTimestamps, time.Now())

	oneMinuteAgo := time.Now().Add(-time.Minute)
	for len(m.eventTimestamps) > 0 && m.eventTimestamps[0].Before(oneMinuteAgo) {
		m.eventTimestamps = m.eventTimestamps[1:]
	}

	m.eventsPerSecondMovingAverage = float64(len(m.eventTimestamps)) / 60.0
	m.eventsPerMinute = float64(len(m.eventTimestamps))

	switch event.Type {
	case "bet":
		m.playerBets[event.PlayerID]++
	case "game_stop":
		if event.HasWon {
			m.playerWins[event.PlayerID]++
		}
	case "deposit":
		m.playerDeposits[event.PlayerID] += event.AmountEUR
	}

	m.topPlayerBets = m.getTopPlayer(m.playerBets)
	m.topPlayerWins = m.getTopPlayer(m.playerWins)
	m.topPlayerDeposits = m.getTopPlayer(m.playerDeposits)
}

func (m *Materializer) getTopPlayer(playerMap map[int]int) PlayerStats {
	var topPlayer PlayerStats
	for id, count := range playerMap {
		if count > topPlayer.Count {
			topPlayer.ID = id
			topPlayer.Count = count
		}
	}
	return topPlayer
}

func (m *Materializer) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data := map[string]interface{}{
		"events_total":                     m.eventsTotal,
		"events_per_minute":                m.eventsPerMinute,
		"events_per_second_moving_average": m.eventsPerSecondMovingAverage,
		"top_player_bets":                  m.topPlayerBets,
		"top_player_wins":                  m.topPlayerWins,
		"top_player_deposits":              m.topPlayerDeposits,
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

func (m *Materializer) StartHTTPServer() {
	http.HandleFunc("/materialized", m.HandleHTTP)
	go func() {
		log.Println("Materialize HTTP server is running on :8080")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()
}
