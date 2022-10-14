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
	"github.com/armory/plank/v4"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	pusher = "test_pusher"
)

func TestNoOpValidatorReturnsNil(t *testing.T) {
	noOp := &NoOpWritePermissionValidator{}
	result := noOp.Validate(pusher)

	assert.Nil(t, result)
}

func TestFiatValidationRunsOnlyWhenFiatIsEnabled(t *testing.T) {
	t.Errorf("Is not yet implemented")
}

func TestValidationPassesWhenUserRolesMatchApplicationWritePermissions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	application := plank.Application{
		Name:        "ApplicationName",
		Permissions: &plank.PermissionsType{Write: []string{"developer", "devops"}},
	}

	mockedPlankClient := NewMockPlankClient(ctrl)

	userRoles := []string{"devops"}

	mockedPlankClient.EXPECT().UserRoles(gomock.Eq(pusher), "").Return(userRoles, nil).Times(1)
	permissionsValidator := FiatPermissionsValidator{
		client:      mockedPlankClient,
		application: application,
	}

	assert.Nil(t, permissionsValidator.Validate(pusher))
}

func TestValidationFailsWhenUserRolesDontMatchApplicationWritePermissions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	application := plank.Application{
		Name:        "ApplicationName",
		Permissions: &plank.PermissionsType{Write: []string{"developer", "devops"}},
	}

	mockedPlankClient := NewMockPlankClient(ctrl)

	userRoles := []string{"qa"}

	mockedPlankClient.EXPECT().UserRoles(gomock.Eq(pusher), "").Return(userRoles, nil).Times(1)
	permissionsValidator := FiatPermissionsValidator{
		client:      mockedPlankClient,
		application: application,
	}

	validationResult := permissionsValidator.Validate(pusher)
	assert.NotNil(t, validationResult)
	assert.Equal(t, validationResult, UserNotAuthorized)
}

func TestValidationFailsWhenUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	application := plank.Application{
		Name:        "ApplicationName",
		Permissions: &plank.PermissionsType{Write: []string{"developer", "devops"}},
	}

	mockedPlankClient := NewMockPlankClient(ctrl)
	response := &plank.FailedResponse{StatusCode: 404}
	mockedPlankClient.EXPECT().UserRoles(gomock.Eq(pusher), "").Return(nil, response).Times(1)
	permissionsValidator := FiatPermissionsValidator{
		client:      mockedPlankClient,
		application: application,
	}

	validationResult := permissionsValidator.Validate(pusher)
	assert.NotNil(t, validationResult)
	assert.Equal(t, validationResult, UserNotFoundError)
}

func TestValidationFailsWhenGenericErrorReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	application := plank.Application{
		Name:        "ApplicationName",
		Permissions: &plank.PermissionsType{Write: []string{"developer", "devops"}},
	}

	mockedPlankClient := NewMockPlankClient(ctrl)
	mockedPlankClient.EXPECT().UserRoles(gomock.Eq(pusher), "").Return(nil, errors.New("some ambiguous error")).Times(1)
	permissionsValidator := FiatPermissionsValidator{
		client:      mockedPlankClient,
		application: application,
	}

	validationResult := permissionsValidator.Validate(pusher)
	assert.NotNil(t, validationResult)
	assert.NotEqual(t, validationResult, UserNotFoundError)
	assert.NotEqual(t, validationResult, UserNotAuthorized)
}
