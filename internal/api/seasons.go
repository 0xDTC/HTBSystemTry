package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Season is a single HTB competitive season.
type Season struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	// State is the season lifecycle label (e.g. "active", "ended").
	State string `json:"state"`
}

// SeasonMachine is a machine that belongs to a season. The IP is only populated
// for the caller's currently active season machine (ActiveSeasonMachine); the
// per-season listing does not expose it.
type SeasonMachine struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Difficulty string `json:"difficulty_text"`
	IP         string `json:"ip"`
	Active     bool   `json:"active"`
	// OS is the machine's operating system (e.g. "Linux", "Windows").
	OS string `json:"os"`
}

// SeasonRank is the caller's standing within a season.
type SeasonRank struct {
	// Rank is the caller's position on the season ladder.
	Rank int `json:"rank"`
	// Points is the caller's accumulated points for the season
	// (mapped from total_season_points).
	Points int `json:"total_season_points"`
	// League is the caller's league/division name for the season.
	League string `json:"league"`
	// TotalRanks is the number of ranked players in the season.
	TotalRanks int `json:"total_ranks"`
}

// ListSeasons returns all seasons known to the platform.
func (c *Client) ListSeasons() ([]Season, error) {
	var resp struct {
		Data []Season `json:"data"`
	}
	if err := c.getJSON("v4", "/season/list", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ActiveSeason returns the currently active season, or an error if none is
// flagged active.
func (c *Client) ActiveSeason() (*Season, error) {
	seasons, err := c.ListSeasons()
	if err != nil {
		return nil, err
	}
	for i := range seasons {
		if seasons[i].Active {
			return &seasons[i], nil
		}
	}
	return nil, fmt.Errorf("no active season")
}

// SeasonMachines returns the machines belonging to the given season.
func (c *Client) SeasonMachines(seasonID int) ([]SeasonMachine, error) {
	var resp struct {
		Data []SeasonMachine `json:"data"`
	}
	if err := c.getJSON("v4", fmt.Sprintf("/season/machines/%d", seasonID), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ActiveSeasonMachine returns the caller's active season machine. It returns a
// nil machine (and nil error) when no machine is currently active.
func (c *Client) ActiveSeasonMachine() (*SeasonMachine, error) {
	var resp struct {
		Data *SeasonMachine `json:"data"`
	}
	if err := c.getJSON("v4", "/season/machine/active", nil, &resp); err != nil {
		return nil, err
	}
	if resp.Data == nil || resp.Data.ID == 0 {
		return nil, nil
	}
	return resp.Data, nil
}

// SeasonUserRank returns the caller's rank and points for the given season.
func (c *Client) SeasonUserRank(seasonID int) (*SeasonRank, error) {
	var resp struct {
		Data SeasonRank `json:"data"`
	}
	if err := c.getJSON("v4", fmt.Sprintf("/season/user/rank/%d", seasonID), nil, &resp); err != nil {
		return nil, err
	}
	rank := resp.Data
	return &rank, nil
}

// UserID extracts the caller's numeric user id from the "sub" claim of the API
// token (a JWT). The token is parsed manually: the payload (second
// dot-separated segment) is base64url-decoded and JSON-decoded. The "sub"
// claim may be encoded as either a JSON number or a string.
func (c *Client) UserID() (int, error) {
	token := c.Token()
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid token: expected JWT with at least 2 segments")
	}

	payload := parts[1]
	// Pad to a multiple of 4 so std/raw decoders that expect padding succeed;
	// RawURLEncoding itself is unpadded, so trim any padding first.
	payload = strings.TrimRight(payload, "=")
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return 0, fmt.Errorf("decode token payload: %w", err)
	}

	var claims struct {
		Sub json.RawMessage `json:"sub"`
	}
	if err := json.Unmarshal(raw, &claims); err != nil {
		return 0, fmt.Errorf("parse token claims: %w", err)
	}
	if len(claims.Sub) == 0 {
		return 0, fmt.Errorf("token has no sub claim")
	}

	// sub may be a JSON number (123) or a JSON string ("123").
	var asString string
	if err := json.Unmarshal(claims.Sub, &asString); err == nil {
		id, err := strconv.Atoi(strings.TrimSpace(asString))
		if err != nil {
			return 0, fmt.Errorf("sub claim %q is not an integer", asString)
		}
		return id, nil
	}

	var asNumber json.Number
	if err := json.Unmarshal(claims.Sub, &asNumber); err == nil {
		id, err := strconv.Atoi(asNumber.String())
		if err != nil {
			return 0, fmt.Errorf("sub claim %q is not an integer", asNumber.String())
		}
		return id, nil
	}

	return 0, fmt.Errorf("unsupported sub claim type: %s", string(claims.Sub))
}
