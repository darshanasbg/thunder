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

// Package service provides the implementation for group management operations.
package service

import (
	"errors"

	"github.com/asgardeo/thunder/internal/group/model"
	"github.com/asgardeo/thunder/internal/group/store"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// GroupServiceInterface defines the interface for the group service.
type GroupServiceInterface interface {
	GetGroupList() ([]model.GroupBasic, error)
	CreateGroup(request model.CreateGroupRequest) (*model.Group, error)
	GetGroup(groupID string) (*model.Group, error)
	UpdateGroup(groupID string, request model.UpdateGroupRequest) (*model.Group, error)
	DeleteGroup(groupID string) error
}

// GroupService is the default implementation of the GroupServiceInterface.
type GroupService struct{}

// GetGroupService creates a new instance of GroupService.
func GetGroupService() GroupServiceInterface {
	return &GroupService{}
}

// GetGroupList retrieves a list of root groups.
func (gs *GroupService) GetGroupList() ([]model.GroupBasic, error) {
	groups, err := store.GetGroupList()
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// CreateGroup creates a new group.
func (gs *GroupService) CreateGroup(request model.CreateGroupRequest) (*model.Group, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "GroupService"))

	// Validate request
	if err := validateCreateGroupRequest(request); err != nil {
		logger.Error("Invalid create group request", log.Error(err))
		return nil, err
	}

	// Create group object
	group := model.Group{
		ID:          utils.GenerateUUID(),
		Name:        request.Name,
		Description: request.Description,
		Parent:      request.Parent,
		Users:       request.Users,
		Groups:      []string{}, // Initialize empty child groups
	}

	// Create group in the database
	err := store.CreateGroup(group)
	if err != nil {
		logger.Error("Failed to create group", log.Error(err))
		return nil, err
	}

	// Return the created group
	createdGroup, err := store.GetGroup(group.ID)
	if err != nil {
		logger.Error("Failed to get created group", log.Error(err))
		return nil, err
	}

	return &createdGroup, nil
}

// GetGroup retrieves a specific group by its id.
func (gs *GroupService) GetGroup(groupID string) (*model.Group, error) {
	if groupID == "" {
		return nil, errors.New("group id is empty")
	}

	group, err := store.GetGroup(groupID)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

// UpdateGroup updates an existing group.
func (gs *GroupService) UpdateGroup(groupID string, request model.UpdateGroupRequest) (*model.Group, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "GroupService"))

	if groupID == "" {
		return nil, errors.New("group id is empty")
	}

	// Validate request
	if err := validateUpdateGroupRequest(request); err != nil {
		logger.Error("Invalid update group request", log.Error(err))
		return nil, err
	}

	// Get existing group to ensure it exists
	existingGroup, err := store.GetGroup(groupID)
	if err != nil {
		return nil, err
	}

	// Create updated group object
	updatedGroup := model.Group{
		ID:          existingGroup.ID,
		Name:        request.Name,
		Description: request.Description,
		Parent:      request.Parent,
		Users:       request.Users,
		Groups:      request.Groups,
	}

	// Update group in the database
	err = store.UpdateGroup(updatedGroup)
	if err != nil {
		logger.Error("Failed to update group", log.Error(err))
		return nil, err
	}

	// Return the updated group
	group, err := store.GetGroup(groupID)
	if err != nil {
		logger.Error("Failed to get updated group", log.Error(err))
		return nil, err
	}

	return &group, nil
}

// DeleteGroup deletes a group.
func (gs *GroupService) DeleteGroup(groupID string) error {
	if groupID == "" {
		return errors.New("group id is empty")
	}

	err := store.DeleteGroup(groupID)
	if err != nil {
		return err
	}

	return nil
}

func validateCreateGroupRequest(request model.CreateGroupRequest) error {
	if request.Name == "" {
		return model.ErrInvalidRequest
	}

	if request.Parent.Type == "" || request.Parent.ID == "" {
		return model.ErrInvalidRequest
	}

	// Use ParentType constants for validation
	if request.Parent.Type != model.ParentTypeGroup && request.Parent.Type != model.ParentTypeOrganizationUnit {
		return model.ErrInvalidRequest
	}

	return nil
}

func validateUpdateGroupRequest(request model.UpdateGroupRequest) error {
	if request.Name == "" {
		return model.ErrInvalidRequest
	}

	if request.Parent.Type == "" || request.Parent.ID == "" {
		return model.ErrInvalidRequest
	}

	// Use ParentType constants for validation
	if request.Parent.Type != model.ParentTypeGroup && request.Parent.Type != model.ParentTypeOrganizationUnit {
		return model.ErrInvalidRequest
	}

	return nil
}
