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
 * software distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package userinfo

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
	clientID      = "userinfo_test_client_123"
	clientSecret  = "userinfo_test_secret_123"
	appName       = "UserInfoTestApp"
	redirectURI   = "https://localhost:3000"
)

var (
	testUserSchema = testutils.UserSchema{
		Name: "userinfo-person",
		Schema: map[string]interface{}{
			"username": map[string]interface{}{
				"type": "string",
			},
			"password": map[string]interface{}{
				"type": "string",
			},
			"email": map[string]interface{}{
				"type": "string",
			},
			"firstName": map[string]interface{}{
				"type": "string",
			},
			"lastName": map[string]interface{}{
				"type": "string",
			},
		},
	}
)

type UserInfoTestSuite struct {
	suite.Suite
	applicationID string
	userSchemaID  string
	userID        string
	client        *http.Client
	ouID          string
}

func TestUserInfoTestSuite(t *testing.T) {
	suite.Run(t, new(UserInfoTestSuite))
}

func (ts *UserInfoTestSuite) SetupSuite() {
	// Create HTTP client that skips TLS verification
	ts.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	// Create test organization unit
	ou := testutils.OrganizationUnit{
		Handle:      "userinfo-test-ou",
		Name:        "UserInfo Test OU",
		Description: "Organization unit for UserInfo integration testing",
		Parent:      nil,
	}
	ouID, err := testutils.CreateOrganizationUnit(ou)
	ts.Require().NoError(err, "Failed to create test organization unit")
	ts.ouID = ouID

	// Create user schema
	testUserSchema.OrganizationUnitId = ts.ouID
	schemaID, err := testutils.CreateUserType(testUserSchema)
	ts.Require().NoError(err, "Failed to create test user schema")
	ts.userSchemaID = schemaID

	// Create test user
	ts.userID = ts.createTestUser()

	// Create OAuth application
	ts.applicationID = ts.createTestApplication()
}

func (ts *UserInfoTestSuite) TearDownSuite() {
	// Clean up application
	if ts.applicationID != "" {
		ts.deleteApplication(ts.applicationID)
	}

	// Clean up user
	if ts.userID != "" {
		testutils.DeleteUser(ts.userID)
	}

	// Clean up organization unit
	if ts.ouID != "" {
		testutils.DeleteOrganizationUnit(ts.ouID)
	}

	// Clean up user schema
	if ts.userSchemaID != "" {
		if err := testutils.DeleteUserType(ts.userSchemaID); err != nil {
			ts.T().Logf("Failed to delete user schema during teardown: %v", err)
		}
	}
}

func (ts *UserInfoTestSuite) createTestUser() string {
	attributes := map[string]interface{}{
		"username":  "userinfo_test_user",
		"password":  "SecurePass123!",
		"email":     "userinfo_test@example.com",
		"firstName": "UserInfo",
		"lastName":  "Test",
	}

	attributesJSON, err := json.Marshal(attributes)
	ts.Require().NoError(err, "Failed to marshal user attributes")

	user := testutils.User{
		Type:             "userinfo-person",
		OrganizationUnit: ts.ouID,
		Attributes:       json.RawMessage(attributesJSON),
	}

	userID, err := testutils.CreateUser(user)
	ts.Require().NoError(err, "Failed to create test user")
	ts.T().Logf("Created test user with ID: %s", userID)

	return userID
}

func (ts *UserInfoTestSuite) createTestApplication() string {
	app := map[string]interface{}{
		"name":                         appName,
		"description":                  "Application for UserInfo integration tests",
		"auth_flow_graph_id":           "auth_flow_config_basic",
		"registration_flow_graph_id":   "registration_flow_config_basic",
		"is_registration_flow_enabled": true,
		"inbound_auth_config": []map[string]interface{}{
			{
				"type": "oauth2",
				"config": map[string]interface{}{
					"client_id":     clientID,
					"client_secret": clientSecret,
					"redirect_uris": []string{redirectURI},
					"grant_types": []string{
						"client_credentials",
						"authorization_code",
						"refresh_token",
						"urn:ietf:params:oauth:grant-type:token-exchange",
					},
					"response_types":             []string{"code"},
					"token_endpoint_auth_method": "client_secret_basic",
					"scopes":                     []string{"openid", "profile", "email"},
				},
			},
		},
		"token_config": map[string]interface{}{
			"id_token": map[string]interface{}{
				"user_attributes": []string{"email", "firstName", "lastName", "name"},
				"scope_claims": map[string][]string{
					"profile": {"firstName", "lastName", "name"},
					"email":   {"email"},
				},
			},
		},
	}

	jsonData, err := json.Marshal(app)
	ts.Require().NoError(err, "Failed to marshal application data")

	req, err := http.NewRequest("POST", testServerURL+"/applications", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to create application")
	defer resp.Body.Close()

	ts.Require().Equal(http.StatusCreated, resp.StatusCode, "Failed to create application")

	var respData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	ts.Require().NoError(err, "Failed to parse response")

	appID := respData["id"].(string)
	ts.T().Logf("Created test application with ID: %s", appID)
	return appID
}

func (ts *UserInfoTestSuite) deleteApplication(appID string) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/applications/%s", testServerURL, appID), nil)
	if err != nil {
		ts.T().Errorf("Failed to create delete request: %v", err)
		return
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Errorf("Failed to delete application: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		ts.T().Errorf("Failed to delete application. Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
	} else {
		ts.T().Logf("Successfully deleted test application with ID: %s", appID)
	}
}

// getClientCredentialsToken gets an access token using client_credentials grant
func (ts *UserInfoTestSuite) getClientCredentialsToken(scope string) (string, error) {
	reqBody := strings.NewReader(fmt.Sprintf("grant_type=client_credentials&scope=%s", scope))
	req, err := http.NewRequest("POST", testServerURL+"/oauth2/token", reqBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := ts.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token not found in response")
	}

	return accessToken, nil
}

// initiateAuthorizationFlow starts the OAuth2 authorization flow
func (ts *UserInfoTestSuite) initiateAuthorizationFlow(clientID, redirectURI, responseType, scope, state string) (*http.Response, error) {
	authURL := testServerURL + "/oauth2/authorize"
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", responseType)
	params.Set("scope", scope)
	params.Set("state", state)

	req, err := http.NewRequest("GET", authURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization request: %w", err)
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send authorization request: %w", err)
	}

	return resp, nil
}

// extractSessionData extracts session data from the authorization redirect
func (ts *UserInfoTestSuite) extractSessionData(location string) (string, string, error) {
	redirectURL, err := url.Parse(location)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse redirect URL: %w", err)
	}

	sessionDataKey := redirectURL.Query().Get("sessionDataKey")
	if sessionDataKey == "" {
		return "", "", fmt.Errorf("sessionDataKey not found in redirect")
	}

	flowID := redirectURL.Query().Get("flowId")
	if flowID == "" {
		return "", "", fmt.Errorf("flowId not found in redirect")
	}

	return sessionDataKey, flowID, nil
}

// executeAuthenticationFlow executes an authentication flow
func (ts *UserInfoTestSuite) executeAuthenticationFlow(flowID string, inputs map[string]string) (map[string]interface{}, error) {
	flowData := map[string]interface{}{
		"flowId": flowID,
	}
	if len(inputs) > 0 {
		flowData["inputs"] = inputs
	}

	flowJSON, err := json.Marshal(flowData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow data: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/flow/execute", bytes.NewBuffer(flowJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create flow request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute flow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("flow execution failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var flowStep map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&flowStep)
	if err != nil {
		return nil, fmt.Errorf("failed to decode flow response: %w", err)
	}

	return flowStep, nil
}

// completeAuthorization completes the authorization using the assertion
func (ts *UserInfoTestSuite) completeAuthorization(sessionDataKey, assertion string) (map[string]interface{}, error) {
	authzData := map[string]interface{}{
		"sessionDataKey": sessionDataKey,
		"assertion":      assertion,
	}

	authzJSON, err := json.Marshal(authzData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authorization data: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/oauth2/authorize", bytes.NewBuffer(authzJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization completion request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to complete authorization: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authorization completion failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var authzResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&authzResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode authorization response: %w", err)
	}

	return authzResponse, nil
}

// extractAuthorizationCode extracts the authorization code from the redirect URI
func (ts *UserInfoTestSuite) extractAuthorizationCode(redirectURI string) (string, error) {
	parsedURL, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("failed to parse redirect URI: %w", err)
	}

	code := parsedURL.Query().Get("code")
	if code == "" {
		return "", fmt.Errorf("authorization code not found in redirect URI")
	}

	return code, nil
}

// requestToken performs a token request
func (ts *UserInfoTestSuite) requestToken(clientID, clientSecret, code, redirectURI, grantType string) (map[string]interface{}, int, error) {
	tokenURL := testServerURL + "/oauth2/token"
	tokenData := url.Values{}
	tokenData.Set("grant_type", grantType)
	tokenData.Set("code", code)
	tokenData.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(tokenData.Encode()))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		var tokenResponse map[string]interface{}
		if err := json.Unmarshal(body, &tokenResponse); err != nil {
			return nil, resp.StatusCode, fmt.Errorf("failed to unmarshal token response: %w", err)
		}
		return tokenResponse, resp.StatusCode, nil
	}

	return nil, resp.StatusCode, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
}

// getAuthorizationCodeToken gets an access token using authorization_code grant
func (ts *UserInfoTestSuite) getAuthorizationCodeToken(scope string) (string, error) {
	// Step 1: Initiate authorization flow
	authzResp, err := ts.initiateAuthorizationFlow(clientID, redirectURI, "code", scope, "test_state")
	if err != nil {
		return "", fmt.Errorf("failed to initiate authorization: %w", err)
	}
	defer authzResp.Body.Close()

	location := authzResp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header in authorization response")
	}

	// Step 2: Extract session data
	sessionDataKey, flowID, err := ts.extractSessionData(location)
	if err != nil {
		return "", fmt.Errorf("failed to extract session data: %w", err)
	}

	// Step 3: Execute authentication flow
	authInputs := map[string]string{
		"username": "userinfo_test_user",
		"password": "SecurePass123!",
	}
	flowStep, err := ts.executeAuthenticationFlow(flowID, authInputs)
	if err != nil {
		return "", fmt.Errorf("failed to execute authentication flow: %w", err)
	}

	assertion, ok := flowStep["assertion"].(string)
	if !ok || assertion == "" {
		return "", fmt.Errorf("assertion not found in flow step")
	}

	// Step 4: Complete authorization
	authzResponse, err := ts.completeAuthorization(sessionDataKey, assertion)
	if err != nil {
		return "", fmt.Errorf("failed to complete authorization: %w", err)
	}

	redirectURIStr, ok := authzResponse["redirect_uri"].(string)
	if !ok {
		return "", fmt.Errorf("redirect_uri not found in authorization response")
	}

	// Step 5: Extract authorization code
	code, err := ts.extractAuthorizationCode(redirectURIStr)
	if err != nil {
		return "", fmt.Errorf("failed to extract authorization code: %w", err)
	}

	// Step 6: Exchange code for token
	tokenResp, statusCode, err := ts.requestToken(clientID, clientSecret, code, redirectURI, "authorization_code")
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}

	if statusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", statusCode)
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok || accessToken == "" {
		return "", fmt.Errorf("access token not found in response")
	}

	return accessToken, nil
}

// getRefreshToken gets a refresh token and then uses it to get a new access token
func (ts *UserInfoTestSuite) getRefreshToken(scope string) (string, error) {
	// First get an access token with refresh token using authorization_code grant
	// Step 1: Initiate authorization flow
	authzResp, err := ts.initiateAuthorizationFlow(clientID, redirectURI, "code", scope, "test_state")
	if err != nil {
		return "", fmt.Errorf("failed to initiate authorization: %w", err)
	}
	defer authzResp.Body.Close()

	location := authzResp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header in authorization response")
	}

	// Step 2: Extract session data
	sessionDataKey, flowID, err := ts.extractSessionData(location)
	if err != nil {
		return "", fmt.Errorf("failed to extract session data: %w", err)
	}

	// Step 3: Execute authentication flow
	authInputs := map[string]string{
		"username": "userinfo_test_user",
		"password": "SecurePass123!",
	}
	flowStep, err := ts.executeAuthenticationFlow(flowID, authInputs)
	if err != nil {
		return "", fmt.Errorf("failed to execute authentication flow: %w", err)
	}

	assertion, ok := flowStep["assertion"].(string)
	if !ok || assertion == "" {
		return "", fmt.Errorf("assertion not found in flow step")
	}

	// Step 4: Complete authorization
	authzResponse, err := ts.completeAuthorization(sessionDataKey, assertion)
	if err != nil {
		return "", fmt.Errorf("failed to complete authorization: %w", err)
	}

	redirectURIStr, ok := authzResponse["redirect_uri"].(string)
	if !ok {
		return "", fmt.Errorf("redirect_uri not found in authorization response")
	}

	// Step 5: Extract authorization code
	code, err := ts.extractAuthorizationCode(redirectURIStr)
	if err != nil {
		return "", fmt.Errorf("failed to extract authorization code: %w", err)
	}

	// Step 6: Exchange code for token (this should include refresh_token)
	tokenResp, statusCode, err := ts.requestToken(clientID, clientSecret, code, redirectURI, "authorization_code")
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}

	if statusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", statusCode)
	}

	refreshToken, ok := tokenResp["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return "", fmt.Errorf("refresh_token not found in response")
	}

	// Step 7: Use refresh token to get a new access token
	tokenData := url.Values{}
	tokenData.Set("grant_type", "refresh_token")
	tokenData.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", testServerURL+"/oauth2/token", bytes.NewBufferString(tokenData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := ts.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send refresh token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("refresh token request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var refreshTokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&refreshTokenResp); err != nil {
		return "", fmt.Errorf("failed to decode refresh token response: %w", err)
	}

	newAccessToken, ok := refreshTokenResp["access_token"].(string)
	if !ok || newAccessToken == "" {
		return "", fmt.Errorf("access_token not found in refresh token response")
	}

	return newAccessToken, nil
}

// getTokenExchangeToken gets an access token using token_exchange grant
func (ts *UserInfoTestSuite) getTokenExchangeToken(scope string) (string, error) {
	// First get a subject token using authorization_code
	subjectToken, err := ts.getAuthorizationCodeToken(scope)
	if err != nil {
		return "", err
	}

	// Exchange the subject token for a new token
	tokenData := url.Values{}
	tokenData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	tokenData.Set("subject_token", subjectToken)
	tokenData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	tokenData.Set("scope", scope)

	req, err := http.NewRequest("POST", testServerURL+"/oauth2/token", bytes.NewBufferString(tokenData.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := ts.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token not found in token exchange response")
	}

	return accessToken, nil
}

// callUserInfo calls the UserInfo endpoint with the given access token
func (ts *UserInfoTestSuite) callUserInfo(accessToken string) (*http.Response, error) {
	req, err := http.NewRequest("GET", testServerURL+"/oauth2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// TestUserInfo_ClientCredentialsGrant_Rejected tests that client_credentials grant tokens are rejected
func (ts *UserInfoTestSuite) TestUserInfo_ClientCredentialsGrant_Rejected() {
	// Get access token using client_credentials grant
	accessToken, err := ts.getClientCredentialsToken("read write")
	ts.Require().NoError(err, "Failed to get client_credentials token")
	ts.Require().NotEmpty(accessToken, "Access token should not be empty")

	// Call UserInfo endpoint
	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	assert.Equal(ts.T(), http.StatusUnauthorized, resp.StatusCode, "Should return 401 for client_credentials grant")

	// Parse error response
	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	ts.Require().NoError(err, "Failed to parse error response")

	// Verify error details
	assert.Equal(ts.T(), "invalid_token", errorResp["error"], "Error should be invalid_token")
	assert.Contains(ts.T(), errorResp["error_description"].(string), "client_credentials", "Error description should mention client_credentials")
}

// TestUserInfo_AuthorizationCodeGrant_Allowed tests that authorization_code grant tokens are allowed
func (ts *UserInfoTestSuite) TestUserInfo_AuthorizationCodeGrant_Allowed() {
	// Get access token using authorization_code grant
	accessToken, err := ts.getAuthorizationCodeToken("openid profile email")
	ts.Require().NoError(err, "Failed to get authorization_code token")
	ts.Require().NotEmpty(accessToken, "Access token should not be empty")

	// Call UserInfo endpoint
	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 200 OK
	assert.Equal(ts.T(), http.StatusOK, resp.StatusCode, "Should return 200 for authorization_code grant")

	// Parse response
	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	ts.Require().NoError(err, "Failed to parse UserInfo response")

	// Verify response contains sub claim
	assert.Contains(ts.T(), userInfo, "sub", "Response should contain sub claim")
	assert.NotEmpty(ts.T(), userInfo["sub"], "Sub claim should not be empty")

	// Verify response contains user attributes based on scopes
	// Note: The actual attributes depend on the token configuration
	ts.T().Logf("UserInfo response: %+v", userInfo)
}

// TestUserInfo_RefreshTokenGrant_Allowed tests that refresh_token grant tokens are allowed
func (ts *UserInfoTestSuite) TestUserInfo_RefreshTokenGrant_Allowed() {
	// Get access token using refresh_token grant
	accessToken, err := ts.getRefreshToken("openid profile email")
	ts.Require().NoError(err, "Failed to get refresh_token token")
	ts.Require().NotEmpty(accessToken, "Access token should not be empty")

	// Call UserInfo endpoint
	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 200 OK
	assert.Equal(ts.T(), http.StatusOK, resp.StatusCode, "Should return 200 for refresh_token grant")

	// Parse response
	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	ts.Require().NoError(err, "Failed to parse UserInfo response")

	// Verify response contains sub claim
	assert.Contains(ts.T(), userInfo, "sub", "Response should contain sub claim")
	assert.NotEmpty(ts.T(), userInfo["sub"], "Sub claim should not be empty")

	ts.T().Logf("UserInfo response: %+v", userInfo)
}

// TestUserInfo_TokenExchangeGrant_Allowed tests that token_exchange grant tokens are allowed
func (ts *UserInfoTestSuite) TestUserInfo_TokenExchangeGrant_Allowed() {
	// Get access token using token_exchange grant
	accessToken, err := ts.getTokenExchangeToken("openid profile email")
	ts.Require().NoError(err, "Failed to get token_exchange token")
	ts.Require().NotEmpty(accessToken, "Access token should not be empty")

	// Call UserInfo endpoint
	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 200 OK
	assert.Equal(ts.T(), http.StatusOK, resp.StatusCode, "Should return 200 for token_exchange grant")

	// Parse response
	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	ts.Require().NoError(err, "Failed to parse UserInfo response")

	// Verify response contains sub claim
	assert.Contains(ts.T(), userInfo, "sub", "Response should contain sub claim")
	assert.NotEmpty(ts.T(), userInfo["sub"], "Sub claim should not be empty")

	ts.T().Logf("UserInfo response: %+v", userInfo)
}

// TestUserInfo_InvalidToken tests that invalid tokens are rejected
func (ts *UserInfoTestSuite) TestUserInfo_InvalidToken() {
	// Call UserInfo endpoint with invalid token
	resp, err := ts.callUserInfo("invalid_token")
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	assert.Equal(ts.T(), http.StatusUnauthorized, resp.StatusCode, "Should return 401 for invalid token")

	// Parse error response
	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	ts.Require().NoError(err, "Failed to parse error response")

	// Verify error details
	assert.Equal(ts.T(), "invalid_token", errorResp["error"], "Error should be invalid_token")
}

// TestUserInfo_MissingToken tests that missing tokens are rejected
func (ts *UserInfoTestSuite) TestUserInfo_MissingToken() {
	// Call UserInfo endpoint without token
	req, err := http.NewRequest("GET", testServerURL+"/oauth2/userinfo", nil)
	ts.Require().NoError(err, "Failed to create request")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 401 Unauthorized (OAuth2 spec: missing authentication returns 401)
	assert.Equal(ts.T(), http.StatusUnauthorized, resp.StatusCode, "Should return 401 for missing token")

	// Parse error response
	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	ts.Require().NoError(err, "Failed to parse error response")

	// Verify error details
	assert.Equal(ts.T(), "invalid_request", errorResp["error"], "Error should be invalid_request")
}
