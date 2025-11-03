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

package userschema

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateUserReturnsTrueWhenValidationPasses(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", "employee").
		Return(UserSchema{
			Name:   "employee",
			Schema: json.RawMessage(`{"email":{"type":"string","required":true}}`),
		}, nil).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
	}

	ok, svcErr := service.ValidateUser("employee", json.RawMessage(`{"email":"employee@example.com"}`))

	require.True(t, ok)
	require.Nil(t, svcErr)
}

func TestValidateUserReturnsInternalErrorWhenSchemaLoadFails(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", "employee").
		Return(UserSchema{}, errors.New("db failure")).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
	}

	ok, svcErr := service.ValidateUser("employee", json.RawMessage(`{}`))

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestValidateUserUniquenessReturnsTrueWhenNoConflicts(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", "employee").
		Return(UserSchema{
			Name:   "employee",
			Schema: json.RawMessage(`{"email":{"type":"string","unique":true}}`),
		}, nil).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
	}

	ok, svcErr := service.ValidateUserUniqueness(
		"employee",
		json.RawMessage(`{"email":"unique@example.com"}`),
		func(filters map[string]interface{}) (*string, error) {
			require.Equal(t, map[string]interface{}{"email": "unique@example.com"}, filters)
			return nil, nil
		},
	)

	require.True(t, ok)
	require.Nil(t, svcErr)
}

func TestValidateUserReturnsSchemaNotFoundWhenSchemaMissing(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", "employee").
		Return(UserSchema{}, ErrUserSchemaNotFound).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
	}

	ok, svcErr := service.ValidateUser("employee", json.RawMessage(`{"email":"employee@example.com"}`))

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserSchemaNotFound, *svcErr)
}

func TestValidateUserUniquenessReturnsSchemaNotFoundWhenSchemaMissing(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", "employee").
		Return(UserSchema{}, ErrUserSchemaNotFound).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
	}

	ok, svcErr := service.ValidateUserUniqueness(
		"employee",
		json.RawMessage(`{}`),
		func(map[string]interface{}) (*string, error) { return nil, nil },
	)

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserSchemaNotFound, *svcErr)
}

func TestValidateUserUniquenessReturnsInternalErrorWhenSchemaLoadFails(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", "employee").
		Return(UserSchema{}, errors.New("db failure")).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
	}

	ok, svcErr := service.ValidateUserUniqueness(
		"employee",
		json.RawMessage(`{}`),
		func(map[string]interface{}) (*string, error) { return nil, nil },
	)

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}
