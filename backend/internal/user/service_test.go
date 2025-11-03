/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package user

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

func TestValidateUserAndUniquenessReturnsInternalErrorWhenSchemaValidationFails(t *testing.T) {
	mockSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	mockSchemaService.
		On("ValidateUser", "employee", mock.Anything).
		Return(false, &serviceerror.ServiceError{
			Code:  "USRS-5000",
			Type:  serviceerror.ServerErrorType,
			Error: "schema validation failed",
		}).
		Once()

	service := &userService{
		userSchemaService: mockSchemaService,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

	err := service.validateUserAndUniqueness("employee", []byte(`{"email":"employee@example.com"}`), logger)

	require.NotNil(t, err)
	require.Equal(t, ErrorInternalServerError.Code, err.Code)

	mockSchemaService.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
}

func TestValidateUserAndUniquenessReturnsUserSchemaNotFoundWhenSchemaMissing(t *testing.T) {
	mockSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	mockSchemaService.
		On("ValidateUser", "employee", mock.Anything).
		Return(false, &userschema.ErrorUserSchemaNotFound).
		Once()

	service := &userService{
		userSchemaService: mockSchemaService,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

	err := service.validateUserAndUniqueness("employee", []byte(`{}`), logger)

	require.NotNil(t, err)
	require.Equal(t, ErrorUserSchemaNotFound, *err)

	mockSchemaService.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
}

func TestValidateUserAndUniquenessReturnsInternalErrorWhenSchemaLookupFails(t *testing.T) {
	mockSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	mockSchemaService.
		On("ValidateUser", "employee", mock.Anything).
		Return(false, &serviceerror.ServiceError{
			Code:  "USRS-5000",
			Type:  serviceerror.ServerErrorType,
			Error: "unexpected error",
		}).
		Once()

	service := &userService{
		userSchemaService: mockSchemaService,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

	err := service.validateUserAndUniqueness("employee", []byte(`{}`), logger)

	require.NotNil(t, err)
	require.Equal(t, ErrorInternalServerError, *err)

	mockSchemaService.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
}

func TestValidateUserAndUniquenessReturnsSchemaValidationFailed(t *testing.T) {
	mockSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	mockSchemaService.
		On("ValidateUser", "employee", mock.Anything).
		Return(false, nil).
		Once()

	service := &userService{
		userSchemaService: mockSchemaService,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

	err := service.validateUserAndUniqueness("employee", []byte(`{"email":"employee@example.com"}`), logger)

	require.NotNil(t, err)
	require.Equal(t, ErrorSchemaValidationFailed, *err)

	mockSchemaService.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
}

func TestValidateUserAndUniquenessReturnsInternalErrorWhenUniquenessValidationFails(t *testing.T) {
	mockSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	mockSchemaService.
		On("ValidateUser", "employee", mock.Anything).
		Return(true, nil).
		Once()

	mockSchemaService.
		On("ValidateUserUniqueness", "employee", mock.Anything, mock.Anything).
		Return(false, &serviceerror.ServiceError{
			Code:  "USRS-5000",
			Type:  serviceerror.ServerErrorType,
			Error: "validation failed",
		}).
		Once()

	service := &userService{
		userSchemaService: mockSchemaService,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

	err := service.validateUserAndUniqueness("employee", []byte(`{"email":"employee@example.com"}`), logger)

	require.NotNil(t, err)
	require.Equal(t, ErrorInternalServerError, *err)
}

func TestValidateUserAndUniquenessReturnsUserSchemaNotFoundWhenUniquenessSchemaMissing(t *testing.T) {
	mockSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	mockSchemaService.
		On("ValidateUser", "employee", mock.Anything).
		Return(true, nil).
		Once()

	mockSchemaService.
		On("ValidateUserUniqueness", "employee", mock.Anything, mock.Anything).
		Return(false, &userschema.ErrorUserSchemaNotFound).
		Once()

	service := &userService{
		userSchemaService: mockSchemaService,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

	err := service.validateUserAndUniqueness("employee", []byte(`{"email":"employee@example.com"}`), logger)

	require.NotNil(t, err)
	require.Equal(t, ErrorUserSchemaNotFound, *err)
}

func TestValidateUserAndUniquenessReturnsAttributeConflictWhenUniquenessCheckFails(t *testing.T) {
	mockSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	existingUserID := "user-123"
	userStoreMock := newUserStoreInterfaceMock(t)
	userStoreMock.
		On("IdentifyUser", mock.AnythingOfType("map[string]interface {}")).
		Return(&existingUserID, nil).
		Once()

	mockSchemaService.
		On("ValidateUser", "employee", mock.Anything).
		Return(true, nil).
		Once()

	mockSchemaService.
		On("ValidateUserUniqueness", "employee", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			identify := args.Get(2).(func(map[string]interface{}) (*string, error))

			id, err := identify(map[string]interface{}{"email": "employee@example.com"})
			require.NoError(t, err)
			require.NotNil(t, id)
			require.Equal(t, existingUserID, *id)
		}).
		Return(false, nil).
		Once()

	service := &userService{
		userSchemaService: mockSchemaService,
		userStore:         userStoreMock,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

	err := service.validateUserAndUniqueness("employee", []byte(`{"email":"employee@example.com"}`), logger)

	require.NotNil(t, err)
	require.Equal(t, ErrorAttributeConflict, *err)
}

func TestValidateUserAndUniquenessReturnsNilWhenValidationSucceeds(t *testing.T) {
	mockSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	userStoreMock := newUserStoreInterfaceMock(t)
	userStoreMock.
		On("IdentifyUser", mock.AnythingOfType("map[string]interface {}")).
		Return((*string)(nil), nil).
		Once()

	mockSchemaService.
		On("ValidateUser", "employee", mock.Anything).
		Return(true, nil).
		Once()

	mockSchemaService.
		On("ValidateUserUniqueness", "employee", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			identify := args.Get(2).(func(map[string]interface{}) (*string, error))

			id, err := identify(map[string]interface{}{"email": "employee@example.com"})
			require.NoError(t, err)
			require.Nil(t, id)
		}).
		Return(true, nil).
		Once()

	service := &userService{
		userSchemaService: mockSchemaService,
		userStore:         userStoreMock,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

	err := service.validateUserAndUniqueness("employee", []byte(`{"email":"employee@example.com"}`), logger)

	require.Nil(t, err)
}

func TestValidateUserAndUniquenessReturnsInternalErrorWhenIdentifyFails(t *testing.T) {
	mockSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	userStoreMock := newUserStoreInterfaceMock(t)
	userStoreMock.
		On("IdentifyUser", mock.AnythingOfType("map[string]interface {}")).
		Return((*string)(nil), errors.New("store failure")).
		Once()

	mockSchemaService.
		On("ValidateUser", "employee", mock.Anything).
		Return(true, nil).
		Once()

	mockSchemaService.
		On("ValidateUserUniqueness", "employee", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			identify := args.Get(2).(func(map[string]interface{}) (*string, error))

			id, err := identify(map[string]interface{}{"email": "employee@example.com"})
			require.Error(t, err)
			require.Nil(t, id)
		}).
		Return(false, &serviceerror.ServiceError{
			Code:  "USRS-5000",
			Type:  serviceerror.ServerErrorType,
			Error: "validation failed",
		}).
		Once()

	service := &userService{
		userSchemaService: mockSchemaService,
		userStore:         userStoreMock,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

	err := service.validateUserAndUniqueness("employee", []byte(`{"email":"employee@example.com"}`), logger)

	require.NotNil(t, err)
	require.Equal(t, ErrorInternalServerError, *err)
}
