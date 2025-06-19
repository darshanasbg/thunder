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

// Package services handles the registration of routes and services for the system.
//
//nolint:dupl // Ignoring false positive duplicate code
package services

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/group/handler"
	"github.com/asgardeo/thunder/internal/system/server"
)

// GroupService is the service for group management operations.
type GroupService struct {
	groupHandler *handler.GroupHandler
}

// NewGroupService creates a new instance of GroupService.
func NewGroupService(mux *http.ServeMux) *GroupService {
	instance := &GroupService{
		groupHandler: &handler.GroupHandler{},
	}
	instance.RegisterRoutes(mux)

	return instance
}

// RegisterRoutes registers the routes for group management operations.
func (s *GroupService) RegisterRoutes(mux *http.ServeMux) {
	opts1 := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "GET, POST",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	server.WrapHandleFunction(mux, "POST /groups", &opts1, s.groupHandler.HandleGroupPostRequest)
	server.WrapHandleFunction(mux, "GET /groups", &opts1, s.groupHandler.HandleGroupListRequest)
	server.WrapHandleFunction(mux, "OPTIONS /groups", &opts1, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	opts2 := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "GET, PUT, DELETE",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	server.WrapHandleFunction(mux, "GET /groups/", &opts2, s.groupHandler.HandleGroupGetRequest)
	server.WrapHandleFunction(mux, "PUT /groups/", &opts2, s.groupHandler.HandleGroupPutRequest)
	server.WrapHandleFunction(mux, "DELETE /groups/", &opts2, s.groupHandler.HandleGroupDeleteRequest)
	server.WrapHandleFunction(mux, "OPTIONS /groups/", &opts2, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}
