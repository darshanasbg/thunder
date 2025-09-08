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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserValidationTestSuite struct {
	suite.Suite
	client         *http.Client
	createdSchemas []string // Track schemas for cleanup
	createdUsers   []string // Track users for cleanup
}

func TestUserValidationTestSuite(t *testing.T) {
	suite.Run(t, new(UserValidationTestSuite))
}

func (ts *UserValidationTestSuite) SetupSuite() {
	ts.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	ts.createdSchemas = []string{}
	ts.createdUsers = []string{}

	ts.createEmployeeSchema()
	ts.createSchemaWithEnum()
	ts.createSchemaWithNestedObject()
	ts.createSchemaWithArray()
}

func (ts *UserValidationTestSuite) TearDownSuite() {
	// Clean up created users first
	for _, userID := range ts.createdUsers {
		ts.deleteUser(userID)
	}
	// Then clean up created schemas
	for _, schemaID := range ts.createdSchemas {
		ts.deleteSchema(schemaID)
	}
}

// TestCreateUserWithValidSchema tests user creation with valid schema
func (ts *UserValidationTestSuite) TestCreateUserWithValidSchema() {
	// Create a user that conforms to the schema
	createUserReq := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "employee",
		Attributes: json.RawMessage(`{
			"firstName": "John",
			"lastName": "Doe",
			"email": "john.doe@company.com",
			"department": "Engineering",
			"isManager": false
		}`),
	}

	userID := ts.createUserAndExpectSuccess(createUserReq)
	ts.createdUsers = append(ts.createdUsers, userID)
}

// TestCreateUserWithInvalidStringType tests user creation with invalid string value
func (ts *UserValidationTestSuite) TestCreateUserWithInvalidStringType() {
	// Create a user with invalid string type (number instead of string)
	createUserReq := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "employee",
		Attributes: json.RawMessage(`{
			"firstName": 123,
			"lastName": "Doe",
			"email": "john.doe@company.com",
			"department": "Engineering",
			"isManager": false
		}`),
	}

	ts.createUserAndExpectError(createUserReq, "USR-1019")
}

// TestCreateUserWithInvalidBooleanType tests user creation with invalid boolean value
func (ts *UserValidationTestSuite) TestCreateUserWithInvalidBooleanType() {
	// Create a user with invalid boolean type (string instead of boolean)
	createUserReq := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "employee",
		Attributes: json.RawMessage(`{
			"firstName": "John",
			"lastName": "Doe",
			"email": "john.doe@company.com",
			"department": "Engineering",
			"isManager": "yes"
		}`),
	}

	ts.createUserAndExpectError(createUserReq, "USR-1019")
}

// TestCreateUserWithEnumValidation tests user creation with enum constraints
func (ts *UserValidationTestSuite) TestCreateUserWithEnumValidation() {
	// Test valid enum value
	createUserReq := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "student",
		Attributes: json.RawMessage(`{
			"name": "Alice",
			"grade": "A",
			"semester": 1
		}`),
	}

	userID := ts.createUserAndExpectSuccess(createUserReq)
	ts.createdUsers = append(ts.createdUsers, userID)

	// Test invalid enum value
	createUserReq2 := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "student",
		Attributes: json.RawMessage(`{
			"name": "Bob",
			"grade": "Z",
			"semester": 1
		}`),
	}

	ts.createUserAndExpectError(createUserReq2, "USR-1019")
}

// TestCreateUserWithNestedObjectValidation tests user creation with nested object validation
func (ts *UserValidationTestSuite) TestCreateUserWithNestedObjectValidation() {
	// Test valid nested object
	createUserReq := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "customer",
		Attributes: json.RawMessage(`{
			"name": "John Smith",
			"address": {
				"street": "123 Main St",
				"city": "Seattle",
				"zipCode": "98101"
			}
		}`),
	}

	userID := ts.createUserAndExpectSuccess(createUserReq)
	ts.createdUsers = append(ts.createdUsers, userID)

	// Test invalid nested object (wrong type for zipCode)
	createUserReq2 := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "customer",
		Attributes: json.RawMessage(`{
			"name": "Jane Smith",
			"address": {
				"street": "456 Oak Ave",
				"city": "Portland",
				"zipCode": 97201
			}
		}`),
	}

	ts.createUserAndExpectError(createUserReq2, "USR-1019")
}

// TestCreateUserWithArrayValidation tests user creation with array validation
func (ts *UserValidationTestSuite) TestCreateUserWithArrayValidation() {
	// Test valid array
	createUserReq := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "teacher",
		Attributes: json.RawMessage(`{
			"name": "Prof. Johnson",
			"subjects": ["Math", "Physics", "Chemistry"]
		}`),
	}

	userID := ts.createUserAndExpectSuccess(createUserReq)
	ts.createdUsers = append(ts.createdUsers, userID)

	// Test invalid array (number instead of string)
	createUserReq2 := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "teacher",
		Attributes: json.RawMessage(`{
			"name": "Prof. Smith",
			"subjects": ["Math", 123, "Chemistry"]
		}`),
	}

	ts.createUserAndExpectError(createUserReq2, "USR-1019")
}

// TestUpdateUserWithValidSchema tests user update with valid schema
func (ts *UserValidationTestSuite) TestUpdateUserWithValidSchema() {
	// Create a user
	createUserReq := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "employee",
		Attributes: json.RawMessage(`{
			"firstName": "John",
			"lastName": "Doe",
			"email": "john.doe@company.com",
			"department": "Engineering",
			"isManager": false
		}`),
	}

	userID := ts.createUserAndExpectSuccess(createUserReq)
	ts.createdUsers = append(ts.createdUsers, userID)

	// Update user with valid data
	updateUserReq := UpdateUserRequest{
		Type: "employee",
		Attributes: json.RawMessage(`{
			"firstName": "John",
			"lastName": "Smith",
			"email": "john.smith@company.com",
			"department": "Product",
			"isManager": true
		}`),
	}

	ts.updateUserAndExpectSuccess(userID, updateUserReq)
}

// TestUpdateUserWithInvalidSchema tests user update with invalid schema
func (ts *UserValidationTestSuite) TestUpdateUserWithInvalidSchema() {
	// Create a user
	createUserReq := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "employee",
		Attributes: json.RawMessage(`{
			"firstName": "John",
			"lastName": "Doe",
			"email": "john.doe@company.com",
			"department": "Engineering",
			"isManager": false
		}`),
	}

	userID := ts.createUserAndExpectSuccess(createUserReq)
	ts.createdUsers = append(ts.createdUsers, userID)

	// Update user with invalid data (wrong type for isManager)
	updateUserReq := UpdateUserRequest{
		Type: "employee",
		Attributes: json.RawMessage(`{
			"firstName": "John",
			"lastName": "Smith",
			"email": "john.smith@company.com",
			"department": "Product",
			"isManager": "true"
		}`),
	}

	ts.updateUserAndExpectError(userID, updateUserReq, "USR-1019")
}

// TestCreateUserWithoutSchema tests user creation without a defined schema (should succeed)
func (ts *UserValidationTestSuite) TestCreateUserWithoutSchema() {
	// Create a user with a type that has no schema
	createUserReq := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "untyped-user",
		Attributes: json.RawMessage(`{
			"anyField": "anyValue",
			"randomNumber": 42
		}`),
	}

	userID := ts.createUserAndExpectSuccess(createUserReq)
	ts.createdUsers = append(ts.createdUsers, userID)
}

// Helper methods

func (ts *UserValidationTestSuite) createEmployeeSchema() string {
	schema := CreateUserSchemaRequest{
		Name: "employee",
		Schema: json.RawMessage(`{
			"firstName": {"type": "string"},
			"lastName": {"type": "string"},
			"email": {"type": "string"},
			"department": {"type": "string"},
			"isManager": {"type": "boolean"}
		}`),
	}

	return ts.createSchema(schema)
}

func (ts *UserValidationTestSuite) createSchemaWithEnum() string {
	schema := CreateUserSchemaRequest{
		Name: "student",
		Schema: json.RawMessage(`{
			"name": {"type": "string"},
			"grade": {"type": "string", "enum": ["A", "B", "C", "D", "F"]},
			"semester": {"type": "number", "enum": [1, 2, 3, 4, 5, 6, 7, 8]}
		}`),
	}

	return ts.createSchema(schema)
}

func (ts *UserValidationTestSuite) createSchemaWithNestedObject() string {
	schema := CreateUserSchemaRequest{
		Name: "customer",
		Schema: json.RawMessage(`{
			"name": {"type": "string"},
			"address": {
				"type": "object",
				"properties": {
					"street": {"type": "string"},
					"city": {"type": "string"},
					"zipCode": {"type": "string"}
				}
			}
		}`),
	}

	return ts.createSchema(schema)
}

func (ts *UserValidationTestSuite) createSchemaWithArray() string {
	schema := CreateUserSchemaRequest{
		Name: "teacher",
		Schema: json.RawMessage(`{
			"name": {"type": "string"},
			"subjects": {
				"type": "array",
				"items": {"type": "string"}
			}
		}`),
	}

	return ts.createSchema(schema)
}

func (ts *UserValidationTestSuite) createSchema(schema CreateUserSchemaRequest) string {
	jsonData, err := json.Marshal(schema)
	ts.Require().NoError(err, "Failed to marshal schema request")

	req, err := http.NewRequest("POST", testServerURL+"/user-schemas", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create schema request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send schema request")
	defer resp.Body.Close()

	ts.Require().Equal(201, resp.StatusCode, "Schema creation should succeed")

	body, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read schema response body")

	var createdSchema UserSchema
	err = json.Unmarshal(body, &createdSchema)
	ts.Require().NoError(err, "Failed to unmarshal schema response")

	ts.createdSchemas = append(ts.createdSchemas, createdSchema.ID)
	return createdSchema.ID
}

func (ts *UserValidationTestSuite) createUserAndExpectSuccess(createUserReq CreateUserRequest) string {
	jsonData, err := json.Marshal(createUserReq)
	ts.Require().NoError(err, "Failed to marshal user request")

	req, err := http.NewRequest("POST", testServerURL+"/users", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create user request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send user request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read user response body")

	if resp.StatusCode != 201 {
		ts.T().Logf("User creation failed with status %d: %s", resp.StatusCode, string(body))
	}
	ts.Require().Equal(201, resp.StatusCode, "User creation should succeed")

	var createdUser User
	err = json.Unmarshal(body, &createdUser)
	ts.Require().NoError(err, "Failed to unmarshal user response")

	return createdUser.ID
}

func (ts *UserValidationTestSuite) createUserAndExpectError(createUserReq CreateUserRequest, expectedErrorCode string) {
	jsonData, err := json.Marshal(createUserReq)
	ts.Require().NoError(err, "Failed to marshal user request")

	req, err := http.NewRequest("POST", testServerURL+"/users", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create user request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send user request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read user response body")

	ts.Require().Equal(400, resp.StatusCode, "User creation should fail with validation error")

	var errorResp ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	ts.Require().NoError(err, "Failed to unmarshal error response")
	ts.Require().Equal(expectedErrorCode, errorResp.Code, "Error code should match expected")
}

func (ts *UserValidationTestSuite) updateUserAndExpectSuccess(userID string, updateUserReq UpdateUserRequest) {
	jsonData, err := json.Marshal(updateUserReq)
	ts.Require().NoError(err, "Failed to marshal update user request")

	req, err := http.NewRequest("PUT", testServerURL+"/users/"+userID, bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create update user request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send update user request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read update user response body")

	if resp.StatusCode != 200 {
		ts.T().Logf("User update failed with status %d: %s", resp.StatusCode, string(body))
	}
	ts.Require().Equal(200, resp.StatusCode, "User update should succeed")
}

func (ts *UserValidationTestSuite) updateUserAndExpectError(userID string, updateUserReq UpdateUserRequest, expectedErrorCode string) {
	jsonData, err := json.Marshal(updateUserReq)
	ts.Require().NoError(err, "Failed to marshal update user request")

	req, err := http.NewRequest("PUT", testServerURL+"/users/"+userID, bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create update user request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send update user request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read update user response body")

	ts.Require().Equal(400, resp.StatusCode, "User update should fail with validation error")

	var errorResp ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	ts.Require().NoError(err, "Failed to unmarshal error response")
	ts.Require().Equal(expectedErrorCode, errorResp.Code, "Error code should match expected")
}

func (ts *UserValidationTestSuite) deleteUser(userID string) {
	req, err := http.NewRequest("DELETE", testServerURL+"/users/"+userID, nil)
	if err != nil {
		ts.T().Logf("Failed to create delete user request: %v", err)
		return
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Logf("Failed to send delete user request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 404 {
		body, _ := io.ReadAll(resp.Body)
		ts.T().Logf("Failed to delete user %s: status %d, body: %s", userID, resp.StatusCode, string(body))
	}
}

func (ts *UserValidationTestSuite) deleteSchema(schemaID string) {
	req, err := http.NewRequest("DELETE", testServerURL+"/user-schemas/"+schemaID, nil)
	if err != nil {
		ts.T().Logf("Failed to create delete schema request: %v", err)
		return
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Logf("Failed to send delete schema request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 404 {
		body, _ := io.ReadAll(resp.Body)
		ts.T().Logf("Failed to delete schema %s: status %d, body: %s", schemaID, resp.StatusCode, string(body))
	}
}
