package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Sherlock is a single HTB Sherlock (defensive/forensics investigation), as
// returned by the listing and detail endpoints.
type Sherlock struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Difficulty string  `json:"difficulty"`
	Category   string  `json:"category_name"`
	Solved     bool    `json:"is_owned"`
	Rating     float64 `json:"rating"`
	// Progress is the completion percentage (0-100) for the calling user.
	Progress int `json:"progress"`
	// Solves is the global number of users who have completed the Sherlock.
	Solves int `json:"solves"`
	// CategoryID is the raw category identifier, used to resolve Category when
	// the listing does not embed the human-readable name.
	CategoryID int `json:"category_id"`
	// Avatar is the Sherlock's logo image URL (may be a relative /storage path).
	Avatar string `json:"avatar"`
	// ReleaseDate is the publish timestamp (used for sorting newest-first).
	ReleaseDate string `json:"release_date"`
	// State is the listing state ("active"/"retired"/...).
	State string `json:"state"`
	// Description is only populated by GetSherlock (the /info endpoint).
	Description string `json:"description"`
}

// SherlockTask is a single question/objective within a Sherlock. Flags are
// submitted per task.
type SherlockTask struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// ListSherlocks returns every Sherlock available to the caller. It walks the
// paginated /sherlocks endpoint (stopping at the reported last page, or after
// a hard cap of 1000 items) and resolves each item's category name from the
// categories list when the listing only provides a category id.
func (c *Client) ListSherlocks() ([]Sherlock, error) {
	const maxItems = 120 // tray displays at most 40; one page is plenty

	cats, err := c.sherlockCategories()
	if err != nil {
		// Category resolution is best-effort; a failure here should not block
		// the (more important) listing. Fall back to whatever names the list
		// endpoint embeds.
		cats = nil
	}

	var all []Sherlock
	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("per_page", "100")

		var resp struct {
			Data []Sherlock `json:"data"`
			Meta struct {
				CurrentPage int `json:"current_page"`
				LastPage    int `json:"last_page"`
				Total       int `json:"total"`
			} `json:"meta"`
		}
		if err := c.getJSON("v4", "/sherlocks", q, &resp); err != nil {
			return nil, err
		}

		for i := range resp.Data {
			if resp.Data[i].Category == "" {
				if name, ok := cats[resp.Data[i].CategoryID]; ok {
					resp.Data[i].Category = name
				}
			}
			all = append(all, resp.Data[i])
			if len(all) >= maxItems {
				return all, nil
			}
		}

		// Stop when the API reports no further pages, or when a page returns
		// nothing (defensive against a missing/zero last_page).
		if len(resp.Data) == 0 || (resp.Meta.LastPage > 0 && page >= resp.Meta.LastPage) {
			break
		}
	}
	return all, nil
}

// sherlockCategories returns a map of category id -> category name from
// /sherlocks/categories/list.
func (c *Client) sherlockCategories() (map[int]string, error) {
	var resp struct {
		Info []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"info"`
	}
	if err := c.getJSON("v4", "/sherlocks/categories/list", nil, &resp); err != nil {
		return nil, err
	}
	out := make(map[int]string, len(resp.Info))
	for _, cat := range resp.Info {
		out[cat.ID] = cat.Name
	}
	return out, nil
}

// GetSherlock returns detail for a single Sherlock by id.
//
// Note: the /sherlocks/{id}/info endpoint returns a sparse object (id,
// description, user_owns_count); it does not echo the name, difficulty or
// category. Use ListSherlocks for those fields.
func (c *Client) GetSherlock(id int) (*Sherlock, error) {
	var resp struct {
		Data struct {
			ID          int    `json:"id"`
			Description string `json:"description"`
			UserOwns    int    `json:"user_owns_count"`
		} `json:"data"`
	}
	if err := c.getJSON("v4", fmt.Sprintf("/sherlocks/%d/info", id), nil, &resp); err != nil {
		return nil, err
	}
	return &Sherlock{
		ID:          resp.Data.ID,
		Description: resp.Data.Description,
		Solves:      resp.Data.UserOwns,
	}, nil
}

// PlaySherlock starts (or continues) the given Sherlock for the caller. The
// endpoint is GET (operationId getSherlockPlay); a POST returns HTTP 405.
func (c *Client) PlaySherlock(id int) error {
	return c.getJSON("v4", fmt.Sprintf("/sherlocks/%d/play", id), nil, nil)
}

// SherlockTasks returns the ordered list of tasks for a Sherlock.
func (c *Client) SherlockTasks(id int) ([]SherlockTask, error) {
	var resp struct {
		Data []SherlockTask `json:"data"`
	}
	if err := c.getJSON("v4", fmt.Sprintf("/sherlocks/%d/tasks", id), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// SubmitSherlockFlag submits a flag for a single task of a Sherlock.
//
// The endpoint returns 201 with a message on success and 400 on an incorrect
// flag; there is no boolean "success" field in the body, so Success is derived
// from a successful (2xx) response.
func (c *Client) SubmitSherlockFlag(sherlockID, taskID int, flag string) (*FlagResponse, error) {
	body := struct {
		Flag string `json:"flag"`
	}{Flag: flag}

	var resp struct {
		Message string `json:"message"`
	}
	path := fmt.Sprintf("/sherlocks/%d/tasks/%d/flag", sherlockID, taskID)
	if err := c.sendJSON(http.MethodPost, "v4", path, body, &resp); err != nil {
		return nil, err
	}
	return &FlagResponse{Success: true, Message: resp.Message}, nil
}

// SherlockDownloadLink returns a temporary URL for downloading the Sherlock's
// evidence/artifact bundle.
func (c *Client) SherlockDownloadLink(id int) (string, error) {
	var resp struct {
		URL       string `json:"url"`
		ExpiresIn int    `json:"expires_in"`
	}
	if err := c.getJSON("v4", fmt.Sprintf("/sherlocks/%d/download_link", id), nil, &resp); err != nil {
		return "", err
	}
	return resp.URL, nil
}
