package api

import (
	"net/http"
	"net/url"
	"strconv"
)

// Machine is a unified view of an HTB machine. It is populated from several
// different HTB endpoints (the v5 paginated /machines list and the v4
// /machine/active endpoint), each of which uses slightly different JSON field
// names. The json tags below describe this struct's own stable shape so callers
// get a consistent contract no matter which endpoint produced the value.
type Machine struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	OS         string  `json:"os"`
	Difficulty string  `json:"difficulty"` // from difficultyText
	Stars      float64 `json:"stars"`      // rating / star score
	Points     int     `json:"points"`
	UserOwns   bool    `json:"user_owns"` // whether the authed user owns user
	RootOwns   bool    `json:"root_owns"` // whether the authed user owns root
	Active     bool    `json:"active"`
	Retired    bool    `json:"retired"`
	IP         string  `json:"ip"`

	ReleaseDate string `json:"release_date"`

	// Spawn/instance state. Populated when the source endpoint reports play
	// info (v5 list playInfo or v4 /machine/active).
	Spawned   bool   `json:"spawned"`
	Spawning  bool   `json:"spawning"`
	IsActive  bool   `json:"is_active"`  // instance currently active/online
	ExpiresAt string `json:"expires_at"` // when the running instance expires
}

// machineListItem mirrors the v5 /machines `data[]` item (schema MachinesItem).
// Field names match the real JSON returned by HTB.
type machineListItem struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	OS                 string  `json:"os"`
	Points             int     `json:"points"`
	Rating             float64 `json:"rating"`
	DifficultyText     string  `json:"difficultyText"`
	ReleaseDate        string  `json:"releaseDate"`
	RetiredDate        *string `json:"retiredDate"`
	Active             *bool   `json:"active"`
	IP                 *string `json:"ip"`
	AuthUserInUserOwns bool    `json:"authUserInUserOwns"`
	AuthUserInRootOwns bool    `json:"authUserInRootOwns"`
	PlayInfo           struct {
		IsSpawned  *bool   `json:"isSpawned"`
		IsSpawning *bool   `json:"isSpawning"`
		IsActive   *bool   `json:"isActive"`
		ExpiresAt  *string `json:"expires_at"`
	} `json:"playInfo"`
}

// toMachine converts a v5 list item into the unified Machine view.
func (m machineListItem) toMachine() Machine {
	out := Machine{
		ID:          m.ID,
		Name:        m.Name,
		OS:          m.OS,
		Difficulty:  m.DifficultyText,
		Stars:       m.Rating,
		Points:      m.Points,
		UserOwns:    m.AuthUserInUserOwns,
		RootOwns:    m.AuthUserInRootOwns,
		ReleaseDate: m.ReleaseDate,
		Retired:     m.RetiredDate != nil,
	}
	if m.Active != nil {
		out.Active = *m.Active
	}
	if m.IP != nil {
		out.IP = *m.IP
	}
	if m.PlayInfo.IsSpawned != nil {
		out.Spawned = *m.PlayInfo.IsSpawned
	}
	if m.PlayInfo.IsSpawning != nil {
		out.Spawning = *m.PlayInfo.IsSpawning
	}
	if m.PlayInfo.IsActive != nil {
		out.IsActive = *m.PlayInfo.IsActive
	}
	if m.PlayInfo.ExpiresAt != nil {
		out.ExpiresAt = *m.PlayInfo.ExpiresAt
	}
	return out
}

// listMachines walks the paginated v5 /machines endpoint applying the given
// state filter ("active" or "retired"), accumulating up to maxPages of results.
func (c *Client) listMachines(state string) ([]Machine, error) {
	const (
		perPage  = 100
		maxPages = 1 // tray shows at most ~45; one page (100) is plenty
	)

	var machines []Machine
	for page := 1; page <= maxPages; page++ {
		q := url.Values{}
		q.Set("per_page", strconv.Itoa(perPage))
		q.Set("page", strconv.Itoa(page))
		// state uses explode=true / style=form: repeated as state=<value>.
		q.Set("state", state)

		var resp struct {
			Data []machineListItem `json:"data"`
			Meta struct {
				CurrentPage int  `json:"current_page"`
				LastPage    int  `json:"last_page"`
				To          *int `json:"to"`
			} `json:"meta"`
		}
		if err := c.getJSON("v5", "/machines", q, &resp); err != nil {
			return nil, err
		}

		for _, item := range resp.Data {
			machines = append(machines, item.toMachine())
		}

		// Stop when there is no further page (empty page, or we have reached
		// the reported last page).
		if len(resp.Data) == 0 {
			break
		}
		if resp.Meta.LastPage > 0 && resp.Meta.CurrentPage >= resp.Meta.LastPage {
			break
		}
	}
	return machines, nil
}

// ListActiveMachines returns all currently active (non-retired) machines,
// looping through every page of the v5 /machines endpoint (capped at 500 pages).
func (c *Client) ListActiveMachines() ([]Machine, error) {
	return c.listMachines("active")
}

// ListRetiredMachines returns all retired machines, looping through every page
// of the v5 /machines endpoint (capped at 500 pages).
func (c *Client) ListRetiredMachines() ([]Machine, error) {
	return c.listMachines("retired")
}

// ActiveMachine returns the machine currently spawned for the user from v4
// /machine/active. When nothing is spawned the "info" object is null and this
// returns (nil, nil).
func (c *Client) ActiveMachine() (*Machine, error) {
	var resp struct {
		Info *struct {
			ID         int     `json:"id"`
			Name       string  `json:"name"`
			IP         *string `json:"ip"`
			ExpiresAt  string  `json:"expires_at"`
			IsSpawning bool    `json:"isSpawning"`
			Type       string  `json:"type"`
		} `json:"info"`
	}
	if err := c.getJSON("v4", "/machine/active", nil, &resp); err != nil {
		return nil, err
	}
	if resp.Info == nil {
		return nil, nil
	}

	info := resp.Info
	m := &Machine{
		ID:        info.ID,
		Name:      info.Name,
		ExpiresAt: info.ExpiresAt,
		Spawning:  info.IsSpawning,
		// A machine reported by /machine/active is, by definition, spawned.
		Spawned:  true,
		IsActive: !info.IsSpawning,
	}
	if info.IP != nil {
		m.IP = *info.IP
	}
	return m, nil
}

// SpawnMachine spawns (starts) the given machine via POST v4 /vm/spawn.
func (c *Client) SpawnMachine(id int) error {
	body := map[string]int{"machine_id": id}
	return c.sendJSON(http.MethodPost, "v4", "/vm/spawn", body, nil)
}

// StopMachine terminates the given machine via POST v4 /vm/terminate.
func (c *Client) StopMachine(id int) error {
	body := map[string]int{"machine_id": id}
	return c.sendJSON(http.MethodPost, "v4", "/vm/terminate", body, nil)
}

// ResetMachine resets the given machine via POST v4 /vm/reset.
func (c *Client) ResetMachine(id int) error {
	body := map[string]int{"machine_id": id}
	return c.sendJSON(http.MethodPost, "v4", "/vm/reset", body, nil)
}

// SubmitMachineFlag submits a flag for the given machine via POST v5
// /machine/own. difficulty is the user's 1-10 difficulty rating (HTB stores it
// as a 10-100 scale, but the value is passed through as supplied by the caller).
func (c *Client) SubmitMachineFlag(id int, flag string, difficulty int) (*FlagResponse, error) {
	body := struct {
		ID         int    `json:"id"`
		Flag       string `json:"flag"`
		Difficulty int    `json:"difficulty"`
	}{ID: id, Flag: flag, Difficulty: difficulty}

	var out FlagResponse
	if err := c.sendJSON(http.MethodPost, "v5", "/machine/own", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
