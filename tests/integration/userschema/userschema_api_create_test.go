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

type CreateUserSchemaTestSuite struct {
	suite.Suite
	client         *http.Client
	createdSchemas []string // Track schemas for cleanup
}

func TestCreateUserSchemaTestSuite(t *testing.T) {
	suite.Run(t, new(CreateUserSchemaTestSuite))
}

func (ts *CreateUserSchemaTestSuite) SetupSuite() {
	ts.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	ts.createdSchemas = []string{}
}

func (ts *CreateUserSchemaTestSuite) TearDownSuite() {
	// Clean up created schemas
	for _, schemaID := range ts.createdSchemas {
		ts.deleteSchema(schemaID)
	}
}

// TestCreateUserSchema tests POST /user-schemas with valid data
func (ts *CreateUserSchemaTestSuite) TestCreateUserSchema() {
	schema := CreateUserSchemaRequest{
		Name: "employee-schema-test",
		Schema: json.RawMessage(`{
			"firstName": {"type": "string"},
			"lastName": {"type": "string"},
			"email": {"type": "string", "regex": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"},
			"department": {"type": "string"},
			"isManager": {"type": "boolean"}
		}`),
	}

	jsonData, err := json.Marshal(schema)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/user-schemas", bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusCreated, resp.StatusCode, "Should return 201 Created")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var createdSchema UserSchema
	err = json.Unmarshal(bodyBytes, &createdSchema)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify created schema according to API spec
	ts.Assert().NotEmpty(createdSchema.ID, "Created schema should have ID")
	ts.Assert().Equal(schema.Name, createdSchema.Name, "Name should match")
	ts.Assert().JSONEq(string(schema.Schema), string(createdSchema.Schema), "Schema data should match")

	// Track for cleanup
	ts.createdSchemas = append(ts.createdSchemas, createdSchema.ID)
}

// TestCreateUserSchemaWithComplexSchema tests POST /user-schemas with complex JSON schema
func (ts *CreateUserSchemaTestSuite) TestCreateUserSchemaWithComplexSchema() {
	schema := CreateUserSchemaRequest{
		Name: "complex-customer-schema",
		Schema: json.RawMessage(`{
			"personalInfo": {
				"type": "object",
				"properties": {
					"firstName": {"type": "string"},
					"lastName": {"type": "string"},
					"dateOfBirth": {"type": "string", "regex": "^\\d{4}-\\d{2}-\\d{2}$"}
				}
			},
			"contactInfo": {
				"type": "object",
				"properties": {
					"email": {
						"type": "string",
						"regex": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
					},
					"phone": {"type": "string"},
					"addresses": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"street": {"type": "string"},
								"city": {"type": "string"},
								"zipCode": {"type": "string"}
							}
						}
					}
				}
			},
			"preferences": {
				"type": "object",
				"properties": {
					"newsletter": {"type": "boolean"},
					"theme": {
						"type": "string",
						"enum": ["light", "dark", "auto"]
					}
				}
			}
		}`),
	}

	jsonData, err := json.Marshal(schema)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/user-schemas", bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusCreated, resp.StatusCode, "Should return 201 Created")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var createdSchema UserSchema
	err = json.Unmarshal(bodyBytes, &createdSchema)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify complex schema was stored correctly
	ts.Assert().NotEmpty(createdSchema.ID, "Created schema should have ID")
	ts.Assert().Equal(schema.Name, createdSchema.Name, "Name should match")
	ts.Assert().JSONEq(string(schema.Schema), string(createdSchema.Schema), "Complex schema data should match")

	// Track for cleanup
	ts.createdSchemas = append(ts.createdSchemas, createdSchema.ID)
}

// TestCreateUserSchemaWithDuplicateName tests POST /user-schemas with duplicate name
func (ts *CreateUserSchemaTestSuite) TestCreateUserSchemaWithDuplicateName() {
	// First create a schema
	schema1 := CreateUserSchemaRequest{
		Name:   "duplicate-name-test",
		Schema: json.RawMessage(`{"field1": {"type": "string"}}`),
	}

	createdID := ts.createSchemaHelper(schema1)
	ts.createdSchemas = append(ts.createdSchemas, createdID)

	// Try to create another schema with same name
	schema2 := CreateUserSchemaRequest{
		Name:   "duplicate-name-test", // Same name
		Schema: json.RawMessage(`{"field2": {"type": "string"}}`),
	}

	jsonData, err := json.Marshal(schema2)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/user-schemas", bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusConflict, resp.StatusCode, "Should return 409 Conflict for duplicate name")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var errorResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errorResp)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal error response: %v", err)
	}

	ts.Assert().NotEmpty(errorResp.Code, "Error should have code")
	ts.Assert().NotEmpty(errorResp.Message, "Error should have message")
}

// TestCreateUserSchemaWithInvalidData tests POST /user-schemas with invalid request data
func (ts *CreateUserSchemaTestSuite) TestCreateUserSchemaWithInvalidData() {
	testCases := []struct {
		name        string
		requestBody string
	}{
		{
			name:        "empty name",
			requestBody: `{"name": "", "schema": {"field": {"type": "string"}}}`,
		},
		{
			name:        "missing name",
			requestBody: `{"schema": {"field": {"type": "string"}}}`,
		},
		{
			name:        "empty schema",
			requestBody: `{"name": "test-schema", "schema": {}}`,
		},
		{
			name:        "missing schema",
			requestBody: `{"name": "test-schema"}`,
		},
		{
			name:        "invalid JSON",
			requestBody: `{"name": "test-schema", "schema": invalid}`,
		},
		{
			name:        "malformed JSON",
			requestBody: `{"name": "test-schema"`,
		},
	}

	for _, tc := range testCases {
		ts.T().Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", testServerURL+"/user-schemas", bytes.NewBufferString(tc.requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := ts.client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			ts.Assert().Equal(http.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request for: %s", tc.name)

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var errorResp ErrorResponse
			err = json.Unmarshal(bodyBytes, &errorResp)
			if err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			ts.Assert().NotEmpty(errorResp.Code, "Error should have code")
			ts.Assert().NotEmpty(errorResp.Message, "Error should have message")
		})
	}
}

// TestCreateUserSchemaWithoutContentType tests POST /user-schemas without Content-Type header
func (ts *CreateUserSchemaTestSuite) TestCreateUserSchemaWithoutContentType() {
	schema := CreateUserSchemaRequest{
		Name:   "no-content-type-test",
		Schema: json.RawMessage(`{"field": {"type": "string"}}`),
	}

	jsonData, err := json.Marshal(schema)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/user-schemas", bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	// Intentionally not setting Content-Type header

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Should handle gracefully, either accept or reject with appropriate error
	ts.Assert().True(resp.StatusCode == http.StatusBadRequest ||
		resp.StatusCode == http.StatusCreated ||
		resp.StatusCode == http.StatusUnsupportedMediaType,
		"Should handle missing content-type appropriately, got status: %d", resp.StatusCode)

	// Clean up if created successfully
	if resp.StatusCode == http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		var createdSchema UserSchema
		if json.Unmarshal(bodyBytes, &createdSchema) == nil {
			ts.createdSchemas = append(ts.createdSchemas, createdSchema.ID)
		}
	}
}

// Helper function to create a schema and return its ID
func (ts *CreateUserSchemaTestSuite) createSchemaHelper(schema CreateUserSchemaRequest) string {
	jsonData, err := json.Marshal(schema)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/user-schemas", bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Expected status 201, got %d. Response: %s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var createdSchema UserSchema
	err = json.Unmarshal(bodyBytes, &createdSchema)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal response: %v", err)
	}

	return createdSchema.ID
}

// Helper function to delete a schema
func (ts *CreateUserSchemaTestSuite) deleteSchema(schemaID string) {
	req, err := http.NewRequest("DELETE", testServerURL+"/user-schemas/"+schemaID, nil)
	if err != nil {
		return
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
