package plank

import (
	"strings"
)

// User is returned by Fiat's /authorize endpoint.
type User struct {
	Name         string          `json:"name"`
	Admin        bool            `json:"admin"`
	Accounts     []Authorization `json:"accounts"`
	Applications []Authorization `json:"applications"`
}

// Authorization describes permissinos for an account or application.
type Authorization struct {
	Name string `json:"name"`
	// Authorizations can be 'READ' 'WRITE'
	Authorizations []string `json:"authorizations"`
}

// IsAdmin returns true if the user has admin permissions
func (u *User) IsAdmin() bool {
	return u.Admin == true
}

// HasAppWriteAccess returns true if user has write access to given app.
func (u *User) HasAppWriteAccess(app string) bool {
	for _, a := range u.Applications {
		if a.Name != app {
			continue
		}
		for _, x := range a.Authorizations {
			if strings.ToLower(x) == "write" {
				return true
			}
		}
	}
	return false
}

// Admin returns whether or not a user is an admin.
func (c *Client) IsAdmin(username string) (bool, error) {
	u, err := c.GetUser(username)
	if err != nil {
		return false, err
	}
	return u.IsAdmin(), nil
}

// HasAppWriteAccess returns whether or not a user can write pipelines/configs/etc. for an app.
func (c *Client) HasAppWriteAccess(username, app string) (bool, error) {
	u, err := c.GetUser(username)
	if err != nil {
		return false, err
	}
	return u.HasAppWriteAccess(app), nil
}

// GetUser gets a user by name.
func (c *Client) GetUser(name string) (*User, error) {
	var u User
	if err := c.Get(c.URLs["fiat"]+"/authorize/"+name, &u); err != nil {
		return nil, err
	}
	return &u, nil
}
