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

// Package constants defines the constants used in the flow execution service and engine.
package constants

// GraphType defines the type of graph used in the flow execution.
type GraphType string

const (
	// GraphTypeAuthentication represents a graph type for authentication flows.
	GraphTypeAuthentication GraphType = "AUTHENTICATION"
	// GraphTypeRegistration represents a graph type for registration flows.
	GraphTypeRegistration GraphType = "REGISTRATION"
)

// FlowStatus defines the status of a flow execution.
type FlowStatus string

const (
	// FlowStatusComplete indicates that the flow execution is complete.
	FlowStatusComplete FlowStatus = "COMPLETE"
	// FlowStatusIncomplete indicates that the flow execution is incomplete.
	FlowStatusIncomplete FlowStatus = "INCOMPLETE"
	// FlowStatusError indicates that there was an error during the flow execution.
	FlowStatusError FlowStatus = "ERROR"
)

// FlowStepType defines the type of a step in the flow execution.
type FlowStepType string

const (
	// StepTypeView represents a step in the flow that requires user interaction.
	StepTypeView FlowStepType = "VIEW"
	// StepTypeRedirection represents a step in the flow that redirects the user to another URL.
	StepTypeRedirection FlowStepType = "REDIRECTION"
)

// NodeType defines the node types in the flow execution.
type NodeType string

const (
	// NodeTypeAuthSuccess represents a node that does auth assertion
	NodeTypeAuthSuccess NodeType = "AUTHENTICATION_SUCCESS"
	// NodeTypeTaskExecution represents a task execution node
	NodeTypeTaskExecution NodeType = "TASK_EXECUTION"
	// NodeTypePromptOnly represents a prompt-only node
	NodeTypePromptOnly NodeType = "PROMPT_ONLY"
	// NodeTypeDecision represents a decision node
	NodeTypeDecision NodeType = "DECISION"
)

// NodeStatus defines the status of a node in the flow execution.
type NodeStatus string

const (
	// NodeStatusComplete indicates that the node has completed its execution successfully.
	NodeStatusComplete NodeStatus = "COMPLETE"
	// NodeStatusIncomplete indicates that the node has not completed its execution.
	NodeStatusIncomplete NodeStatus = "INCOMPLETE"
	// NodeStatusFailure indicates that the node has failed during its execution.
	NodeStatusFailure NodeStatus = "FAILURE"
)

// NodeResponseType defines the type of response from a node in the flow execution.
type NodeResponseType string

const (
	// NodeResponseTypeView indicates that the node response is a view type, requiring user interaction.
	NodeResponseTypeView NodeResponseType = "VIEW"
	// NodeResponseTypeRedirection indicates that the node response is a redirection type, redirecting to another URL.
	NodeResponseTypeRedirection NodeResponseType = "REDIRECTION"
	// NodeResponseTypeRetry indicates that the node response is a retry type, indicating a retry action.
	NodeResponseTypeRetry NodeResponseType = "RETRY"
)

// ExecutorStatus defines the status of an executor in the flow execution.
type ExecutorStatus string

const (
	// ExecComplete indicates that the executor has completed its execution successfully.
	ExecComplete ExecutorStatus = "COMPLETE"
	// ExecUserInputRequired indicates that the executor requires user input to proceed.
	ExecUserInputRequired ExecutorStatus = "USER_INPUT_REQUIRED"
	// ExecExternalRedirection indicates that the executor is redirecting to an external URL.
	ExecExternalRedirection ExecutorStatus = "EXTERNAL_REDIRECTION"
	// ExecFailure indicates that the executor has failed during its execution.
	ExecFailure ExecutorStatus = "FAILURE"
	// ExecRetry indicates that the executor is retrying its execution.
	ExecRetry ExecutorStatus = "RETRY"
)

const (
	// DataIDPName is the key used for the identity provider name in the flow response.
	DataIDPName = "idpName"
)

// ActionType defines the type of action that can be performed in a decision node.
type ActionType string

const (
	// ActionTypeView indicates that the action is a view type, requiring user selection.
	ActionTypeView ActionType = "VIEW"
	// ActionTypeUserInput indicates that the action requires user input to proceed.
	ActionTypeUserInput ActionType = "USER_INPUT"
)
