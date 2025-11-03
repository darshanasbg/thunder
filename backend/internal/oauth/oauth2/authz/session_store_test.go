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

package authz

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
)

type SessionDataStoreTestSuite struct {
	suite.Suite
	store sessionDataStoreInterface
}

func TestSessionDataStoreSuite(t *testing.T) {
	suite.Run(t, new(SessionDataStoreTestSuite))
}

func (suite *SessionDataStoreTestSuite) SetupTest() {
	suite.store = newSessionDataStore()
	suite.store.ClearSessionStore()
}

func (suite *SessionDataStoreTestSuite) TearDownTest() {
	if suite.store != nil {
		suite.store.ClearSessionStore()
	}
}

func (suite *SessionDataStoreTestSuite) TestGetSessionDataStore() {
	store := newSessionDataStore()
	assert.NotNil(suite.T(), store)
	assert.Implements(suite.T(), (*sessionDataStoreInterface)(nil), store)
}

func (suite *SessionDataStoreTestSuite) TestAddSession() {
	sessionData := SessionData{
		OAuthParameters: model.OAuthParameters{
			ClientID:         "test-client",
			RedirectURI:      "https://example.com/callback",
			ResponseType:     "code",
			StandardScopes:   []string{"openid"},
			PermissionScopes: []string{"read", "write"},
			State:            "test-state",
		},
		AuthTime: time.Now(),
	}

	key := suite.store.AddSession(sessionData)
	found, retrievedData := suite.store.GetSession(key)
	assert.True(suite.T(), found)
	assert.Equal(suite.T(), sessionData.OAuthParameters.ClientID, retrievedData.OAuthParameters.ClientID)
	assert.Equal(suite.T(), sessionData.AuthTime, retrievedData.AuthTime)
}

func (suite *SessionDataStoreTestSuite) TestGetSession() {
	sessionData := SessionData{
		OAuthParameters: model.OAuthParameters{
			ClientID: "test-client",
			State:    "test-state",
		},
		AuthTime: time.Now(),
	}

	key := suite.store.AddSession(sessionData)
	found, retrievedData := suite.store.GetSession(key)
	assert.True(suite.T(), found)
	assert.Equal(suite.T(), sessionData.OAuthParameters.ClientID, retrievedData.OAuthParameters.ClientID)
	assert.Equal(suite.T(), sessionData.OAuthParameters.State, retrievedData.OAuthParameters.State)
}

func (suite *SessionDataStoreTestSuite) TestGetSessionNotFound() {
	found, _ := suite.store.GetSession("non-existent-key")
	assert.False(suite.T(), found)
}

func (suite *SessionDataStoreTestSuite) TestGetSessionWithEmptyKey() {
	found, _ := suite.store.GetSession("")
	assert.False(suite.T(), found)
}

func (suite *SessionDataStoreTestSuite) TestClearSession() {
	sessionData := SessionData{
		OAuthParameters: model.OAuthParameters{
			ClientID: "test-client",
		},
	}

	key := suite.store.AddSession(sessionData)

	found, _ := suite.store.GetSession(key)
	assert.True(suite.T(), found)

	suite.store.ClearSession(key)
	found, _ = suite.store.GetSession(key)
	assert.False(suite.T(), found)
}

func (suite *SessionDataStoreTestSuite) TestClearSessionWithEmptyKey() {
	suite.store.ClearSession("")
}

func (suite *SessionDataStoreTestSuite) TestClearSessionStore() {
	clientIDs := []string{"client1", "client2", "client3"}
	keys := make([]string, 0, len(clientIDs))

	for _, clientID := range clientIDs {
		keys = append(keys, suite.store.AddSession(SessionData{
			OAuthParameters: model.OAuthParameters{
				ClientID: clientID,
			},
		}))
	}

	for _, key := range keys {
		found, _ := suite.store.GetSession(key)
		assert.True(suite.T(), found)
	}

	suite.store.ClearSessionStore()
	for _, key := range keys {
		found, _ := suite.store.GetSession(key)
		assert.False(suite.T(), found)
	}
}

func (suite *SessionDataStoreTestSuite) TestSessionExpiry() {
	sessionData := SessionData{
		OAuthParameters: model.OAuthParameters{
			ClientID: "test-client",
		},
		AuthTime: time.Now(),
	}

	key := suite.store.AddSession(sessionData)

	found, _ := suite.store.GetSession(key)
	assert.True(suite.T(), found)
}

func (suite *SessionDataStoreTestSuite) TestConcurrentAccess() {
	numGoroutines := 100
	var wg sync.WaitGroup
	keys := make([]string, numGoroutines)
	var keysMutex sync.Mutex

	// Test concurrent AddSession operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			sessionData := SessionData{
				OAuthParameters: model.OAuthParameters{
					ClientID: "test-client-" + string(rune('0'+index%10)),
					State:    "state-" + string(rune('0'+index%10)),
				},
				AuthTime: time.Now(),
			}
			key := suite.store.AddSession(sessionData)

			keysMutex.Lock()
			keys[index] = key
			keysMutex.Unlock()
		}(i)
	}
	wg.Wait()

	// Verify all keys are unique
	keyMap := make(map[string]bool)
	for _, key := range keys {
		assert.NotEmpty(suite.T(), key, "Generated key should not be empty")
		assert.False(suite.T(), keyMap[key], "Keys should be unique, found duplicate: "+key)
		keyMap[key] = true
	}
	assert.Equal(suite.T(), numGoroutines, len(keyMap), "All keys should be unique")

	// Test concurrent GetSession operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			key := keys[index]
			found, retrievedData := suite.store.GetSession(key)
			assert.True(suite.T(), found, "Session should be found for key: "+key)
			assert.NotEmpty(suite.T(), retrievedData.OAuthParameters.ClientID)
		}(i)
	}
	wg.Wait()

	// Test concurrent ClearSession operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			key := keys[index]
			suite.store.ClearSession(key)
		}(i)
	}
	wg.Wait()

	// Verify all sessions are cleared
	for i, key := range keys {
		found, _ := suite.store.GetSession(key)
		assert.False(suite.T(), found, "Session should be cleared for key at index %d: %s", i, key)
	}
}
