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

// Package constants defines error constants for group management operations.
package constants

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for group management operations.
var (
	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "GRP-1001",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed or contains invalid data",
	}
	// ErrorMissingGroupID is the error returned when group ID is missing.
	ErrorMissingGroupID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "GRP-1002",
		Error:            "Invalid request format",
		ErrorDescription: "Group ID is required",
	}
	// ErrorGroupNotFound is the error returned when a group is not found.
	ErrorGroupNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "GRP-1003",
		Error:            "Group not found",
		ErrorDescription: "The group with the specified id does not exist",
	}
	// ErrorGroupNameConflict is the error returned when a group name conflicts.
	ErrorGroupNameConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "GRP-1004",
		Error:            "Group name conflict",
		ErrorDescription: "A group with the same name exists under the same parent",
	}
	// ErrorParentNotFound is the error returned when parent is not found.
	ErrorParentNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "GRP-1005",
		Error:            "Parent not found",
		ErrorDescription: "Parent group or organization unit not found",
	}
	// ErrorCannotDeleteGroup is the error returned when group cannot be deleted.
	ErrorCannotDeleteGroup = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "GRP-1006",
		Error:            "Cannot delete group",
		ErrorDescription: "Cannot delete group with child groups",
	}
	// ErrorInvalidUserID is the error returned when user ID is invalid.
	ErrorInvalidUserID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "GRP-1007",
		Error:            "Invalid user ID",
		ErrorDescription: "One or more user IDs in the request do not exist",
	}
	// ErrorInvalidLimit is the error returned when limit parameter is invalid.
	ErrorInvalidLimit = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "GRP-1011",
		Error:            "Invalid limit parameter",
		ErrorDescription: "The limit parameter must be a positive integer",
	}
	// ErrorInvalidOffset is the error returned when offset parameter is invalid.
	ErrorInvalidOffset = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "GRP-1012",
		Error:            "Invalid offset parameter",
		ErrorDescription: "The offset parameter must be a non-negative integer",
	}
)

// Server errors for group management operations.
var (
	// ErrorInternalServerError is the error returned when an internal server error occurs.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "GRP-5000",
		Error:            "Internal server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)
