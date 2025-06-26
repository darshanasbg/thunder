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

package model

import (
	"encoding/json"
	"errors"

	"github.com/asgardeo/thunder/internal/flow/constants"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// GraphInterface defines the graph structure
type GraphInterface interface {
	GetID() string
	GetType() constants.GraphType
	AddNode(node NodeInterface) error
	GetNode(nodeID string) (NodeInterface, bool)
	AddEdge(fromNodeID, toNodeID string) error
	GetNodes() map[string]NodeInterface
	GetEdges() map[string][]string
	GetStartNodeID() string
	GetStartNode() (NodeInterface, error)
	SetStartNode(startNodeID string) error
	ToJSON() (string, error)
}

// Graph implements the GraphInterface for the flow execution
type Graph struct {
	id          string
	_type       constants.GraphType
	nodes       map[string]NodeInterface
	edges       map[string][]string
	startNodeID string
}

// NewGraph creates a new Graph with a unique ID
func NewGraph(id string, _type constants.GraphType) GraphInterface {
	if id == "" {
		id = sysutils.GenerateUUID()
	}
	if _type == "" {
		_type = constants.GraphTypeAuthentication
	}

	return &Graph{
		id:    id,
		_type: _type,
		nodes: make(map[string]NodeInterface),
		edges: make(map[string][]string),
	}
}

// GetID returns the unique ID of the graph
func (g *Graph) GetID() string {
	return g.id
}

// GetType returns the type of the graph
func (g *Graph) GetType() constants.GraphType {
	return g._type
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(node NodeInterface) error {
	if node == nil {
		return errors.New("node cannot be nil")
	}

	g.nodes[node.GetID()] = node
	return nil
}

// GetNode retrieves a node by its ID
func (g *Graph) GetNode(nodeID string) (NodeInterface, bool) {
	if node, exists := g.nodes[nodeID]; exists {
		return node, true
	}
	return nil, false
}

// AddEdge adds an edge from one node to another
func (g *Graph) AddEdge(fromNodeID, toNodeID string) error {
	if fromNodeID == "" || toNodeID == "" {
		return errors.New("fromNodeID and toNodeID cannot be empty")
	}
	if _, exists := g.nodes[fromNodeID]; !exists {
		return errors.New("node with fromNodeID does not exist")
	}
	if _, exists := g.nodes[toNodeID]; !exists {
		return errors.New("node with toNodeID does not exist")
	}

	if _, exists := g.edges[fromNodeID]; !exists {
		g.edges[fromNodeID] = []string{}
	}
	g.edges[fromNodeID] = append(g.edges[fromNodeID], toNodeID)
	return nil
}

// GetNodes returns all nodes in the graph
func (g *Graph) GetNodes() map[string]NodeInterface {
	return g.nodes
}

// GetEdges returns all edges in the graph
func (g *Graph) GetEdges() map[string][]string {
	return g.edges
}

// GetStartNodeID returns the start node ID of the graph
func (g *Graph) GetStartNodeID() string {
	return g.startNodeID
}

// GetStartNode retrieves the start node of the graph
func (g *Graph) GetStartNode() (NodeInterface, error) {
	if g.startNodeID == "" {
		return nil, errors.New("start node not set for the graph")
	}
	node, exists := g.nodes[g.startNodeID]
	if !exists {
		return nil, errors.New("start node does not exist in the graph")
	}
	return node, nil
}

// SetStartNode sets the start node ID for the graph
func (g *Graph) SetStartNode(startNodeID string) error {
	node, exists := g.nodes[startNodeID]
	if !exists {
		return errors.New("node with startNodeID does not exist")
	}
	g.startNodeID = startNodeID
	node.SetAsStartNode()

	return nil
}

// ToJSON converts the graph to a JSON string representation
func (g *Graph) ToJSON() (string, error) {
	type JSONInputData struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Required bool   `json:"required"`
	}

	type JSONNode struct {
		ID                 string          `json:"id"`
		Type               string          `json:"type"`
		IsStartNode        bool            `json:"isStartNode,omitempty"`
		IsFinalNode        bool            `json:"isFinalNode,omitempty"`
		NextNodeIDList     []string        `json:"nextNodeIds"`
		PreviousNodeIDList []string        `json:"previousNodeIds"`
		InputData          []JSONInputData `json:"inputData,omitempty"`
		Executor           string          `json:"executor,omitempty"`
	}

	type JSONGraph struct {
		ID          string              `json:"id"`
		Nodes       map[string]JSONNode `json:"nodes"`
		Edges       map[string][]string `json:"edges"`
		StartNodeID string              `json:"startNodeId"`
	}

	jsonGraph := JSONGraph{
		ID:          g.id,
		Nodes:       make(map[string]JSONNode),
		Edges:       g.edges,
		StartNodeID: g.startNodeID,
	}

	// Convert nodes to JSONNode
	for id, node := range g.nodes {
		jsonNode := JSONNode{
			ID:                 id,
			Type:               string(node.GetType()),
			IsStartNode:        node.IsStartNode(),
			IsFinalNode:        node.IsFinalNode(),
			NextNodeIDList:     node.GetNextNodeList(),
			PreviousNodeIDList: node.GetPreviousNodeList(),
		}

		// Set executor if available
		if executor := node.GetExecutor(); executor != nil {
			jsonNode.Executor = executor.GetName()
		}

		// Convert and set input data
		inputData := node.GetInputData()
		if len(inputData) > 0 {
			jsonNode.InputData = make([]JSONInputData, len(inputData))
			for i, input := range inputData {
				jsonNode.InputData[i] = JSONInputData(input)
			}
		}

		jsonGraph.Nodes[id] = jsonNode
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(jsonGraph)
	if err != nil {
		return "", errors.New("failed to marshal graph to JSON: " + err.Error())
	}

	return string(jsonBytes), nil
}
