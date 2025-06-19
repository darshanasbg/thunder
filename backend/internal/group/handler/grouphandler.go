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

// Package handler provides the implementation for group management operations.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/group/model"
	"github.com/asgardeo/thunder/internal/group/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// GroupHandler is the handler for group management operations.
type GroupHandler struct {
}

// HandleGroupListRequest handles the get groups list request.
func (gh *GroupHandler) HandleGroupListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "GroupHandler"))

	groupProvider := provider.NewGroupProvider()
	groupService := groupProvider.GetGroupService()
	groups, err := groupService.GetGroupList()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(groups)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the groups response
	logger.Debug("Groups GET (list) response sent")
}

// HandleGroupPostRequest handles the create group request.
func (gh *GroupHandler) HandleGroupPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "GroupHandler"))

	var createRequest model.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		http.Error(w, "Bad Request: The request body is malformed or contains invalid data.", http.StatusBadRequest)
		return
	}

	// Create the group using the group service
	groupProvider := provider.NewGroupProvider()
	groupService := groupProvider.GetGroupService()
	createdGroup, err := groupService.CreateGroup(createRequest)
	if err != nil {
		if errors.Is(err, model.ErrGroupNameConflict) {
			http.Error(w, "Conflict: A group with the same name exists under the same parent.", http.StatusConflict)
		} else if errors.Is(err, model.ErrParentNotFound) {
			http.Error(w, "Bad Request: Parent group or organization unit not found.", http.StatusBadRequest)
		} else if errors.Is(err, model.ErrInvalidRequest) {
			// TODO: Check whether this case is present and needed
			http.Error(w, "Bad Request: The request body is malformed or contains invalid data.", http.StatusBadRequest)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(createdGroup)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the group creation response
	logger.Debug("Group POST response sent", log.String("group id", createdGroup.ID))
}

// HandleGroupGetRequest handles the get group by id request.
func (gh *GroupHandler) HandleGroupGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "GroupHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/groups/")
	if id == "" {
		http.Error(w, "Bad Request: Missing group id.", http.StatusBadRequest)
		return
	}

	// Get the group using the group service
	groupProvider := provider.NewGroupProvider()
	groupService := groupProvider.GetGroupService()
	group, err := groupService.GetGroup(id)
	if err != nil {
		if errors.Is(err, model.ErrGroupNotFound) {
			http.Error(w, "Not Found: The group with the specified id does not exist.", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(group)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the group response
	logger.Debug("Group GET response sent", log.String("group id", id))
}

// HandleGroupPutRequest handles the update group request.
func (gh *GroupHandler) HandleGroupPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "GroupHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/groups/")
	if id == "" {
		http.Error(w, "Bad Request: Missing group id.", http.StatusBadRequest)
		return
	}

	var updateRequest model.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		http.Error(w, "Bad Request: The request body is malformed or contains invalid data.", http.StatusBadRequest)
		return
	}

	// Update the group using the group service
	groupProvider := provider.NewGroupProvider()
	groupService := groupProvider.GetGroupService()
	group, err := groupService.UpdateGroup(id, updateRequest)
	if err != nil {
		if errors.Is(err, model.ErrGroupNotFound) {
			http.Error(w, "Not Found: The group with the specified id does not exist.", http.StatusNotFound)
		} else if errors.Is(err, model.ErrGroupNameConflict) {
			// TODO: Check whether it exclude name validation when name is not changed
			http.Error(w, "Conflict: A group with the new name exists under the same parent.", http.StatusConflict)
		} else if errors.Is(err, model.ErrParentNotFound) {
			http.Error(w, "Bad Request: Parent group or organization unit not found.", http.StatusBadRequest)
		} else if errors.Is(err, model.ErrInvalidRequest) {
			// TODO: Check whether this case is present and needed
			http.Error(w, "Bad Request: The request body is malformed or contains invalid data.", http.StatusBadRequest)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(group)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the group response
	logger.Debug("Group PUT response sent", log.String("group id", id))
}

// HandleGroupDeleteRequest handles the delete group request.
func (gh *GroupHandler) HandleGroupDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "GroupHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/groups/")
	if id == "" {
		http.Error(w, "Bad Request: Missing group id.", http.StatusBadRequest)
		return
	}

	// Delete the group using the group service
	groupProvider := provider.NewGroupProvider()
	groupService := groupProvider.GetGroupService()
	err := groupService.DeleteGroup(id)
	if err != nil {
		if errors.Is(err, model.ErrCannotDeleteGroupWithChildren) {
			http.Error(w, "Bad Request: Cannot delete group with child groups.", http.StatusBadRequest)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log the group response
	logger.Debug("Group DELETE response sent", log.String("group id", id))
}
