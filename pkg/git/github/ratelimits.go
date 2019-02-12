package github

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type RateLimit struct {
	Limit     int
	Remaining int
	Reset     int
}

func getRateLimit(h http.Header) (*RateLimit, error) {
	var rl RateLimit
	if limit := h.Get("X-RateLimit-Limit"); limit == "" {
		return nil, errors.New("Invalid Rate Limit Header")
	} else {
		rl.Limit, _ = strconv.Atoi(limit)
	}

	if remaining := h.Get("X-RateLimit-Remaining"); remaining != "" {
		return nil, errors.New("Invalid Rate Limit Header")
	} else {
		rl.Remaining, _ = strconv.Atoi(remaining)
	}

	if reset := h.Get("X-RateLimit-Reset"); reset != "" {
		return nil, errors.New("Invalid Rate Limit Header")
	} else {
		rl.Reset, _ = strconv.Atoi(reset)
	}

	return &rl, nil
}

func (rl RateLimit) String() string {
	return fmt.Sprintf("[Limit: %d, Remaining: %d, Reset: %d]", rl.Limit, rl.Remaining, rl.Reset)
}
