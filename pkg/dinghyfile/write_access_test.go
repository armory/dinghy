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
	"fmt"
	"github.com/armory/plank/v4"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	pusher = "test_pusher"
)

func TestNoOpWritePermissionValidator_NoPermissionsChecked(t *testing.T) {
	noOp := &NoOpWritePermissionValidator{}
	result := noOp.Validate(pusher)

	assert.Nil(t, result)
}

func TestPermissionValidator_CorrectImplementationSelected(t *testing.T) {

	noApp := plank.Application{}
	fiatValidator := GetWritePermissionsValidator(true, nil, noApp)
	_, ok := fiatValidator.(*FiatPermissionsValidator)
	assert.True(t, ok)

	noOpValidator := GetWritePermissionsValidator(false, nil, noApp)
	_, ok = noOpValidator.(*NoOpWritePermissionValidator)
	assert.True(t, ok)
}

func TestFiatPermissionsValidator_NoErrorWhenUserHasWritePermissions(t *testing.T) {
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

func TestFiatPermissionsValidator_UserNotAuthorized(t *testing.T) {
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

func TestFiatPermissionsValidator_UserNotFound(t *testing.T) {
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

func TestFiatPermissionsValidator_ErrorReturnedWhenUnexpectedIssue(t *testing.T) {
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

func TestWritePermissionsUserFilter_ShouldProcess(t *testing.T) {

	tests := []struct {
		name     string
		fields   []string
		arg      string
		expected bool
	}{
		{
			name:     "Should not ignore user if they are not on a list of users to ignore",
			fields:   []string{"jon", "danny"},
			arg:      "matt",
			expected: false,
		},
		{
			name:     "Should ignore user if they are on a list of users to ignore ",
			fields:   []string{"jon", "danny"},
			arg:      "danny",
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &WritePermissionUserFilter{
				usersToIgnore: tt.fields}
			assert.Equal(t, tt.expected, filter.ShouldIgnore(tt.arg), fmt.Sprintf("ShouldIgnore(%v)", tt.arg))
		})
	}
}
