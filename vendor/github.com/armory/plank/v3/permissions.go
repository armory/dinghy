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
	"fmt"
	"reflect"
	"strings"
	"sync"
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

type FiatRole struct {
	Name string `json:"name"`
	Source string `json:"source"`
}

func (c *Client) UserRoles(username string) ([]string, error) {
	baseUrl := c.URLs["fiat"]
	var roles []FiatRole
	if err := c.Get(baseUrl + "/authorize/" + username + "/roles", &roles); err != nil {
		return nil, err
	}
	var names []string
	for _, r := range roles {
		names = append(names, r.Name)
	}
	return names, nil
}

// ReadPermissable is an interface for things that should be readable
type ReadPermissable interface {
	GetName() string
	GetPermissions() []string
}

type PermissionsEvaluator interface {
	HasReadPermission(user string, rp ReadPermissable) (bool, error)
}

type FiatClient interface {
	UserRoles(username string) ([]string, error)
}

type FiatClientFactory func(opts ...ClientOption) FiatClient

type FiatPermissionEvaluator struct {
	clientOps []ClientOption
	clientFactory FiatClientFactory
	userRoleCache sync.Map
	// orMode is used as a flag for whether permissable objects need all roles or just one
	// see https://github.com/spinnaker/fiat/blob/87a0d2410244f1a906a3220903aa121ab5042b24/fiat-roles/src/main/java/com/netflix/spinnaker/fiat/config/FiatRoleConfig.java#L28
	// and https://github.com/spinnaker/fiat/blob/0d58386152ad78234b2554f5efdf12d26b77d57c/fiat-roles/src/main/java/com/netflix/spinnaker/fiat/providers/BaseServiceAccountResourceProvider.java#L36-L40
	orMode bool
}

func (f *FiatPermissionEvaluator) HasReadPermission(user string, rp ReadPermissable) (bool, error) {
	if isNil(rp) {
		return false, fmt.Errorf("object for permissions check should not be nil")
	}

	uroles, err := f.userRoles(user)
	if err != nil {
		return false, err
	}

	if f.orMode {
		// if we have at least 1 role in common, return true
		for _, p := range rp.GetPermissions() {
			if isContained(p, uroles) {
				return true, nil
			}
		}
		return false, nil
	}

	for _, p := range rp.GetPermissions() {
		// if we detect one role is missing, return false
		if !isContained(p, uroles) {
			return false, nil
		}
	}

	return true, nil
}

func isNil(rp ReadPermissable) bool {
	// rp == nil won't work for interfaces because they have both type and value. rp == nil on a
	// nil pointer will return false, so we also need the second check.
	// reflect.ValueOf(rp).IsNil() will accurately return true, but will panic if called on a struct
	// (ie, rp is not a nil pointer), so we only make that call if rp is a pointer.
	return rp == nil || (reflect.TypeOf(rp).Kind() == reflect.Ptr && reflect.ValueOf(rp).IsNil())
}

func (f *FiatPermissionEvaluator) userRoles(username string) ([]string, error) {
	if v, ok := f.userRoleCache.Load(username); ok {
		return v.([]string), nil
	}

	// hacky workaround to make up for the fact that i can't pass
	// custom headers into client.Get
	// calls to `/authorize/{user}/roles` require that the user
	// in the path matches the X-Spinnaker-User header
	// i'm not a huge fan of recreating this client every time
	opts := append([]ClientOption{WithFiatUser(username)}, f.clientOps...)
	c := f.clientFactory(opts...)
	uroles, err := c.UserRoles(username)
	if err != nil {
		return nil, err
	}
	// cache result for use later
	f.userRoleCache.Store(username, uroles)
	return uroles, nil
}

func isContained(needle string, haystack []string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func NewFiatPermissionEvaluator(opts ...PermissionEvaluatorOpt) *FiatPermissionEvaluator {
	defaultClientFactory := func(opts ...ClientOption) FiatClient {
		return New(opts...)
	}
	fpe := &FiatPermissionEvaluator{
		clientFactory: defaultClientFactory,
		orMode: false,
	}

	for _, opt := range opts {
		opt(fpe)
	}
	return fpe
}

type PermissionEvaluatorOpt func(fpe *FiatPermissionEvaluator)

func WithOrMode(orMode bool) PermissionEvaluatorOpt {
	return func(fpe *FiatPermissionEvaluator) {
		fpe.orMode = orMode
	}
}

func WithClientOptions(opts ...ClientOption) PermissionEvaluatorOpt {
	return func(fpe *FiatPermissionEvaluator) {
		fpe.clientOps = opts
	}
}
