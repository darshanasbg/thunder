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

// Package model defines the data structures and interfaces for user management.
package model

import (
	"encoding/json"
	"errors"
)

// User represents a user in the system.
type User struct {
	ID               string          `json:"id,omitempty"`
	OrganizationUnit string          `json:"organizationUnit,omitempty"`
	Type             string          `json:"type,omitempty"`
	Attributes       json.RawMessage `json:"attributes,omitempty"`
}

// Credentials represents the credentials of a user.
type Credentials struct {
	CredentialType string `json:"credentialType"`
	StorageType    string `json:"storageType"`
	StorageAlgo    string `json:"storageAlgo"`
	Value          string `json:"value"`
	Salt           string `json:"salt"`
}

// Link represents a pagination link.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// UserListResponse represents the response for listing users with pagination.
type UserListResponse struct {
	TotalResults int    `json:"totalResults"`
	StartIndex   int    `json:"startIndex"`
	Count        int    `json:"count"`
	Users        []User `json:"users"`
	Links        []Link `json:"links"`
}

// UserGroup represents a group with basic information for user endpoints.
type UserGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UserGroupListResponse represents the response for listing groups that a user belongs to.
type UserGroupListResponse struct {
	TotalResults int         `json:"totalResults"`
	StartIndex   int         `json:"startIndex"`
	Count        int         `json:"count"`
	Groups       []UserGroup `json:"groups"`
	Links        []Link      `json:"links"`
}

// CreateUserRequest represents the request body for creating a user.
type CreateUserRequest struct {
	OrganizationUnit string          `json:"organizationUnit"`
	Type             string          `json:"type"`
	Groups           []string        `json:"groups,omitempty"`
	Attributes       json.RawMessage `json:"attributes,omitempty"`
}

// UpdateUserRequest represents the request body for updating a user.
type UpdateUserRequest struct {
	OrganizationUnit string          `json:"organizationUnit,omitempty"`
	Type             string          `json:"type,omitempty"`
	Groups           []string        `json:"groups,omitempty"`
	Attributes       json.RawMessage `json:"attributes,omitempty"`
}

// CreateUserByPathRequest represents the request body for creating a user under a handle path.
type CreateUserByPathRequest struct {
	Type       string          `json:"type"`
	Groups     []string        `json:"groups,omitempty"`
	Attributes json.RawMessage `json:"attributes,omitempty"`
}

// AuthenticateUserRequest represents the request body for authenticating a user.
type AuthenticateUserRequest map[string]interface{}

// AuthenticateUserResponse represents the response body for authenticating a user.
type AuthenticateUserResponse struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	OrganizationUnit string `json:"organizationUnit"`
}

// ErrUserNotFound is returned when the user is not found in the system.
var ErrUserNotFound = errors.New("user not found")

// ErrBadAttributesInRequest is returned when the attributes in the request are invalid.
var ErrBadAttributesInRequest = errors.New("failed to marshal attributes")
