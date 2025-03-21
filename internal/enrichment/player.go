package enrichment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/postgres"
)

type PlayerRepository struct {
	db *postgres.PostgresDB
}

func NewPlayerRepository(db *postgres.PostgresDB) (*PlayerRepository, error) {
	if db == nil {
		return nil, errors.New("postgres database is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging postgres: %w", err)
	}

	return &PlayerRepository{db: db}, nil
}

func (r *PlayerRepository) GetPlayerByID(playerID int) (casino.Player, error) {
	const query = `SELECT email, last_signed_in_at FROM players WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(ctx, query, playerID)
	var player casino.Player
	err := row.Scan(&player.Email, &player.LastSignedInAt)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		log.Printf("Player %d not found in database", playerID)
		return casino.Player{}, nil // Return zero-value player, no error
	case err != nil:
		return casino.Player{}, fmt.Errorf("error fetching player: %w", err)
	}

	return player, nil
}

func (r *PlayerRepository) Close() {
	r.db.Close()
}

type PlayerService struct {
	repo *PlayerRepository
}

func NewPlayerService(repo *PlayerRepository) *PlayerService {
	return &PlayerService{repo: repo}
}

func (s *PlayerService) GetPlayer(playerID int) (casino.Player, error) {
	player, err := s.repo.GetPlayerByID(playerID)
	if err != nil {
		return casino.Player{}, fmt.Errorf("player service: %w", err)
	}
	return player, nil
}
