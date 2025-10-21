package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL = "https://labs.hackthebox.com/api/v4"
)

type Client struct {
	apiToken   string
	httpClient *http.Client
}

// Machine represents an HTB machine
type Machine struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	OS               string   `json:"os"`
	Difficulty       string   `json:"difficulty"`
	DifficultyChart  map[string]int `json:"difficultyChart"`
	Star             float64  `json:"star"`
	UserOwns         int      `json:"user_owns"`
	RootOwns         int      `json:"root_owns"`
	Active           bool     `json:"active"`
	Retired          bool     `json:"retired"`
	Free             bool     `json:"free"`
	Avatar           string   `json:"avatar"`
	IP               string   `json:"ip"`
	Release          string   `json:"release"`
	RetiredDate      string   `json:"retired_date"`
	PlayInfo         *PlayInfo `json:"playInfo,omitempty"`
}

type PlayInfo struct {
	IsSpawned     bool   `json:"isSpawned"`
	IsSpawning    bool   `json:"isSpawning"`
	IsActive      bool   `json:"isActive"`
	ActiveMachine int    `json:"active_player_machine_id"`
	ExpiresAt     string `json:"expires_at"`
}

// Challenge represents an HTB challenge
type Challenge struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Difficulty      string   `json:"difficulty"`
	DifficultyChart map[string]int `json:"difficultyChart"`
	Category        string   `json:"category_name"`
	Points          int      `json:"points"`
	Solves          int      `json:"solves"`
	Likes           int      `json:"likes"`
	Dislikes        int      `json:"dislikes"`
	Retired         bool     `json:"retired"`
	Docker          bool     `json:"docker"`
	Downloaded      bool     `json:"downloaded"`
	IsSolved        bool     `json:"isSolved"`
}

// VPNServer represents a VPN server location
type VPNServer struct {
	ID           int    `json:"id"`
	FriendlyName string `json:"friendly_name"`
	Name         string `json:"name"`
	Location     string `json:"location"`
	CurrentClients int  `json:"current_clients"`
}

// FlagSubmission represents a flag submission request
type FlagSubmission struct {
	Flag       string `json:"flag"`
	Difficulty int    `json:"difficulty"`
	ID         int    `json:"id"`
}

// FlagResponse represents the response from flag submission
type FlagResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewClient creates a new HTB API client
func NewClient(apiToken string) *Client {
	return &Client{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := BaseURL + endpoint
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HTB-Tool/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ListMachines fetches all available machines
func (c *Client) ListMachines() ([]Machine, error) {
	data, err := c.doRequest("GET", "/machine/list", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Info []Machine `json:"info"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse machines: %w", err)
	}

	return response.Info, nil
}

// ListActiveMachines fetches only active machines
func (c *Client) ListActiveMachines() ([]Machine, error) {
	machines, err := c.ListMachines()
	if err != nil {
		return nil, err
	}

	var active []Machine
	for _, m := range machines {
		if m.Active {
			active = append(active, m)
		}
	}
	return active, nil
}

// ListRetiredMachines fetches only retired machines
func (c *Client) ListRetiredMachines() ([]Machine, error) {
	machines, err := c.ListMachines()
	if err != nil {
		return nil, err
	}

	var retired []Machine
	for _, m := range machines {
		if m.Retired {
			retired = append(retired, m)
		}
	}
	return retired, nil
}

// SpawnMachine spawns a machine instance
func (c *Client) SpawnMachine(machineID int) error {
	_, err := c.doRequest("POST", fmt.Sprintf("/machine/play/%d", machineID), nil)
	return err
}

// TerminateMachine terminates a running machine
func (c *Client) TerminateMachine(machineID int) error {
	_, err := c.doRequest("POST", fmt.Sprintf("/machine/stop/%d", machineID), nil)
	return err
}

// ResetMachine resets a machine
func (c *Client) ResetMachine(machineID int) error {
	_, err := c.doRequest("POST", fmt.Sprintf("/machine/reset/%d", machineID), nil)
	return err
}

// GetMachine fetches a specific machine by ID
func (c *Client) GetMachine(machineID int) (*Machine, error) {
	data, err := c.doRequest("GET", fmt.Sprintf("/machine/profile/%d", machineID), nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Info Machine `json:"info"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse machine: %w", err)
	}

	return &response.Info, nil
}

// ListChallenges fetches all challenges
func (c *Client) ListChallenges() ([]Challenge, error) {
	data, err := c.doRequest("GET", "/challenge/list", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Challenges []Challenge `json:"challenges"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse challenges: %w", err)
	}

	return response.Challenges, nil
}

// SubmitMachineFlag submits a flag for a machine
func (c *Client) SubmitMachineFlag(machineID int, flag string, difficulty int) (*FlagResponse, error) {
	submission := FlagSubmission{
		Flag:       flag,
		Difficulty: difficulty,
		ID:         machineID,
	}

	data, err := c.doRequest("POST", "/machine/own", submission)
	if err != nil {
		return nil, err
	}

	var response FlagResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// SubmitChallengeFlag submits a flag for a challenge
func (c *Client) SubmitChallengeFlag(challengeID int, flag string) (*FlagResponse, error) {
	submission := map[string]interface{}{
		"flag":        flag,
		"challenge_id": challengeID,
	}

	data, err := c.doRequest("POST", "/challenge/own", submission)
	if err != nil {
		return nil, err
	}

	var response FlagResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// DownloadVPN downloads a VPN config file
func (c *Client) DownloadVPN(serverID int, protocol string) ([]byte, error) {
	endpoint := fmt.Sprintf("/access/ovpnfile/%d/%s", serverID, protocol)
	return c.doRequest("GET", endpoint, nil)
}

// ListVPNServers fetches available VPN servers
func (c *Client) ListVPNServers() ([]VPNServer, error) {
	data, err := c.doRequest("GET", "/access/servers", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Servers []VPNServer `json:"servers"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse VPN servers: %w", err)
	}

	return response.Servers, nil
}
