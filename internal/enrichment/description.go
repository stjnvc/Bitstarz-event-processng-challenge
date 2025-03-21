package enrichment

import (
	"fmt"
	"log"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

func GetHumanReadableDescription(event casino.Event) string {
	switch event.Type {
	case "game_start":
		return gameStartEventDescription(event)
	case "game_stop":
		return gameStopEventDescription(event)
	case "bet":
		return betEventDescription(event)
	case "deposit":
		return depositEventDescription(event)
	default:
		log.Printf("ERROR: Unknown event type: %s", event.Type)
		return "Unknown event"
	}
}

func gameStartEventDescription(event casino.Event) string {
	return fmt.Sprintf("Player #%d started playing a game \"%s\" on %s.",
		event.PlayerID,
		getGameTitle(event.GameID),
		formatTimestamp(event.CreatedAt),
	)
}

func gameStopEventDescription(event casino.Event) string {
	return fmt.Sprintf("Player #%d stopped playing a game \"%s\" on %s.",
		event.PlayerID,
		getGameTitle(event.GameID),
		formatTimestamp(event.CreatedAt),
	)
}

func betEventDescription(event casino.Event) string {
	return fmt.Sprintf("Player #%d (%s) placed a bet of %.2f %s (%.2f EUR) on a game \"%s\" on %s.",
		event.PlayerID,
		event.Player.Email,
		float64(event.Amount)/100,
		event.Currency,
		float64(event.AmountEUR)/100,
		getGameTitle(event.GameID),
		formatTimestamp(event.CreatedAt),
	)
}

func depositEventDescription(event casino.Event) string {
	return fmt.Sprintf("Player #%d made a deposit of %.2f EUR on %s.",
		event.PlayerID,
		float64(event.Amount)/100,
		formatTimestamp(event.CreatedAt),
	)
}

func getGameTitle(gameId int) string {
	if game, found := casino.Games[gameId]; found {
		return game.Title
	}
	log.Printf("WARNING: Game ID %d not found", gameId) // Use log.Printf

	return "Unknown Game"
}

func formatTimestamp(t time.Time) string {
	return t.Format("January 2nd, 2006 at 15:04 UTC")
}
