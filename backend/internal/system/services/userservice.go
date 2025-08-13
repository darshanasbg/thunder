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

	"github.com/asgardeo/thunder/internal/system/server"
	"github.com/asgardeo/thunder/internal/user/handler"
	"github.com/asgardeo/thunder/internal/user/provider"
)

// UserService is the service for user management operations.
type UserService struct {
	ServerOpsService server.ServerOperationServiceInterface
	userHandler      *handler.UserHandler
}

// NewUserService creates a new instance of UserService.
func NewUserService(mux *http.ServeMux) ServiceInterface {
	userProvider := provider.NewUserProvider()
	instance := &UserService{
		ServerOpsService: server.NewServerOperationService(),
		userHandler:      handler.NewUserHandler(userProvider),
	}
	instance.RegisterRoutes(mux)

	return instance
}

// RegisterRoutes registers the routes for user management operations.
func (s *UserService) RegisterRoutes(mux *http.ServeMux) {
	opts1 := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "GET, POST",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	s.ServerOpsService.WrapHandleFunction(mux, "POST /users", &opts1, s.userHandler.HandleUserPostRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "GET /users", &opts1, s.userHandler.HandleUserListRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /users", &opts1,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

	opts2 := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "GET, PUT, DELETE",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	s.ServerOpsService.WrapHandleFunction(mux, "GET /users/", &opts2, s.userHandler.HandleUserGetRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "PUT /users/", &opts2, s.userHandler.HandleUserPutRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "DELETE /users/", &opts2, s.userHandler.HandleUserDeleteRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /users/", &opts2,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

	opts3 := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "GET, POST",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	s.ServerOpsService.WrapHandleFunction(mux, "GET /users/tree/{path...}", &opts3,
		s.userHandler.HandleUserListByPathRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "POST /users/tree/{path...}", &opts3,
		s.userHandler.HandleUserPostByPathRequest)
	s.ServerOpsService.WrapHandleFunction(
		mux,
		"OPTIONS /users/tree/{path...}",
		&opts3,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	)

	opts4 := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "POST",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	s.ServerOpsService.WrapHandleFunction(
		mux, "POST /users/authenticate", &opts4, s.userHandler.HandleUserAuthenticateRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /users/authenticate", &opts4,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
}
