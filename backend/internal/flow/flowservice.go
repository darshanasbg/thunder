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

// Package flow provides the FlowService interface and its implementation.
package flow

import (
	"errors"
	"sync"

	appservice "github.com/asgardeo/thunder/internal/application/service"
	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/dao"
	"github.com/asgardeo/thunder/internal/flow/engine"
	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

var (
	instance *FlowService
	once     sync.Once
)

// FlowServiceInterface defines the interface for flow orchestration and acts as the entry point for flow execution
type FlowServiceInterface interface {
	Init() error
	Execute(appID, flowID, actionID string, inputData map[string]string) (*model.FlowStep, *serviceerror.ServiceError)
}

// FlowService is the implementation of FlowServiceInterface
type FlowService struct{}

// GetFlowService returns a singleton instance of FlowService
func GetFlowService() FlowServiceInterface {
	once.Do(func() {
		instance = &FlowService{}
	})
	return instance
}

// Init initializes the FlowService by loading the necessary components.
func (s *FlowService) Init() error {
	flowDAO := dao.GetFlowDAO()
	if err := flowDAO.Init(); err != nil {
		return errors.New("failed to initialize flow service: " + err.Error())
	}

	if err := validateDefaultFlowConfigs(flowDAO); err != nil {
		return errors.New("default flow config validation failed: " + err.Error())
	}

	return nil
}

// Execute executes a flow with the given data
func (s *FlowService) Execute(appID, flowID, actionID string,
	inputData map[string]string) (*model.FlowStep, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowService"))

	var context *model.EngineContext
	var loadErr *serviceerror.ServiceError

	if isNewFlow(flowID, actionID) {
		context, loadErr = s.loadNewContext(appID, actionID, inputData, logger)
	} else {
		context, loadErr = s.loadPrevContext(flowID, actionID, inputData, logger)
	}

	if loadErr != nil {
		return nil, loadErr
	}

	flowStep, flowErr := engine.GetFlowEngine().Execute(context)
	if flowErr != nil {
		s.removeContext(flowID, logger)
		return nil, flowErr
	}

	if isComplete(flowStep) {
		s.removeContext(flowID, logger)
	} else {
		s.updateContext(context, &flowStep, logger)
	}

	return &flowStep, nil
}

// initContext initializes a new flow context with the given details.
func (s *FlowService) loadNewContext(appID, actionID string,
	inputData map[string]string, logger *log.Logger) (*model.EngineContext, *serviceerror.ServiceError) {
	ctx, err := s.initContext(appID, logger)
	if err != nil {
		return nil, err
	}

	prepareContext(ctx, actionID, inputData)
	return ctx, nil
}

// initContext initializes a new flow context with the given details.
func (s *FlowService) initContext(appID string, logger *log.Logger) (*model.EngineContext,
	*serviceerror.ServiceError) {
	graphID, svcErr := validateApplication(appID)
	if svcErr != nil {
		return nil, svcErr
	}

	ctx := model.EngineContext{}
	flowID := sysutils.GenerateUUID()
	ctx.FlowID = flowID

	flowDAO := dao.GetFlowDAO()
	graph, ok := flowDAO.GetGraph(graphID)
	if !ok {
		logger.Error("Graph not found for the graph id")
		return nil, &constants.ErrorFlowGraphNotFound
	}
	ctx.FlowType = graph.GetType()
	ctx.Graph = graph
	ctx.AppID = appID

	return &ctx, nil
}

// loadPrevContext retrieves the flow context from the store based on the given details.
func (s *FlowService) loadPrevContext(flowID, actionID string, inputData map[string]string,
	logger *log.Logger) (*model.EngineContext, *serviceerror.ServiceError) {
	ctx, err := s.loadContextFromStore(flowID, logger)
	if err != nil {
		return nil, err
	}

	prepareContext(ctx, actionID, inputData)
	return ctx, nil
}

// loadContextFromStore retrieves the flow context from the store based on the given details.
func (s *FlowService) loadContextFromStore(flowID string, logger *log.Logger) (*model.EngineContext,
	*serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &constants.ErrorInvalidFlowID
	}

	flowDAO := dao.GetFlowDAO()
	ctx, exists := flowDAO.GetContextFromStore(flowID)
	if !exists {
		return nil, &constants.ErrorInvalidFlowID
	}
	s.removeContext(flowID, logger)

	return &ctx, nil
}

// removeContext removes the flow context from the store.
func (s *FlowService) removeContext(flowID string, logger *log.Logger) {
	flowDAO := dao.GetFlowDAO()
	err := flowDAO.RemoveContextFromStore(flowID)
	if err != nil {
		logger.Error("Failed to remove flow context from the store", log.String("flowID", flowID), log.Error(err))
		return
	}
	logger.Debug("Flow context removed from the store", log.String("flowID", flowID))
}

// updateContext updates the flow context in the store based on the flow step status.
func (s *FlowService) updateContext(ctx *model.EngineContext, flowStep *model.FlowStep, logger *log.Logger) {
	if flowStep.Status == constants.FlowStatusComplete {
		s.removeContext(ctx.FlowID, logger)
	} else {
		logger.Debug("Flow execution is incomplete, storing the flow context",
			log.String("flowID", ctx.FlowID))

		flowDAO := dao.GetFlowDAO()
		if err := flowDAO.StoreContextInStore(ctx.FlowID, *ctx); err != nil {
			logger.Error("Failed to store flow context in the store", log.String("flowID", ctx.FlowID), log.Error(err))
			return
		}
		logger.Debug("Flow context stored in the store", log.String("flowID", ctx.FlowID))
	}
}

// validateDefaultFlowConfigs validates the default flow configurations.
func validateDefaultFlowConfigs(flowDAO dao.FlowDAOInterface) error {
	flowConfig := config.GetThunderRuntime().Config.Flow

	// Validate auth flow.
	if flowConfig.Authn.DefaultFlow == "" {
		return errors.New("default authentication flow is not configured")
	}
	if !flowDAO.IsValidGraphID(flowConfig.Authn.DefaultFlow) {
		return errors.New("default authentication flow graph ID is invalid")
	}

	return nil
}

// isNewFlow checks if the flow is a new flow based on the provided flowID and actionID.
func isNewFlow(flowID string, actionID string) bool {
	return flowID == "" && actionID == ""
}

// isComplete checks if the flow step status indicates completion.
func isComplete(step model.FlowStep) bool {
	return step.Status == constants.FlowStatusComplete
}

// prepareContext prepares the flow context by merging any data.
func prepareContext(ctx *model.EngineContext, actionID string, inputData map[string]string) {
	// Append any input data present to the context
	if len(inputData) > 0 {
		ctx.UserInputData = sysutils.MergeStringMaps(ctx.UserInputData, inputData)
	}

	if ctx.UserInputData == nil {
		ctx.UserInputData = make(map[string]string)
	}
	if ctx.RuntimeData == nil {
		ctx.RuntimeData = make(map[string]string)
	}

	// Set the action ID if provided
	if actionID != "" {
		ctx.CurrentActionID = actionID
	}
}

// validateApplication checks if the provided application ID is valid and returns the associated auth flow graph ID.
func validateApplication(appID string) (string, *serviceerror.ServiceError) {
	if appID == "" {
		return "", &constants.ErrorInvalidAppID
	}

	appSvc := appservice.GetApplicationService()
	app, err := appSvc.GetApplication(appID)
	if err != nil {
		return "", &constants.ErrorInvalidAppID
	}
	if app == nil {
		return "", &constants.ErrorInvalidAppID
	}

	// At this point, we assume auth flow graph is configured for the application.
	if app.AuthFlowGraphID == "" {
		return "", &constants.ErrorAuthFlowNotConfiguredForApplication
	}

	return app.AuthFlowGraphID, nil
}
