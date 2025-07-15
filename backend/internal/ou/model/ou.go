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

// Package model defines the data structures and interfaces for organization unit management.
package model

// OrganizationUnitBasic represents the basic information of an organization unit.
type OrganizationUnitBasic struct {
	ID                string   `json:"id"`
	Handle            string   `json:"handle"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	Parent            *string  `json:"parent"`
	OrganizationUnits []string `json:"organizationUnits"`
}

// OrganizationUnit represents a complete organization unit with users, groups, and sub organization units.
type OrganizationUnit struct {
	ID                string   `json:"id"`
	Handle            string   `json:"handle"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	Parent            *string  `json:"parent"`
	Users             []string `json:"users,omitempty"`
	Groups            []string `json:"groups,omitempty"`
	OrganizationUnits []string `json:"organizationUnits"`
}

// OrganizationUnitRequest represents the request body for creating an organization unit.
type OrganizationUnitRequest struct {
	Handle      string  `json:"handle"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent"`
}

// Link represents a pagination link.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// OrganizationUnitListResponse represents the response for listing organization units with pagination.
type OrganizationUnitListResponse struct {
	TotalResults      int                     `json:"totalResults"`
	StartIndex        int                     `json:"startIndex"`
	Count             int                     `json:"count"`
	OrganizationUnits []OrganizationUnitBasic `json:"organizationUnits"`
	Links             []Link                  `json:"links"`
}
