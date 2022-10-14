package dinghyfile

import (
	"errors"
	"github.com/armory/dinghy/pkg/util"
	"github.com/armory/plank/v4"
)

const WritePermissionsCheckEnabledFlag = "check_user_write_access"

var UserNotFoundError = errors.New("user was not found")
var UserNotAuthorized = errors.New("user not authorized")

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
// It fetches user roles from Fiat and comapres them with application permissions
type FiatPermissionsValidator struct {
	client      util.PlankClient
	application plank.Application
}

func (v *FiatPermissionsValidator) Validate(pusher string) error {
	userRoles, err := v.client.UserRoles(pusher, v.application.Name)
	if err != nil {
		//TODO Check for 404 before returning error below
		return UserNotFoundError
	}

	for _, applicationPermission := range v.application.Permissions.Write {
		for _, role := range userRoles {
			if applicationPermission == role {
				//It's a match! No error to return
				return nil
			}
		}
	}

	return UserNotAuthorized
}

// A GetWritePermissionsValidator is a factory method that produces
// implementation of WritePermissionsValidator based on settings value
func GetWritePermissionsValidator(settings map[string]interface{}, client util.PlankClient, application plank.Application) WritePermissionsValidator {
	val, found := settings[WritePermissionsCheckEnabledFlag]
	if found && val == true {
		return &FiatPermissionsValidator{
			client:      client,
			application: application,
		}
	}
	return &NoOpWritePermissionValidator{}
}
