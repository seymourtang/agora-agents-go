package core

import (
	"fmt"
	"log"
)

const (
	DefaultExpirySeconds = 86400
	MaxExpirySeconds     = 86400
)

func ValidateExpiresIn(secs int) (int, error) {
	if secs <= 0 {
		return 0, fmt.Errorf("expiresIn must be between 1 and 86400 seconds (24h)")
	}
	if secs > MaxExpirySeconds {
		log.Println("agora-agent-sdk: expiresIn capped at 24h (Agora max)")
		return MaxExpirySeconds, nil
	}
	return secs, nil
}
