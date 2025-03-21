package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/redis"
	"log"
	"net/http"
	"os"
	"time"
)

type ExchangeRates struct {
	Rates map[string]float64 `json:"rates"`
}

func supportedEventType(eventType string) bool {
	if eventType != "bet" && eventType != "deposit" {
		return false
	}
	return true
}

func supportedCurrency(currency string) bool {
	for _, c := range casino.Currencies {
		if c == currency {
			return true
		}
	}
	return false
}

func getExchangeRates(currency string) (float64, error) {
	exchangeRateApiUrl := os.Getenv("EXCHANGE_RATE_API_URL")
	response, err := http.Get(fmt.Sprintf("%s?base=%s&symbols=EUR", exchangeRateApiUrl, currency))

	if err != nil {
		return 0, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	var exchangeRates ExchangeRates
	if err := json.NewDecoder(response.Body).Decode(&exchangeRates); err != nil {
		return 0, err
	}

	rate, found := exchangeRates.Rates["EUR"]
	if !found {
		return 0, fmt.Errorf("exchange rate not found")
	}

	return rate, nil
}

func GetCommonCurrency(event casino.Event, cache redis.Cache) casino.Event {
	if !supportedEventType(event.Type) {
		log.Println("Unsuported event type")
		return event
	}

	if !supportedCurrency(event.Currency) {
		log.Println("Unsupported currency")
		return event
	}

	if event.Currency == "EUR" {
		event.AmountEUR = event.Amount
		return event
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("exchange_rate_%s", event.Currency)
	rate, found := cache.Get(ctx, cacheKey)

	if !found {
		rate, err := getExchangeRates(event.Currency)
		if err != nil {
			log.Println("Failed to get exchange rate")
			return event
		}
		err = cache.Set(ctx, cacheKey, rate, time.Minute)
		if err != nil {
			log.Println("Failed to set cache api currency")
			return event
		}

	}

	event.AmountEUR = int(float64(event.Amount) * rate)

	return event
}
