package models

import "time"

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// StartExecutionResponse represents the response for starting an execution
type StartExecutionResponse struct {
	ExecutionID    string      `json:"executionId"`
	StateMachineID string      `json:"stateMachineId"`
	Name           string      `json:"name"`
	Status         string      `json:"status"`
	StartTime      time.Time   `json:"startTime"`
	Input          interface{} `json:"input,omitempty"`
}

// ExecutionResponse represents a full execution response
type ExecutionResponse struct {
	ExecutionID    string                 `json:"executionId"`
	StateMachineID string                 `json:"stateMachineId"`
	Name           string                 `json:"name"`
	Status         string                 `json:"status"`
	CurrentState   string                 `json:"currentState"`
	Input          interface{}            `json:"input,omitempty"`
	Output         interface{}            `json:"output,omitempty"`
	StartTime      *time.Time             `json:"startTime"`
	EndTime        *time.Time             `json:"endTime,omitempty"`
	Error          string                 `json:"error,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// StateHistoryResponse represents a state history item
type StateHistoryResponse struct {
	ID             string                 `json:"id"`
	ExecutionID    string                 `json:"executionId"`
	StateName      string                 `json:"stateName"`
	StateType      string                 `json:"stateType"`
	Status         string                 `json:"status"`
	Input          interface{}            `json:"input,omitempty"`
	Output         interface{}            `json:"output,omitempty"`
	StartTime      *time.Time             `json:"startTime"`
	EndTime        *time.Time             `json:"endTime,omitempty"`
	Error          string                 `json:"error,omitempty"`
	RetryCount     int                    `json:"retryCount"`
	SequenceNumber int                    `json:"sequenceNumber"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ListExecutionsResponse represents a paginated list of executions
type ListExecutionsResponse struct {
	Executions []*ExecutionResponse `json:"executions"`
	Total      int64                `json:"total"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
}

// BatchExecutionResponse represents the response for batch execution
type BatchExecutionResponse struct {
	BatchID       string `json:"batchId"`
	TotalEnqueued int    `json:"totalEnqueued"`
	TotalFailed   int    `json:"totalFailed"`
	Mode          string `json:"mode"`
}

// ResumeByCorrelationResponse represents the response for resume by correlation
type ResumeByCorrelationResponse struct {
	ResumedCount int      `json:"resumedCount"`
	ExecutionIDs []string `json:"executionIds"`
}

// EnqueueExecutionResponse represents the response for enqueuing an execution
type EnqueueExecutionResponse struct {
	TaskID     string    `json:"taskId"`
	Queue      string    `json:"queue"`
	EnqueuedAt time.Time `json:"enqueuedAt"`
}

// QueueStatsResponse represents queue statistics
type QueueStatsResponse struct {
	Queues map[string]QueueStats `json:"queues"`
}

// QueueStats represents statistics for a single queue
type QueueStats struct {
	Pending int `json:"pending"`
	Active  int `json:"active"`
	Failed  int `json:"failed"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// StateMachineResponse represents a state machine definition response
type StateMachineResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Definition  interface{}            `json:"definition"`
	Type        string                 `json:"type,omitempty"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ListStateMachinesResponse represents a list of state machines
type ListStateMachinesResponse struct {
	StateMachines []*StateMachineResponse `json:"stateMachines"`
	Total         int                     `json:"total"`
}
