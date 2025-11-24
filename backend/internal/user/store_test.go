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
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/tests/mocks/database/clientmock"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

type UserStoreTestSuite struct {
	suite.Suite
	providerMock *providermock.DBProviderInterfaceMock
	dbClientMock *clientmock.DBClientInterfaceMock
	store        *userStore
}

const storeUserID = "user-1"

func TestUserStoreTestSuite(t *testing.T) {
	suite.Run(t, new(UserStoreTestSuite))
}

func (suite *UserStoreTestSuite) SetupTest() {
	suite.providerMock = providermock.NewDBProviderInterfaceMock(suite.T())
	suite.dbClientMock = clientmock.NewDBClientInterfaceMock(suite.T())
	suite.store = &userStore{dbProvider: suite.providerMock}
}

func (suite *UserStoreTestSuite) expectDBClient() {
	suite.providerMock.
		On("GetUserDBClient").
		Return(suite.dbClientMock, nil).
		Once()
}

func (suite *UserStoreTestSuite) TestUpdateUserCredentials() {
	credentials := []Credential{
		{
			CredentialType: "password",
			StorageType:    "hash",
			StorageAlgo:    hash.SHA256,
			Value:          "hashed",
			Salt:           "salty",
		},
	}
	credJSON, err := json.Marshal(credentials)
	suite.Require().NoError(err)

	tests := []struct {
		name    string
		setup   func()
		wantErr string
		errIs   error
	}{
		{
			name: "success",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Execute", QueryUpdateUserCredentialsByUserID, storeUserID, string(credJSON)).
					Return(int64(1), nil).
					Once()
			},
		},
		{
			name: "user not found",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Execute", QueryUpdateUserCredentialsByUserID, storeUserID, string(credJSON)).
					Return(int64(0), nil).
					Once()
			},
			errIs: ErrUserNotFound,
		},
		{
			name: "execute error",
			setup: func() {
				suite.expectDBClient()
				suite.dbClientMock.
					On("Execute", QueryUpdateUserCredentialsByUserID, storeUserID, string(credJSON)).
					Return(int64(0), errors.New("exec err")).
					Once()
			},
			wantErr: "failed to execute query",
		},
		{
			name: "db client error",
			setup: func() {
				suite.providerMock.
					On("GetUserDBClient").
					Return(nil, errors.New("db err")).
					Once()
			},
			wantErr: "failed to get database client",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			if tc.setup != nil {
				tc.setup()
			}

			err := suite.store.UpdateUserCredentials(storeUserID, credentials)

			if tc.errIs != nil {
				suite.ErrorIs(err, tc.errIs)
				return
			}

			if tc.wantErr != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.NoError(err)
		})
	}
}
