package models

// CreateStateMachineRequest represents a request to create a new state machine
type CreateStateMachineRequest struct {
	ID          string                 `json:"id" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Definition  interface{}            `json:"definition" binding:"required"`
	Type        string                 `json:"type"`
	Version     string                 `json:"version"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UpdateStateMachineRequest represents a request to update a state machine
type UpdateStateMachineRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Definition  interface{}            `json:"definition"`
	Version     string                 `json:"version"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// StartExecutionRequest represents a request to start an execution
type StartExecutionRequest struct {
	Name  string      `json:"name" binding:"required"`
	Input interface{} `json:"input"`
}

// ResumeExecutionRequest represents a request to resume a paused execution
type ResumeExecutionRequest struct {
	Output interface{} `json:"output"`
}

// ResumeByCorrelationRequest represents a request to resume executions by correlation
type ResumeByCorrelationRequest struct {
	CorrelationKey   string      `json:"correlationKey" binding:"required"`
	CorrelationValue interface{} `json:"correlationValue" binding:"required"`
	Output           interface{} `json:"output"`
}

// ExecuteBatchRequest represents a request to execute a batch of executions
type ExecuteBatchRequest struct {
	Filter            *ExecutionFilterRequest `json:"filter"`
	NamePrefix        string                  `json:"namePrefix"`
	Concurrency       int                     `json:"concurrency"`
	Mode              string                  `json:"mode"` // "distributed", "concurrent", "sequential"
	StopOnError       bool                    `json:"stopOnError"`
	ExecutionNameList []string                `json:"executionNameList"` // Explicit list of execution names
}

// ExecutionFilterRequest represents filter parameters for listing executions
type ExecutionFilterRequest struct {
	StateMachineID string   `json:"stateMachineId"`
	Status         string   `json:"status"`
	StartTimeFrom  int64    `json:"startTimeFrom"`
	StartTimeTo    int64    `json:"startTimeTo"`
	NamePattern    string   `json:"namePattern"`
	Limit          int      `json:"limit"`
	Offset         int      `json:"offset"`
	States         []string `json:"states"`
}

// EnqueueExecutionRequest represents a request to enqueue an execution task
type EnqueueExecutionRequest struct {
	StateMachineID    string      `json:"stateMachineId" binding:"required"`
	ExecutionName     string      `json:"executionName" binding:"required"`
	Input             interface{} `json:"input"`
	Queue             string      `json:"queue"`
	SourceExecutionID string      `json:"sourceExecutionId"`
	SourceStateName   string      `json:"sourceStateName"`
}
