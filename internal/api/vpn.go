package api

import (
	"fmt"
	"net/http"
	"net/url"
)

// vpnProducts are the product values accepted by /connections/servers (the
// `product` query parameter is required). Pro Labs use a separate per-prolab
// endpoint and are not included here.
var vpnProducts = []string{"labs", "competitive", "starting_point", "fortresses"}

// VPNServer is a single selectable HTB VPN endpoint, flattened out of the
// nested grouping returned by /connections/servers. The fields here describe
// this struct's own shape (the returned slice is plain); see ListVPNServers for
// how it is populated from the nested response.
type VPNServer struct {
	ID             int    `json:"id"`
	FriendlyName   string `json:"friendly_name"`
	Location       string `json:"location"`
	CurrentClients int    `json:"current_clients"`
	Full           bool   `json:"full"`
	// Product is the top-level grouping key from the nested response
	// (for example "lab", "starting_point", "endgames", "fortresses", "prolab").
	Product string `json:"product"`
	// Tier is the second-level grouping key under the product
	// (for example "free" or "vip"); empty when not present.
	Tier string `json:"tier"`
	// Assigned marks the server the user is currently assigned to - the only
	// server whose .ovpn config can be downloaded.
	Assigned bool `json:"-"`
}

// vpnServer mirrors a single server object (schema Server) inside the nested
// /connections/servers response.
type vpnServer struct {
	ID             int    `json:"id"`
	FriendlyName   string `json:"friendly_name"`
	Location       string `json:"location"`
	CurrentClients int    `json:"current_clients"`
	Full           bool   `json:"full"`
}

// vpnServerGroup mirrors schema ServerGroup: a named region group whose
// `servers` map is keyed by server id (as a string).
type vpnServerGroup struct {
	Location string               `json:"location"`
	Name     string               `json:"name"`
	Servers  map[string]vpnServer `json:"servers"`
}

// ListVPNServers fetches all assignable VPN servers from v4
// /connections/servers and flattens the deeply nested
// data.options[product][tier].servers[id] structure into a flat slice.
func (c *Client) ListVPNServers() ([]VPNServer, error) {
	var servers []VPNServer
	var firstErr error
	assignedID := 0

	for _, product := range vpnProducts {
		q := url.Values{}
		q.Set("product", product)

		var resp struct {
			Data struct {
				Assigned struct {
					ID int `json:"id"`
				} `json:"assigned"`
				// options is map[group]map[tier]ServerGroup.
				Options map[string]map[string]vpnServerGroup `json:"options"`
			} `json:"data"`
		}
		if err := c.getJSON("v4", "/connections/servers", q, &resp); err != nil {
			// A product the user can't access (or a transient error) shouldn't
			// abort the others - record it and keep going.
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if resp.Data.Assigned.ID != 0 {
			assignedID = resp.Data.Assigned.ID
		}

		for _, tiers := range resp.Data.Options {
			for tier, group := range tiers {
				for _, srv := range group.Servers {
					servers = append(servers, VPNServer{
						ID:             srv.ID,
						FriendlyName:   srv.FriendlyName,
						Location:       srv.Location,
						CurrentClients: srv.CurrentClients,
						Full:           srv.Full,
						Product:        product,
						Tier:           tier,
					})
				}
			}
		}
	}

	if assignedID != 0 {
		for i := range servers {
			if servers[i].ID == assignedID {
				servers[i].Assigned = true
			}
		}
	}

	if len(servers) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return servers, nil
}

// DownloadVPN downloads the raw OpenVPN (.ovpn) configuration for the given VPN
// server id from v4. When tcp is false the UDP config is fetched from
// /access/ovpnfile/{id}/0; when tcp is true the TCP config is fetched from
// /access/ovpnfile/{id}/0/1.
func (c *Client) DownloadVPN(serverID int, tcp bool) ([]byte, error) {
	path := fmt.Sprintf("/access/ovpnfile/%d/0", serverID)
	if tcp {
		path = fmt.Sprintf("/access/ovpnfile/%d/0/1", serverID)
	}
	return c.getRaw("v4", path)
}

// SwitchVPN switches the user's active VPN server to the given id via POST v4
// /connections/servers/switch/{vpnId}. The id is supplied in the path; no
// request body is required.
func (c *Client) SwitchVPN(serverID int) error {
	path := fmt.Sprintf("/connections/servers/switch/%d", serverID)
	return c.sendJSON(http.MethodPost, "v4", path, nil, nil)
}
