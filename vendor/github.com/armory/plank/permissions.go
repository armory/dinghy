/*
 * Copyright 2019 Armory, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package plank

import (
	"strings"
)

// User is returned by Fiat's /authorize endpoint.
type User struct {
	Name         string          `json:"name" yaml:"name" hcl:"name"`
	Admin        bool            `json:"admin" yaml:"admin" hcl:"admin"`
	Accounts     []Authorization `json:"accounts" yaml:"accounts" hcl:"accounts"`
	Applications []Authorization `json:"applications" yaml:"applications" hcl:"applications"`
}

// Authorization describes permissinos for an account or application.
type Authorization struct {
	Name string `json:"name" yaml:"name" hcl:"name"`
	// Authorizations can be 'READ' 'WRITE'
	Authorizations []string `json:"authorizations" yaml:"authorizations" hcl:"authorizations"`
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

// ResyncFiat calls to Fiat to tell it to resync its cache of applications
// and permissions.  This uses an endpoint specific to Armory's distribution
// of Fiat; if ArmoryEndpoints is not set (it's false by default) this is
// a no-op.
func (c *Client) ResyncFiat() error {
	if !c.ArmoryEndpoints {
		// This only works if Armory endpoints are available
		return nil
	}

	if c.FiatUser == "" {
		// No FiatUser, no Fiat, do nothing.
		return nil
	}

	var unused interface{}
	return c.Post(c.URLs["fiat"]+"/forceRefresh/all", ApplicationJson, nil, &unused)
}
