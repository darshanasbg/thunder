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

// Package service provides the implementation for user management operations.
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	ouconstants "github.com/asgardeo/thunder/internal/ou/constants"
	ouservice "github.com/asgardeo/thunder/internal/ou/service"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/user/constants"
	"github.com/asgardeo/thunder/internal/user/model"
	"github.com/asgardeo/thunder/internal/user/store"
)

const loggerComponentName = "UserService"

// UserServiceInterface defines the interface for the user service.
type UserServiceInterface interface {
	GetUserList(limit, offset int) (*model.UserListResponse, *serviceerror.ServiceError)
	GetUsersByPath(handlePath string, limit, offset int) (*model.UserListResponse, *serviceerror.ServiceError)
	CreateUser(user *model.User) (*model.User, *serviceerror.ServiceError)
	CreateUserByPath(handlePath string, request model.CreateUserByPathRequest) (*model.User, *serviceerror.ServiceError)
	GetUser(userID string) (*model.User, *serviceerror.ServiceError)
	UpdateUser(userID string, user *model.User) (*model.User, *serviceerror.ServiceError)
	DeleteUser(userID string) *serviceerror.ServiceError
	IdentifyUser(filters map[string]interface{}) (*string, *serviceerror.ServiceError)
	VerifyUser(userID, credType, credValue string) (*model.User, *serviceerror.ServiceError)
	AuthenticateUser(request model.AuthenticateUserRequest) (*model.AuthenticateUserResponse, *serviceerror.ServiceError)
	ValidateUserIDs(userIDs []string) ([]string, *serviceerror.ServiceError)
}

// UserService is the default implementation of the UserServiceInterface.
type UserService struct {
	ouService ouservice.OrganizationUnitServiceInterface
}

// GetUserService creates a new instance of UserService.
func GetUserService() UserServiceInterface {
	return &UserService{
		ouService: ouservice.GetOrganizationUnitService(),
	}
}

// GetUserList lists the users.
func (as *UserService) GetUserList(limit, offset int) (*model.UserListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := store.GetUserListCount()
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list count", err)
	}

	users, err := store.GetUserList(limit, offset)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to get user list", err)
	}

	var links []model.Link
	if offset+limit < totalCount {
		links = append(links, model.Link{
			Href: fmt.Sprintf("users?offset=%d&limit=%d", offset+limit, limit),
			Rel:  "next",
		})
	}

	response := &model.UserListResponse{
		TotalResults: totalCount,
		StartIndex:   offset + 1,
		Count:        len(users),
		Users:        users,
		Links:        links,
	}

	return response, nil
}

// GetUsersByPath retrieves a list of users by hierarchical handle path.
func (as *UserService) GetUsersByPath(
	handlePath string, limit, offset int,
) (*model.UserListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Getting users by path", log.String("path", handlePath))

	serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, svcErr := as.ouService.GetOrganizationUnitByPath(handlePath)
	if svcErr != nil {
		if svcErr.Code == ouconstants.ErrorOrganizationUnitNotFound.Code {
			return nil, &constants.ErrorOrganizationUnitNotFound
		}
		return nil, logErrorAndReturnServerError(logger,
			"Failed to get organization unit using the handle path from organization service", nil)
	}
	organizationUnitID := ou.ID

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	ouResponse, svcErr := as.ouService.GetOrganizationUnitUsers(organizationUnitID, limit, offset)
	if svcErr != nil {
		return nil, svcErr
	}

	users := make([]model.User, len(ouResponse.Users))
	for i, ouUser := range ouResponse.Users {
		users[i] = model.User{
			ID: ouUser.ID,
		}
	}

	response := &model.UserListResponse{
		TotalResults: ouResponse.TotalResults,
		StartIndex:   ouResponse.StartIndex,
		Count:        ouResponse.Count,
		Users:        users,
		Links:        buildPaginationLinks(handlePath, limit, offset, ouResponse.TotalResults),
	}

	return response, nil
}

// CreateUser creates the user.
func (as *UserService) CreateUser(user *model.User) (*model.User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if user == nil {
		return nil, &constants.ErrorInvalidRequestFormat
	}

	user.ID = utils.GenerateUUID()

	credentials, err := extractCredentials(user)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to create user DTO", err)
	}

	// Create the user in the database.
	err = store.CreateUser(*user, *credentials)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to create user", err)
	}

	logger.Debug("Successfully created user", log.String("id", user.ID))
	return user, nil
}

// CreateUserByPath creates a new user under the organization unit specified by the handle path.
func (as *UserService) CreateUserByPath(
	handlePath string, request model.CreateUserByPathRequest,
) (*model.User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating user by path", log.String("path", handlePath), log.String("type", request.Type))

	serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, svcErr := as.ouService.GetOrganizationUnitByPath(handlePath)
	if svcErr != nil {
		if svcErr.Code == ouconstants.ErrorOrganizationUnitNotFound.Code {
			return nil, &constants.ErrorOrganizationUnitNotFound
		}
		return nil, logErrorAndReturnServerError(logger,
			"Failed to get organization unit using the handle path from organization service", nil)
	}

	user := &model.User{
		OrganizationUnit: ou.ID,
		Type:             request.Type,
		Attributes:       request.Attributes,
	}

	return as.CreateUser(user)
}

// extractCredentials extracts the credentials from the user attributes and returns a Credentials object.
func extractCredentials(user *model.User) (*model.Credentials, error) {
	var attrsMap map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrsMap); err != nil {
		return nil, err
	}

	if pw, ok := attrsMap["password"].(string); ok {
		// Generate a salt
		pwSalt, err := hash.GenerateSalt()
		if err != nil {
			return nil, err
		}

		// Hash the password with the salt
		pwHash, err := hash.HashStringWithSalt(pw, pwSalt)
		if err != nil {
			return nil, err
		}

		delete(attrsMap, "password")
		updatedAttrs, err := json.Marshal(attrsMap)
		if err != nil {
			return nil, err
		}
		user.Attributes = updatedAttrs

		credentials := model.Credentials{
			CredentialType: "password",
			StorageType:    "hash",
			StorageAlgo:    "SHA-256",
			Value:          pwHash,
			Salt:           pwSalt,
		}

		return &credentials, nil
	}

	return &model.Credentials{}, nil
}

// GetUser get the user for given user id.
func (as *UserService) GetUser(userID string) (*model.User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving user", log.String("id", userID))

	if userID == "" {
		return nil, &constants.ErrorMissingUserID
	}

	user, err := store.GetUser(userID)
	if err != nil {
		if errors.Is(err, constants.ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &constants.ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to retrieve user", err, log.String("id", userID))
	}

	logger.Debug("Successfully retrieved user", log.String("id", userID))
	return &user, nil
}

// UpdateUser update the user for given user id.
func (as *UserService) UpdateUser(userID string, user *model.User) (*model.User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating user", log.String("id", userID))

	if userID == "" {
		return nil, &constants.ErrorMissingUserID
	}

	if user == nil {
		return nil, &constants.ErrorInvalidRequestFormat
	}

	err := store.UpdateUser(user)
	if err != nil {
		if errors.Is(err, constants.ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &constants.ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to update user", err, log.String("id", userID))
	}

	logger.Debug("Successfully updated user", log.String("id", userID))
	return user, nil
}

// DeleteUser delete the user for given user id.
func (as *UserService) DeleteUser(userID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Deleting user", log.String("id", userID))

	if userID == "" {
		return &constants.ErrorMissingUserID
	}

	err := store.DeleteUser(userID)
	if err != nil {
		if errors.Is(err, constants.ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return &constants.ErrorUserNotFound
		}
		return logErrorAndReturnServerError(logger, "Failed to delete user", err, log.String("id", userID))
	}

	logger.Debug("Successfully deleted user", log.String("id", userID))
	return nil
}

// IdentifyUser identifies a user with the given filters.
func (as *UserService) IdentifyUser(filters map[string]interface{}) (*string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(filters) == 0 {
		return nil, &constants.ErrorInvalidRequestFormat
	}

	userID, err := store.IdentifyUser(filters)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			logger.Debug("User not found with provided filters")
			return nil, &constants.ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to identify user", err)
	}

	return userID, nil
}

// VerifyUser validate the specified user with the given credentials.
func (as *UserService) VerifyUser(userID, credType, credValue string) (*model.User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if userID == "" {
		return nil, &constants.ErrorMissingUserID
	}

	if credType == "" {
		return nil, &constants.ErrorInvalidRequestFormat
	}

	if credValue == "" {
		return nil, &constants.ErrorInvalidRequestFormat
	}

	user, credentials, err := store.VerifyUser(userID)
	if err != nil {
		if errors.Is(err, constants.ErrUserNotFound) {
			logger.Debug("User not found", log.String("id", userID))
			return nil, &constants.ErrorUserNotFound
		}
		return nil, logErrorAndReturnServerError(logger, "Failed to verify user", err, log.String("id", userID))
	}

	// Fix the comparison to check for an empty Credentials struct instead of nil.
	if credentials == (model.Credentials{}) {
		return nil, logErrorAndReturnServerError(logger, "Credentials not found", nil, log.String("userID", userID))
	}
	if credentials.CredentialType == "" || credentials.Value == "" || credentials.Salt == "" {
		return nil, logErrorAndReturnServerError(logger, "Incomplete credentials", nil, log.String("userID", userID))
	}

	hashToCompare, err := hash.HashStringWithSalt(credValue, credentials.Salt)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to hash credential value", err)
	}
	if credentials.CredentialType != credType {
		logger.Debug("Invalid credential type",
			log.String("userID", userID),
			log.String("expectedType", credType),
			log.String("actualType", credentials.CredentialType))
		return nil, &constants.ErrorAuthenticationFailed
	}
	if credentials.Value != hashToCompare {
		logger.Debug("Invalid credentials provided", log.String("userID", userID))
		return nil, &constants.ErrorAuthenticationFailed
	}

	logger.Debug("Successfully verified user", log.String("id", userID))
	return &user, nil
}

// AuthenticateUser authenticates a user by combining identify and verify operations.
func (as *UserService) AuthenticateUser(
	request model.AuthenticateUserRequest,
) (*model.AuthenticateUserResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(request) == 0 {
		return nil, &constants.ErrorInvalidRequestFormat
	}

	identifyFilters := make(map[string]interface{})
	credentials := make(map[string]interface{})

	credentialFields := map[string]bool{
		"password": true,
		"pin":      true,
		"secret":   true,
	}

	for key, value := range request {
		if credentialFields[key] {
			credentials[key] = value
		} else {
			identifyFilters[key] = value
		}
	}

	if len(identifyFilters) == 0 {
		return nil, &constants.ErrorMissingRequiredFields
	}
	if len(credentials) == 0 {
		return nil, &constants.ErrorMissingCredentials
	}

	userID, svcErr := as.IdentifyUser(identifyFilters)
	if svcErr != nil {
		if svcErr.Code == constants.ErrorUserNotFound.Code {
			return nil, &constants.ErrorUserNotFound
		}
		return nil, svcErr
	}

	var credType, credValue string
	if password, exists := credentials["password"]; exists {
		// Only support password authentication
		credType = "password"
		if passwordStr, ok := password.(string); ok {
			credValue = passwordStr
		} else {
			return nil, &constants.ErrorInvalidRequestFormat
		}
	} else {
		return nil, &constants.ErrorMissingCredentials
	}

	user, svcErr := as.VerifyUser(*userID, credType, credValue)
	if svcErr != nil {
		return nil, svcErr
	}

	logger.Info("User authenticated successfully", log.String("userID", *userID))
	return &model.AuthenticateUserResponse{
		ID:               user.ID,
		Type:             user.Type,
		OrganizationUnit: user.OrganizationUnit,
	}, nil
}

// ValidateUserIDs validates that all provided user IDs exist.
func (as *UserService) ValidateUserIDs(userIDs []string) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(userIDs) == 0 {
		return []string{}, nil
	}

	invalidUserIDs, err := store.ValidateUserIDs(userIDs)
	if err != nil {
		return nil, logErrorAndReturnServerError(logger, "Failed to validate user IDs", err)
	}

	return invalidUserIDs, nil
}

// validateAndProcessHandlePath validates and processes the handle path.
func validateAndProcessHandlePath(handlePath string) *serviceerror.ServiceError {
	if strings.TrimSpace(handlePath) == "" {
		return &constants.ErrorInvalidHandlePath
	}

	handles := strings.Split(strings.Trim(handlePath, "/"), "/")
	if len(handles) == 0 {
		return &constants.ErrorInvalidHandlePath
	}

	for _, handle := range handles {
		if strings.TrimSpace(handle) == "" {
			return &constants.ErrorInvalidHandlePath
		}
	}
	return nil
}

// validatePaginationParams validates pagination parameters.
func validatePaginationParams(limit, offset int) *serviceerror.ServiceError {
	if limit < 1 || limit > serverconst.MaxPageSize {
		return &constants.ErrorInvalidLimit
	}
	if offset < 0 {
		return &constants.ErrorInvalidOffset
	}
	return nil
}

// logErrorAndReturnServerError logs the error and returns a server error.
func logErrorAndReturnServerError(
	logger *log.Logger,
	message string,
	err error,
	additionalFields ...log.Field,
) *serviceerror.ServiceError {
	fields := additionalFields
	if err != nil {
		fields = append(fields, log.Error(err))
	}
	logger.Error(message, fields...)
	return &constants.ErrorInternalServerError
}

// buildPaginationLinks builds pagination links for user responses.
func buildPaginationLinks(handlePath string, limit, offset, totalResults int) []model.Link {
	links := make([]model.Link, 0)

	if offset > 0 {
		links = append(links, model.Link{
			Href: fmt.Sprintf("users/tree/%s?offset=0&limit=%d", handlePath, limit),
			Rel:  "first",
		})

		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, model.Link{
			Href: fmt.Sprintf("users/tree/%s?offset=%d&limit=%d", handlePath, prevOffset, limit),
			Rel:  "prev",
		})
	}

	if offset+limit < totalResults {
		nextOffset := offset + limit
		links = append(links, model.Link{
			Href: fmt.Sprintf("users/tree/%s?offset=%d&limit=%d", handlePath, nextOffset, limit),
			Rel:  "next",
		})
	}

	lastPageOffset := ((totalResults - 1) / limit) * limit
	if offset < lastPageOffset {
		links = append(links, model.Link{
			Href: fmt.Sprintf("users/tree/%s?offset=%d&limit=%d", handlePath, lastPageOffset, limit),
			Rel:  "last",
		})
	}

	return links
}
