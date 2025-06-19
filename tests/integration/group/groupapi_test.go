/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package group

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
)

var (
	testOU = "456e8400-e29b-41d4-a716-446655440001"

	groupToCreate = CreateGroupRequest{
		Name: "Test Group",
		Parent: Parent{
			Type: "organizationUnit",
			Id:   testOU,
		},
		Users: []string{"550e8400-e29b-41d4-a716-446655440000"},
	}
)

type GroupAPITestSuite struct {
	suite.Suite
	httpClient   *http.Client
	createdGroup *Group
}

func (suite *GroupAPITestSuite) SetupSuite() {
	// Create HTTP client with insecure TLS config for testing
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	suite.httpClient = &http.Client{Transport: tr}
}

func (suite *GroupAPITestSuite) TestCreateGroup() {
	// Create a new group
	jsonData, err := json.Marshal(groupToCreate)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.httpClient.Do(req)
	if err != nil {
		suite.T().Fatalf("HTTP request failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Failed to close response body: %v", err)
		}
	}()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.T().Fatalf("Failed to read response body: %v", err)
	}

	var createdGroup Group
	err = json.Unmarshal(body, &createdGroup)
	suite.Require().NoError(err)

	// Verify the created group
	suite.NotEmpty(createdGroup.Id)
	suite.Equal(groupToCreate.Name, createdGroup.Name)
	suite.Equal(groupToCreate.Parent.Type, createdGroup.Parent.Type)
	suite.Equal(groupToCreate.Parent.Id, createdGroup.Parent.Id)

	// Store the created group for other tests
	suite.createdGroup = &createdGroup
}

func (suite *GroupAPITestSuite) TestGetGroup() {
	suite.Require().NotNil(suite.createdGroup, "Group must be created first")

	// Get the created group
	req, err := http.NewRequest("GET", testServerURL+"/groups/"+suite.createdGroup.Id, nil)
	suite.Require().NoError(err)

	resp, err := suite.httpClient.Do(req)
	suite.Require().NoError(err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Failed to close response body: %v", err)
		}
	}()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.T().Fatalf("Failed to read response body: %v", err)
	}

	var retrievedGroup Group
	err = json.Unmarshal(body, &retrievedGroup)
	suite.Require().NoError(err)

	// Verify the retrieved group
	suite.Equal(suite.createdGroup.Id, retrievedGroup.Id)
	suite.Equal(suite.createdGroup.Name, retrievedGroup.Name)
	suite.Equal(suite.createdGroup.Parent.Type, retrievedGroup.Parent.Type)
	suite.Equal(suite.createdGroup.Parent.Id, retrievedGroup.Parent.Id)
}

func (suite *GroupAPITestSuite) TestListGroups() {
	suite.Require().NotNil(suite.createdGroup, "Group must be created first")

	// List groups
	req, err := http.NewRequest("GET", testServerURL+"/groups", nil)
	suite.Require().NoError(err)

	resp, err := suite.httpClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.T().Fatalf("Failed to read response body: %v", err)
	}

	var groups []GroupBasic
	err = json.Unmarshal(body, &groups)
	suite.Require().NoError(err)

	// Verify the list contains our created group
	found := false
	for _, group := range groups {
		if group.Id == suite.createdGroup.Id {
			found = true
			suite.Equal(suite.createdGroup.Name, group.Name)
			break
		}
	}
	suite.True(found, "Created group should be in the list")
}

func (suite *GroupAPITestSuite) TestUpdateGroup() {
	suite.Require().NotNil(suite.createdGroup, "Group must be created first")

	// Update the group
	updateRequest := map[string]interface{}{
		"name": "Updated Test Group",
		"parent": map[string]string{
			"type": "organizationUnit",
			"id":   testOU,
		},
		"users":  []string{},
		"groups": []string{},
	}

	jsonData, err := json.Marshal(updateRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("PUT", testServerURL+"/groups/"+suite.createdGroup.Id, bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.httpClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var updatedGroup Group
	err = json.Unmarshal(body, &updatedGroup)
	suite.Require().NoError(err)

	// Verify the update
	suite.Equal(suite.createdGroup.Id, updatedGroup.Id)
	suite.Equal("Updated Test Group", updatedGroup.Name)
}

func (suite *GroupAPITestSuite) TestDeleteGroup() {
	suite.Require().NotNil(suite.createdGroup, "Group must be created first")

	// Delete the group
	req, err := http.NewRequest("DELETE", testServerURL+"/groups/"+suite.createdGroup.Id, nil)
	suite.Require().NoError(err)

	resp, err := suite.httpClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify the group is deleted by trying to get it
	getReq, err := http.NewRequest("GET", testServerURL+"/groups/"+suite.createdGroup.Id, nil)
	suite.Require().NoError(err)

	getResp, err := suite.httpClient.Do(getReq)
	if err != nil {
		suite.T().Fatalf("Failed to execute GET request: %v", err)
	}
	defer getResp.Body.Close()

	suite.Equal(http.StatusNotFound, getResp.StatusCode)
}

func (suite *GroupAPITestSuite) TestGetNonExistentGroup() {
	// Try to get a non-existent group
	req, err := http.NewRequest("GET", testServerURL+"/groups/non-existent-id", nil)
	suite.Require().NoError(err)

	resp, err := suite.httpClient.Do(req)
	if err != nil {
		suite.T().Fatalf("Failed to execute GET request: %v", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

func (suite *GroupAPITestSuite) TestCreateGroupWithInvalidData() {
	// Try to create a group with invalid data (missing name)
	invalidGroup := map[string]interface{}{
		"parent": map[string]string{
			"type": "organizationUnit",
			"id":   testOU,
		},
	}

	jsonData, err := json.Marshal(invalidGroup)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.httpClient.Do(req)
	if err != nil {
		suite.T().Fatalf("Failed to execute POST request: %v", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestGroupAPITestSuite(t *testing.T) {
	suite.Run(t, new(GroupAPITestSuite))
}
