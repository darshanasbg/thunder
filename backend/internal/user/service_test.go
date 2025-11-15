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

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

const testUserType = "employee"

func TestOUStore_ValidateUserAndUniqueness(t *testing.T) {
	type testMocks struct {
		schemaService *userschemamock.UserSchemaServiceInterfaceMock
		userStore     *userStoreInterfaceMock
	}

	payloadWithEmail := []byte(`{"email":"employee@example.com"}`)
	emptyPayload := []byte(`{}`)

	testCases := []struct {
		name    string
		payload []byte
		setup   func(t *testing.T) (*userService, testMocks)
		assert  func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks)
	}{
		{
			name:    "ReturnsInternalErrorWhenSchemaValidationFails",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(false, &serviceerror.ServiceError{
						Code:  "USRS-5000",
						Type:  serviceerror.ServerErrorType,
						Error: "schema validation failed",
					}).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorInternalServerError.Code, err.Code)
				mocks.schemaService.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:    "ReturnsUserSchemaNotFoundWhenSchemaMissing",
			payload: emptyPayload,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(false, &userschema.ErrorUserSchemaNotFound).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorUserSchemaNotFound, *err)
				mocks.schemaService.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:    "ReturnsInternalErrorWhenSchemaLookupFails",
			payload: emptyPayload,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(false, &serviceerror.ServiceError{
						Code:  "USRS-5000",
						Type:  serviceerror.ServerErrorType,
						Error: "unexpected error",
					}).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorInternalServerError, *err)
				mocks.schemaService.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:    "ReturnsSchemaValidationFailed",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(false, nil).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorSchemaValidationFailed, *err)
				mocks.schemaService.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:    "ReturnsInternalErrorWhenUniquenessValidationFails",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
					Return(false, &serviceerror.ServiceError{
						Code:  "USRS-5000",
						Type:  serviceerror.ServerErrorType,
						Error: "validation failed",
					}).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorInternalServerError, *err)
			},
		},
		{
			name:    "ReturnsUserSchemaNotFoundWhenUniquenessSchemaMissing",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
					Return(false, &userschema.ErrorUserSchemaNotFound).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorUserSchemaNotFound, *err)
			},
		},
		{
			name:    "ReturnsAttributeConflictWhenUniquenessCheckFails",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				existingUserID := "user-123"
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userStoreMock := newUserStoreInterfaceMock(t)
				userStoreMock.
					On("IdentifyUser", mock.AnythingOfType("map[string]interface {}")).
					Return(&existingUserID, nil).
					Once()
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						identify := args.Get(2).(func(map[string]interface{}) (*string, error))

						id, err := identify(map[string]interface{}{"email": "employee@example.com"})
						require.NoError(t, err)
						require.NotNil(t, id)
						require.Equal(t, existingUserID, *id)
					}).
					Return(false, nil).
					Once()

				return &userService{
						userSchemaService: schemaMock,
						userStore:         userStoreMock,
					}, testMocks{
						schemaService: schemaMock,
						userStore:     userStoreMock,
					}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorAttributeConflict, *err)
			},
		},
		{
			name:    "ReturnsNilWhenValidationSucceeds",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userStoreMock := newUserStoreInterfaceMock(t)
				userStoreMock.
					On("IdentifyUser", mock.AnythingOfType("map[string]interface {}")).
					Return((*string)(nil), nil).
					Once()
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						identify := args.Get(2).(func(map[string]interface{}) (*string, error))

						id, err := identify(map[string]interface{}{"email": "employee@example.com"})
						require.NoError(t, err)
						require.Nil(t, id)
					}).
					Return(true, nil).
					Once()

				return &userService{
						userSchemaService: schemaMock,
						userStore:         userStoreMock,
					}, testMocks{
						schemaService: schemaMock,
						userStore:     userStoreMock,
					}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.Nil(t, err)
			},
		},
		{
			name:    "ReturnsInternalErrorWhenIdentifyFails",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userStoreMock := newUserStoreInterfaceMock(t)
				userStoreMock.
					On("IdentifyUser", mock.AnythingOfType("map[string]interface {}")).
					Return((*string)(nil), errors.New("store failure")).
					Once()
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
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

				return &userService{
						userSchemaService: schemaMock,
						userStore:         userStoreMock,
					}, testMocks{
						schemaService: schemaMock,
						userStore:     userStoreMock,
					}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorInternalServerError, *err)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			service, mocks := tc.setup(t)
			logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

			err := service.validateUserAndUniqueness(testUserType, tc.payload, logger)
			tc.assert(t, err, mocks)
		})
	}
}

func TestOUStore_ValidateOrganizationUnitForUserType(t *testing.T) {
	type testMocks struct {
		ouService         *oumock.OrganizationUnitServiceInterfaceMock
		userSchemaService *userschemamock.UserSchemaServiceInterfaceMock
	}

	testCases := []struct {
		name        string
		userType    string
		ouID        string
		setup       func(t *testing.T) (*userService, testMocks)
		expectedErr *serviceerror.ServiceError
	}{
		{
			name:     "ReturnsErrorWhenIDEmpty",
			userType: testUserType,
			ouID:     "",
			setup: func(t *testing.T) (*userService, testMocks) {
				return &userService{}, testMocks{}
			},
			expectedErr: &ErrorInvalidOrganizationUnitID,
		},
		{
			name:     "ReturnsErrorWhenIDInvalid",
			userType: testUserType,
			ouID:     "invalid-id",
			setup: func(t *testing.T) (*userService, testMocks) {
				return &userService{}, testMocks{}
			},
			expectedErr: &ErrorInvalidOrganizationUnitID,
		},
		{
			name:     "ReturnsErrorWhenOrganizationUnitMissing",
			userType: testUserType,
			ouID:     "4d8b40d6-3a17-4c19-9a94-5866df9b6bf5",
			setup: func(t *testing.T) (*userService, testMocks) {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"4d8b40d6-3a17-4c19-9a94-5866df9b6bf5").Return(false, (*serviceerror.ServiceError)(nil)).Once()

				return &userService{
						ouService: ouServiceMock,
					}, testMocks{
						ouService: ouServiceMock,
					}
			},
			expectedErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name:     "HandlesClientErrorWhenOrganizationUnitMissing",
			userType: testUserType,
			ouID:     "6c8f5afd-8884-4ea0-a317-3d8579346d86",
			setup: func(t *testing.T) (*userService, testMocks) {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"6c8f5afd-8884-4ea0-a317-3d8579346d86").Return(false, &serviceerror.ServiceError{
					Type: serviceerror.ClientErrorType,
					Code: oupkg.ErrorOrganizationUnitNotFound.Code,
				}).Once()

				return &userService{
						ouService: ouServiceMock,
					}, testMocks{
						ouService: ouServiceMock,
					}
			},
			expectedErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name:     "HandlesClientErrorWhenOrganizationUnitIDInvalid",
			userType: testUserType,
			ouID:     "8d0c2f4e-8bb1-40bc-a0e1-ca5c4aacff63",
			setup: func(t *testing.T) (*userService, testMocks) {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"8d0c2f4e-8bb1-40bc-a0e1-ca5c4aacff63").Return(false, &serviceerror.ServiceError{
					Type: serviceerror.ClientErrorType,
					Code: oupkg.ErrorInvalidRequestFormat.Code,
				}).Once()

				return &userService{
						ouService: ouServiceMock,
					}, testMocks{
						ouService: ouServiceMock,
					}
			},
			expectedErr: &ErrorInvalidOrganizationUnitID,
		},
		{
			name:     "ReturnsMismatchWhenSchemaDoesNotMatchOU",
			userType: testUserType,
			ouID:     "f4e7c7b2-0b11-46a4-83be-4b43a7f69c7e",
			setup: func(t *testing.T) (*userService, testMocks) {
				parentOU := "a88cbecc-53a3-4c3e-958f-7ee4bf2d7a28"
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"f4e7c7b2-0b11-46a4-83be-4b43a7f69c7e").Return(true, (*serviceerror.ServiceError)(nil)).Once()
				ouServiceMock.
					On("IsParent", parentOU, "f4e7c7b2-0b11-46a4-83be-4b43a7f69c7e").
					Return(false, (*serviceerror.ServiceError)(nil)).
					Once()

				userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userSchemaMock.
					On("GetUserSchemaByName", testUserType).
					Return(&userschema.UserSchema{
						OrganizationUnitID: parentOU,
					}, (*serviceerror.ServiceError)(nil)).
					Once()

				return &userService{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}, testMocks{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}
			},
			expectedErr: &ErrorOrganizationUnitMismatch,
		},
		{
			name:     "AllowsChildOrganizationUnit",
			userType: testUserType,
			ouID:     "1b5c7208-0d6f-4d5d-8fb9-6e8573549533",
			setup: func(t *testing.T) (*userService, testMocks) {
				parentOU := "c7e99c3b-e563-4c47-981f-1f7f755c8c68"
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"1b5c7208-0d6f-4d5d-8fb9-6e8573549533").Return(true, (*serviceerror.ServiceError)(nil)).Once()
				ouServiceMock.On("IsParent", parentOU,
					"1b5c7208-0d6f-4d5d-8fb9-6e8573549533").Return(true, (*serviceerror.ServiceError)(nil)).Once()

				userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userSchemaMock.
					On("GetUserSchemaByName", testUserType).
					Return(&userschema.UserSchema{
						OrganizationUnitID: parentOU,
					}, (*serviceerror.ServiceError)(nil)).
					Once()

				return &userService{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}, testMocks{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}
			},
			expectedErr: nil,
		},
		{
			name:     "HandlesParentCheckErrorsOrganizationUnitNotFound",
			userType: testUserType,
			ouID:     "d9e12416-58d3-4c17-a4e4-cc4d96122598",
			setup: func(t *testing.T) (*userService, testMocks) {
				parentOU := "0a08d914-d223-48c2-8939-55d719739a17"
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"d9e12416-58d3-4c17-a4e4-cc4d96122598").Return(true, (*serviceerror.ServiceError)(nil)).Once()
				ouServiceMock.On("IsParent", parentOU,
					"d9e12416-58d3-4c17-a4e4-cc4d96122598").Return(false, &serviceerror.ServiceError{
					Code: oupkg.ErrorOrganizationUnitNotFound.Code,
				}).Once()

				userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userSchemaMock.
					On("GetUserSchemaByName", testUserType).
					Return(&userschema.UserSchema{
						OrganizationUnitID: parentOU,
					}, (*serviceerror.ServiceError)(nil)).
					Once()

				return &userService{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}, testMocks{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}
			},
			expectedErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name:     "HandlesParentCheckErrorsInternalServerError",
			userType: testUserType,
			ouID:     "d9e12416-58d3-4c17-a4e4-cc4d96122598",
			setup: func(t *testing.T) (*userService, testMocks) {
				parentOU := "0a08d914-d223-48c2-8939-55d719739a17"
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"d9e12416-58d3-4c17-a4e4-cc4d96122598").Return(true, (*serviceerror.ServiceError)(nil)).Once()
				ouServiceMock.On("IsParent", parentOU,
					"d9e12416-58d3-4c17-a4e4-cc4d96122598").Return(false, &serviceerror.ServiceError{
					Code: oupkg.ErrorInternalServerError.Code,
				}).Once()

				userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userSchemaMock.
					On("GetUserSchemaByName", testUserType).
					Return(&userschema.UserSchema{
						OrganizationUnitID: parentOU,
					}, (*serviceerror.ServiceError)(nil)).
					Once()

				return &userService{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}, testMocks{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}
			},
			expectedErr: &ErrorInternalServerError,
		},
		{
			name:     "ReturnsNilWhenValid",
			userType: testUserType,
			ouID:     "e5c3aa8a-d7df-46f8-9f3f-bb3245c95d7c",
			setup: func(t *testing.T) (*userService, testMocks) {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"e5c3aa8a-d7df-46f8-9f3f-bb3245c95d7c").Return(true, (*serviceerror.ServiceError)(nil)).Once()

				userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userSchemaMock.
					On("GetUserSchemaByName", testUserType).
					Return(&userschema.UserSchema{
						OrganizationUnitID: "e5c3aa8a-d7df-46f8-9f3f-bb3245c95d7c",
					}, (*serviceerror.ServiceError)(nil)).
					Once()

				return &userService{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}, testMocks{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			service, _ := tc.setup(t)
			logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

			err := service.validateOrganizationUnitForUserType(tc.userType, tc.ouID, logger)
			if tc.expectedErr == nil {
				require.Nil(t, err)
				return
			}

			require.NotNil(t, err)
			require.Equal(t, *tc.expectedErr, *err)
		})
	}
}

func TestUserService_GetUsersByPath_HandlesOUServiceErrors(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(t *testing.T) *userService
		expectedErr *serviceerror.ServiceError
	}{
		{
			name: "ReturnsInvalidHandlePathWhenResolverFails",
			setup: func(t *testing.T) *userService {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{
						Type: serviceerror.ClientErrorType,
						Code: oupkg.ErrorInvalidHandlePath.Code,
					}).
					Once()

				return &userService{
					ouService: ouServiceMock,
				}
			},
			expectedErr: &ErrorInvalidHandlePath,
		},
		{
			name: "ReturnsInvalidLimitWhenListingUsersFails",
			setup: func(t *testing.T) *userService {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{ID: "ou-id"}, (*serviceerror.ServiceError)(nil)).
					Once()
				ouServiceMock.
					On("GetOrganizationUnitUsers", "ou-id", 10, 0).
					Return((*oupkg.UserListResponse)(nil), &serviceerror.ServiceError{
						Type: serviceerror.ClientErrorType,
						Code: oupkg.ErrorInvalidLimit.Code,
					}).
					Once()

				return &userService{
					ouService: ouServiceMock,
				}
			},
			expectedErr: &ErrorInvalidLimit,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			service := tc.setup(t)

			resp, err := service.GetUsersByPath("root", 10, 0, nil)
			require.Nil(t, resp)
			require.NotNil(t, err)
			require.Equal(t, *tc.expectedErr, *err)
		})
	}
}

func TestUserService_CreateUserByPath_HandlesOUServiceErrors(t *testing.T) {
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	ouServiceMock.
		On("GetOrganizationUnitByPath", "root/engineering").
		Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{
			Type: serviceerror.ClientErrorType,
			Code: oupkg.ErrorInvalidHandlePath.Code,
		}).
		Once()

	service := &userService{
		ouService: ouServiceMock,
	}

	resp, err := service.CreateUserByPath("root/engineering", CreateUserByPathRequest{
		Type: testUserType,
	})
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidHandlePath, *err)
}
