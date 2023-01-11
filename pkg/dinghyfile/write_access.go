/*
* Copyright 2022 Armory, Inc.

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

*    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package dinghyfile

import (
	"errors"
	"github.com/armory/dinghy/pkg/util"
	"github.com/armory/plank/v4"
	log "github.com/sirupsen/logrus"
)

var UserNotFoundError = errors.New("user was not found")
var UserNotAuthorized = errors.New("user not authorized")
var UserNameEmpty = errors.New("pusher is empty string")

// A WritePermissionsValidator interface is used to determine
// whether given user has write permissions to application
type WritePermissionsValidator interface {
	Validate(pusherName string) error
}

// A NoOpWritePermissionValidator interface implements WritePermissionsValidator
// and doesn't run any validations - always returns nil
type NoOpWritePermissionValidator struct {
}

func (w NoOpWritePermissionValidator) Validate(string) error {
	return nil
}

// A FiatPermissionsValidator interface implements WritePermissionsValidator
// It fetches user roles from Fiat and compares them with application permissions
type FiatPermissionsValidator struct {
	client      util.PlankClient
	application plank.Application
}

func (v FiatPermissionsValidator) Validate(pusher string) error {
	if pusher == "" {
		log.Errorf("Got empty string as pusher. Either that attribute is mising in a webhook, or there's problem with mapping")
		return UserNameEmpty
	}
	userRoles, err := v.client.UserRoles(pusher, "")
	if err != nil {
		if failedResponse, ok := err.(*plank.FailedResponse); ok {
			if failedResponse.StatusCode == 404 {
				log.Errorf("Either user %s or their roles were not found", pusher)
				return UserNotFoundError
			}
		}
		log.Errorf("Failed to fetch %s's roles from Fiat, because of: %s", pusher, err)
		return err
	}

	log.Infof("%s's write roles were found: %s", pusher, v.application.Permissions.Write)

	for _, applicationPermission := range v.application.Permissions.Write {
		for _, role := range userRoles {
			if applicationPermission == role {
				//It's a match! No error to return
				log.Infof("%s has write permissions to application %s", pusher, v.application.Name)
				return nil
			}
		}
	}

	log.Errorf("%s doesn't have write permissions to application %s", pusher, v.application.Name)
	return UserNotAuthorized
}

// A GetWritePermissionsValidator is a factory method that produces
// implementation of WritePermissionsValidator based on settings value
func GetWritePermissionsValidator(userWritePermissionCheck bool, client util.PlankClient, application plank.Application) WritePermissionsValidator {
	if userWritePermissionCheck {
		return &FiatPermissionsValidator{
			client:      client,
			application: application,
		}
	}
	return &NoOpWritePermissionValidator{}
}

type WritePermissionUserFilter struct {
	usersToIgnore []string
}

// ShouldIgnore function tells whether we should ignore user when checking for write permissions to an application
// Returns true when user is in the ignore list, and we should skip checking for write permissions
// Returns false when user is not in the list and should check theirs write permissions
func (filter *WritePermissionUserFilter) ShouldIgnore(username string) bool {
	for _, name := range filter.usersToIgnore {
		if username == name {
			return true
		}
	}
	return false
}
