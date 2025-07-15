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

// Package handler provides the implementation for organization unit management operations.
package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/asgardeo/thunder/internal/ou/constants"
	"github.com/asgardeo/thunder/internal/ou/model"
	"github.com/asgardeo/thunder/internal/ou/service"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "OrganizationUnitHandler"
const limitDefault = 30

// OrganizationUnitHandler is the handler for organization unit management operations.
type OrganizationUnitHandler struct {
	service service.OrganizationUnitServiceInterface
}

// NewOrganizationUnitHandler creates a new instance of OrganizationUnitHandler
func NewOrganizationUnitHandler() *OrganizationUnitHandler {
	return &OrganizationUnitHandler{
		service: service.GetOrganizationUnitService(),
	}
}

// HandleOUListRequest handles the list organization units request.
func (ouh *OrganizationUnitHandler) HandleOUListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		ouh.handleError(w, logger, svcErr)
		return
	}

	if limit == 0 {
		limit = limitDefault
	}

	ouListResponse, svcErr := ouh.service.GetOrganizationUnitList(limit, offset)
	if svcErr != nil {
		ouh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(ouListResponse); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully listed organization units with pagination",
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", ouListResponse.TotalResults),
		log.Int("count", ouListResponse.Count))
}

// HandleOUPostRequest handles the create organization unit request.
func (ouh *OrganizationUnitHandler) HandleOUPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	createRequest, err := sysutils.DecodeJSONBody[model.OrganizationUnitRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidRequestFormat.Code,
			Message:     constants.ErrorInvalidRequestFormat.Error,
			Description: "Failed to parse request body: " + err.Error(),
		}

		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	sanitizedRequest := ouh.sanitizeOrganizationUnitRequest(*createRequest)

	createdOU, svcErr := ouh.service.CreateOrganizationUnit(sanitizedRequest)
	if svcErr != nil {
		ouh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdOU); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully created organization unit", log.String("ouId", createdOU.ID))
}

// HandleOUGetRequest handles the get organization unit by id request.
func (ouh *OrganizationUnitHandler) HandleOUGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorMissingOrganizationUnitID.Code,
			Message:     constants.ErrorMissingOrganizationUnitID.Error,
			Description: constants.ErrorMissingOrganizationUnitID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	ou, svcErr := ouh.service.GetOrganizationUnit(id)
	if svcErr != nil {
		ouh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(ou); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully retrieved organization unit", log.String("ouId", id))
}

// HandleOUPutRequest handles the update organization unit request.
func (ouh *OrganizationUnitHandler) HandleOUPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorMissingOrganizationUnitID.Code,
			Message:     constants.ErrorMissingOrganizationUnitID.Error,
			Description: constants.ErrorMissingOrganizationUnitID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	updateRequest, err := sysutils.DecodeJSONBody[model.OrganizationUnitRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidRequestFormat.Code,
			Message:     constants.ErrorInvalidRequestFormat.Error,
			Description: "Failed to parse request body: " + err.Error(),
		}

		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	sanitizedRequest := ouh.sanitizeOrganizationUnitRequest(*updateRequest)

	ou, svcErr := ouh.service.UpdateOrganizationUnit(id, sanitizedRequest)
	if svcErr != nil {
		ouh.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(ou); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Successfully updated organization unit", log.String("ouId", id))
}

// HandleOUDeleteRequest handles the delete organization unit request.
func (ouh *OrganizationUnitHandler) HandleOUDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	id := r.PathValue("id")
	if id == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorMissingOrganizationUnitID.Code,
			Message:     constants.ErrorMissingOrganizationUnitID.Error,
			Description: constants.ErrorMissingOrganizationUnitID.ErrorDescription,
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	svcErr := ouh.service.DeleteOrganizationUnit(id)
	if svcErr != nil {
		ouh.handleError(w, logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Debug("Successfully deleted organization unit", log.String("ouId", id))
}

// handleError handles service errors and returns appropriate HTTP responses.
func (ouh *OrganizationUnitHandler) handleError(w http.ResponseWriter, logger *log.Logger,
	svcErr *serviceerror.ServiceError) {
	// TODO: Revisit
	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)

	var statusCode int
	switch svcErr.Type {
	case serviceerror.ClientErrorType:
		statusCode = http.StatusBadRequest
		if svcErr.Code == constants.ErrorOrganizationUnitNotFound.Code {
			statusCode = http.StatusNotFound
		} else if svcErr.Code == constants.ErrorOrganizationUnitNameConflict.Code ||
			svcErr.Code == constants.ErrorOrganizationUnitHandleConflict.Code {
			statusCode = http.StatusConflict
		} else if svcErr.Code == constants.ErrorInvalidLimit.Code ||
			svcErr.Code == constants.ErrorInvalidOffset.Code {
			statusCode = http.StatusBadRequest
		}
	default:
		statusCode = http.StatusInternalServerError
	}

	w.WriteHeader(statusCode)

	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("Error encoding error response", log.Error(err))
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// sanitizeOrganizationUnitRequest sanitizes the create organization unit request input.
func (ouh *OrganizationUnitHandler) sanitizeOrganizationUnitRequest(
	request model.OrganizationUnitRequest,
) model.OrganizationUnitRequest {
	return model.OrganizationUnitRequest{
		Handle:      sysutils.SanitizeString(request.Handle),
		Name:        sysutils.SanitizeString(request.Name),
		Description: sysutils.SanitizeString(request.Description),
		Parent:      request.Parent,
	}
}

// parsePaginationParams parses limit and offset query parameters from the request.
func parsePaginationParams(query url.Values) (int, int, *serviceerror.ServiceError) {
	limit := 0
	offset := 0

	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err != nil {
			return 0, 0, &constants.ErrorInvalidLimit
		} else {
			limit = parsedLimit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err != nil {
			return 0, 0, &constants.ErrorInvalidOffset
		} else {
			offset = parsedOffset
		}
	}

	return limit, offset, nil
}
