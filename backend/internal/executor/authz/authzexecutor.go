/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

// Package authz provides the authorization executor for flow-based permission evaluation.
package authz

import (
	"encoding/json"

	authzsvc "github.com/asgardeo/thunder/internal/authz"
	flowconst "github.com/asgardeo/thunder/internal/flow/common/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/common/model"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const (
	loggerComponentName      = "AuthorizationExecutor"
	userAttributeGroups      = "groups"
	authorizedPermissionsKey = "authorized_permissions"
	requestedPermissionsKey  = "requested_permissions"
)

// AuthorizationExecutor implements the ExecutorInterface for performing authorization checks
// during flow execution. It enriches the flow context with authorized permissions.
type AuthorizationExecutor struct {
	internal     flowmodel.Executor
	authzService authzsvc.AuthorizationServiceInterface
}

var _ flowmodel.ExecutorInterface = (*AuthorizationExecutor)(nil)

// NewAuthorizationExecutor creates a new instance of AuthorizationExecutor.
func NewAuthorizationExecutor(id, name string, properties map[string]string) *AuthorizationExecutor {
	return &AuthorizationExecutor{
		internal: *flowmodel.NewExecutor(
			id, name, flowconst.ExecutorTypeUtility,
			[]flowmodel.InputData{}, []flowmodel.InputData{}, properties,
		),
		authzService: authzsvc.GetAuthorizationService(),
	}
}

// Execute executes the authorization logic by determining required permissions based on context,
// calling the authorization service, and storing authorized permissions in runtime data.
func (a *AuthorizationExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(
		log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, a.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing authorization executor")

	execResp := &flowmodel.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	if !ctx.AuthenticatedUser.IsAuthenticated {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "User is not authenticated"
		return execResp, nil
	}

	// Determine required permissions
	requestedPerms := extractRequestedPermissions(ctx)

	if len(requestedPerms) == 0 {
		logger.Debug("No permissions to check, returning empty permissions")
		execResp.Status = flowconst.ExecComplete
		return execResp, nil
	}

	logger.Debug("Determined required permissions", log.Int("count", len(requestedPerms)))

	// Extract user ID and group IDs
	userID := ctx.AuthenticatedUser.UserID
	groupIDs := a.extractGroupIDs(ctx)

	logger.Debug("Calling authorization service",
		log.String("userID", userID),
		log.Int("groupCount", len(groupIDs)),
		log.Int("permissionCount", len(requestedPerms)))

	// Call authorization service
	authzReq := authzsvc.GetAuthorizedPermissionsRequest{
		UserID:               userID,
		GroupIDs:             groupIDs,
		RequestedPermissions: requestedPerms,
	}

	authzResp, svcErr := a.authzService.GetAuthorizedPermissions(authzReq)
	if svcErr != nil {
		logger.Error("Authorization service call failed", log.String("error", svcErr.Error))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Authorization validation failure"
		return execResp, nil
	}

	setAuthorizedPermissions(execResp, authzResp.AuthorizedPermissions)
	logger.Debug("Authorization completed successfully",
		log.Int("authorizedCount", len(authzResp.AuthorizedPermissions)))

	execResp.Status = flowconst.ExecComplete
	return execResp, nil
}

func extractRequestedPermissions(ctx *flowmodel.NodeContext) []string {
	requestedPermissions := ctx.RuntimeData[requestedPermissionsKey]
	if requestedPermissions != "" {
		return utils.ParseStringArray(requestedPermissions, " ")
	}
	requestedPermissions = ctx.UserInputData[requestedPermissionsKey]
	return utils.ParseStringArray(requestedPermissions, " ")
}

func setAuthorizedPermissions(execResp *flowmodel.ExecutorResponse, authorizedPermissions []string) {
	execResp.RuntimeData[authorizedPermissionsKey] = utils.StringifyStringArray(authorizedPermissions, " ")
}

// extractGroupIDs extracts group IDs from the authenticated user's attributes or runtime data.
func (a *AuthorizationExecutor) extractGroupIDs(ctx *flowmodel.NodeContext) []string {
	// Try to get groups from authenticated user attributes
	if groupsAttr, ok := ctx.AuthenticatedUser.Attributes[userAttributeGroups]; ok {
		// Handle different group attribute formats
		switch v := groupsAttr.(type) {
		case []string:
			return v
		case []interface{}:
			groups := make([]string, 0, len(v))
			for _, item := range v {
				if str, ok := item.(string); ok {
					groups = append(groups, str)
				}
			}
			return groups
		case string:
			// Single group as string
			return []string{v}
		}
	}

	// Try to get groups from runtime data (JSON array string)
	if groupsJSON, ok := ctx.RuntimeData[userAttributeGroups]; ok && groupsJSON != "" {
		var groups []string
		if err := json.Unmarshal([]byte(groupsJSON), &groups); err == nil {
			return groups
		}
	}

	// No groups found
	return []string{}
}

// GetID returns the ID of the executor.
func (a *AuthorizationExecutor) GetID() string {
	return a.internal.GetID()
}

// GetName returns the name of the executor.
func (a *AuthorizationExecutor) GetName() string {
	return a.internal.GetName()
}

// GetProperties returns the properties of the executor.
func (a *AuthorizationExecutor) GetProperties() flowmodel.ExecutorProperties {
	return a.internal.Properties
}

// GetDefaultExecutorInputs returns the default executor inputs.
func (a *AuthorizationExecutor) GetDefaultExecutorInputs() []flowmodel.InputData {
	return a.internal.GetDefaultExecutorInputs()
}

// GetPrerequisites returns the prerequisites for the executor.
func (a *AuthorizationExecutor) GetPrerequisites() []flowmodel.InputData {
	return a.internal.GetPrerequisites()
}

// CheckInputData checks if required input data is available.
func (a *AuthorizationExecutor) CheckInputData(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) bool {
	return a.internal.CheckInputData(ctx, execResp)
}

// ValidatePrerequisites validates whether prerequisites are met.
func (a *AuthorizationExecutor) ValidatePrerequisites(
	ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse,
) bool {
	return a.internal.ValidatePrerequisites(ctx, execResp)
}

// GetUserIDFromContext retrieves the user ID from context.
func (a *AuthorizationExecutor) GetUserIDFromContext(ctx *flowmodel.NodeContext) (string, error) {
	return a.internal.GetUserIDFromContext(ctx)
}

// GetRequiredData returns the required input data.
func (a *AuthorizationExecutor) GetRequiredData(ctx *flowmodel.NodeContext) []flowmodel.InputData {
	return a.internal.GetRequiredData(ctx)
}
