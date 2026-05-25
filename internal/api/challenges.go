package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Challenge is a normalized view of an HTB challenge, suitable for list and
// detail rendering. The list endpoint (ListChallenges) does not return Points
// or Likes, so those may be zero.
type Challenge struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Difficulty  string `json:"difficulty"`
	Category    string `json:"category"`
	Points      int    `json:"points"`
	Solves      int    `json:"solves"`
	Likes       int    `json:"likes"`
	Retired     bool   `json:"retired"`
	Solved      bool   `json:"solved"`
	ReleaseDate string `json:"release_date"`
}

// ChallengeCategories returns the challenge category id -> name mapping.
//
// GET v4 /challenge/categories/list
func (c *Client) ChallengeCategories() (map[int]string, error) {
	var out struct {
		Info []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"info"`
	}
	if err := c.getJSON("v4", "/challenge/categories/list", nil, &out); err != nil {
		return nil, err
	}
	cats := make(map[int]string, len(out.Info))
	for _, ci := range out.Info {
		cats[ci.ID] = ci.Name
	}
	return cats, nil
}

// ListChallenges returns all challenges, fetching every page (capped at 1000
// items). When retired is true only retired challenges are requested, otherwise
// only active (non-retired) ones. Category names are resolved via the category
// id->name map. Note the list endpoint does not expose points or likes.
//
// GET v4 /challenges
func (c *Client) ListChallenges(retired bool) ([]Challenge, error) {
	cats, err := c.ChallengeCategories()
	if err != nil {
		// Category resolution is best-effort; continue with ids only.
		cats = map[int]string{}
	}

	const maxItems = 120 // tray displays at most ~45; one page is plenty
	state := "active"
	if retired {
		state = "retired"
	}

	var challenges []Challenge
	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("per_page", "100")
		q.Set("state", state)

		var out struct {
			Data []struct {
				ID          int    `json:"id"`
				Name        string `json:"name"`
				Difficulty  string `json:"difficulty"`
				CategoryID  int    `json:"category_id"`
				Category    string `json:"category_name"`
				Solves      int    `json:"solves"`
				IsOwned     bool   `json:"is_owned"`
				State       string `json:"state"`
				ReleaseDate string `json:"release_date"`
			} `json:"data"`
			Meta struct {
				CurrentPage int `json:"current_page"`
				LastPage    int `json:"last_page"`
			} `json:"meta"`
		}
		if err := c.getJSON("v4", "/challenges", q, &out); err != nil {
			return nil, err
		}
		if len(out.Data) == 0 {
			break
		}
		for _, ch := range out.Data {
			category := ch.Category
			if category == "" {
				if name, ok := cats[ch.CategoryID]; ok {
					category = name
				}
			}
			isRetired := retired
			if ch.State != "" {
				isRetired = strings.EqualFold(ch.State, "retired")
			}
			challenges = append(challenges, Challenge{
				ID:          ch.ID,
				Name:        ch.Name,
				Difficulty:  ch.Difficulty,
				Category:    category,
				Solves:      ch.Solves,
				Retired:     isRetired,
				Solved:      ch.IsOwned,
				ReleaseDate: ch.ReleaseDate,
			})
			if len(challenges) >= maxItems {
				return challenges, nil
			}
		}
		// Stop when we have consumed the last page (or pagination metadata is absent).
		if out.Meta.LastPage > 0 && page >= out.Meta.LastPage {
			break
		}
	}
	return challenges, nil
}

// SubmitChallengeFlag submits a flag for the given challenge.
//
// POST v4 /challenge/own  body: {"challenge_id", "flag"}
func (c *Client) SubmitChallengeFlag(id int, flag string) (*FlagResponse, error) {
	body := struct {
		ChallengeID int    `json:"challenge_id"`
		Flag        string `json:"flag"`
	}{ChallengeID: id, Flag: strings.TrimSpace(flag)}

	// The endpoint returns {"message", ...} with no explicit success field on
	// the happy path (non-2xx is surfaced as an error by the transport), so a
	// decoded response without an error is treated as success.
	var out FlagResponse
	if err := c.sendJSON(http.MethodPost, "v4", "/challenge/own", body, &out); err != nil {
		return nil, err
	}
	if out.Message != "" && !out.Success {
		out.Success = true
	}
	return &out, nil
}

// StartChallenge starts the challenge's docker container and returns a
// human-readable connection string. The documented response is just
// {"id","message"} (for example "Instance Created!"), but some challenge
// containers also report their docker endpoint, so we decode flexibly and
// prefer an "ip:port" string when host/port fields are present, falling back
// to the response message. A 2xx with nothing useful returns ("", nil).
//
// POST v4 /challenge/start  body: {"challenge_id"}
func (c *Client) StartChallenge(id int) (string, error) {
	body := struct {
		ChallengeID int `json:"challenge_id"`
	}{ChallengeID: id}

	// Flexible shape: HTB has used several field names across versions for the
	// container endpoint, so accept any that show up.
	var out struct {
		Message     string `json:"message"`
		IP          string `json:"ip"`
		Host        string `json:"host"`
		DockerIP    string `json:"docker_ip"`
		Port        int    `json:"port"`
		Ports       []int  `json:"ports"`
		DockerPorts []int  `json:"docker_ports"`
	}
	if err := c.sendJSON(http.MethodPost, "v4", "/challenge/start", body, &out); err != nil {
		return "", err
	}

	host := out.IP
	if host == "" {
		host = out.Host
	}
	if host == "" {
		host = out.DockerIP
	}

	port := out.Port
	if port == 0 && len(out.Ports) > 0 {
		port = out.Ports[0]
	}
	if port == 0 && len(out.DockerPorts) > 0 {
		port = out.DockerPorts[0]
	}

	switch {
	case host != "" && port != 0:
		return fmt.Sprintf("%s:%d", host, port), nil
	case host != "":
		return host, nil
	default:
		return strings.TrimSpace(out.Message), nil
	}
}

// StopChallenge stops the challenge's docker container.
//
// POST v4 /challenge/stop  body: {"challenge_id"}
func (c *Client) StopChallenge(id int) error {
	body := struct {
		ChallengeID int `json:"challenge_id"`
	}{ChallengeID: id}
	return c.sendJSON(http.MethodPost, "v4", "/challenge/stop", body, nil)
}
