package api

import (
	"encoding/json"
	"net/url"
)

// SearchHit is a single global-search result. Name is taken from the entity's
// display value, and Extra carries an optional secondary attribute when the
// underlying entity provides one (e.g. category for challenges/sherlocks, or a
// team motto). It is empty when no such attribute exists for that entity type.
type SearchHit struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Extra string `json:"extra"`
}

// SearchResults groups global-search hits by entity type. Only the lists that
// the API returned for the requested tags are populated; the rest stay nil.
type SearchResults struct {
	Machines   []SearchHit `json:"machines"`
	Challenges []SearchHit `json:"challenges"`
	Sherlocks  []SearchHit `json:"sherlocks"`
	Users      []SearchHit `json:"users"`
	Teams      []SearchHit `json:"teams"`
}

// Search performs an HTB global search. tags restricts the search to specific
// entity kinds; valid values are "machines", "challenges", "users" and "teams"
// (the API accepts at most one). When tags is empty the tags parameter is
// omitted and the server decides which result sets to return.
//
// GET v4 /search/fetch?query=...&tags=...
//
// The tags parameter is encoded the way the HTB API expects: a single query
// value holding a JSON array string, e.g. tags=["challenges"].
func (c *Client) Search(query string, tags []string) (*SearchResults, error) {
	q := url.Values{}
	q.Set("query", query)
	if len(tags) > 0 {
		encoded, err := json.Marshal(tags)
		if err != nil {
			return nil, err
		}
		q.Set("tags", string(encoded))
	}

	// The response is heterogeneous: each entity type has its own item shape,
	// but they all share id + value, plus type-specific extras we surface.
	var raw struct {
		Machines []struct {
			ID    int    `json:"id"`
			Value string `json:"value"`
		} `json:"machines"`
		Challenges []struct {
			ID       int    `json:"id"`
			Value    string `json:"value"`
			Category string `json:"category_name"`
		} `json:"challenges"`
		Sherlocks []struct {
			ID       int    `json:"id"`
			Value    string `json:"value"`
			Category string `json:"category_name"`
		} `json:"sherlocks"`
		Users []struct {
			ID    int    `json:"id"`
			Value string `json:"value"`
		} `json:"users"`
		Teams []struct {
			ID    int    `json:"id"`
			Value string `json:"value"`
			Motto string `json:"motto"`
		} `json:"teams"`
	}
	if err := c.getJSON("v4", "/search/fetch", q, &raw); err != nil {
		return nil, err
	}

	res := &SearchResults{}
	for _, m := range raw.Machines {
		res.Machines = append(res.Machines, SearchHit{ID: m.ID, Name: m.Value})
	}
	for _, ch := range raw.Challenges {
		res.Challenges = append(res.Challenges, SearchHit{ID: ch.ID, Name: ch.Value, Extra: ch.Category})
	}
	for _, s := range raw.Sherlocks {
		res.Sherlocks = append(res.Sherlocks, SearchHit{ID: s.ID, Name: s.Value, Extra: s.Category})
	}
	for _, u := range raw.Users {
		res.Users = append(res.Users, SearchHit{ID: u.ID, Name: u.Value})
	}
	for _, t := range raw.Teams {
		res.Teams = append(res.Teams, SearchHit{ID: t.ID, Name: t.Value, Extra: t.Motto})
	}
	return res, nil
}
