package model

import (
	"fmt"
	"strings"
	"time"
)

// ParsedOptionIdentifier holds fields parsed from an OCC option identifier.
type ParsedOptionIdentifier struct {
	Symbol string  // underlying symbol, e.g. "NVDA"
	Expiry string  // formatted as "2026-05-22"
	Right  string  // "PUT" or "CALL"
	Strike float64 // e.g. 340.0
}

// ParseOptionIdentifier parses an OCC-style option identifier like "NVDA  260522P00340000"
// into its component parts: symbol, expiry, right, strike.
//
// OCC format is exactly 21 chars: 6 (symbol right-padded) + 6 (YYMMDD) + 1 (C/P) + 8 (strike*1000).
func ParseOptionIdentifier(identifier string) (*ParsedOptionIdentifier, error) {
	if len(identifier) != 21 {
		return nil, fmt.Errorf("option identifier must be 21 chars, got %d: %q", len(identifier), identifier)
	}

	symbol := strings.TrimSpace(identifier[:6])

	// Parse date: YYMMDD (chars 6-11)
	dateStr := identifier[6:12]
	t, err := time.Parse("060102", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date in identifier: %q", dateStr)
	}

	// Parse right: C or P (char 12)
	rightChar := identifier[12]
	var right string
	switch rightChar {
	case 'C':
		right = "CALL"
	case 'P':
		right = "PUT"
	default:
		return nil, fmt.Errorf("invalid right character: %c", rightChar)
	}

	// Parse strike: 8 digits (chars 13-20), price * 1000
	strikeStr := identifier[13:21]
	var strikeInt int64
	for _, ch := range strikeStr {
		if ch < '0' || ch > '9' {
			return nil, fmt.Errorf("invalid strike digits: %q", strikeStr)
		}
		strikeInt = strikeInt*10 + int64(ch-'0')
	}
	strike := float64(strikeInt) / 1000.0

	return &ParsedOptionIdentifier{
		Symbol: symbol,
		Expiry: t.Format("2006-01-02"),
		Right:  right,
		Strike: strike,
	}, nil
}

// IsOptionIdentifier checks if the given string is a valid OCC option identifier.
// OCC format is exactly 21 chars: 6 (symbol right-padded) + 6 (YYMMDD) + 1 (C/P) + 8 (strike*1000).
// Example: "NVDA  260522P00340000"
func IsOptionIdentifier(s string) bool {
	if len(s) != 21 {
		return false
	}
	c := s[12]
	return c == 'C' || c == 'P'
}
