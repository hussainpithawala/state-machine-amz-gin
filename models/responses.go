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
	ExecutionID           string                 `json:"executionId"`
	StateMachineID        string                 `json:"stateMachineId"`
	Name                  string                 `json:"name"`
	Status                string                 `json:"status"`
	CurrentState          string                 `json:"currentState"`
	Input                 interface{}            `json:"input,omitempty"`
	Output                interface{}            `json:"output,omitempty"`
	StartTime             *time.Time             `json:"startTime"`
	EndTime               *time.Time             `json:"endTime,omitempty"`
	Error                 string                 `json:"error,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	HistorySequenceNumber int                    `json:"historySequenceNumber,omitempty"`
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

type ListTransformersResponse struct {
	Transformers []string `json:"transformers"`
}

// BulkExecutionResponse represents the response for bulk execution with orchestration
type BulkExecutionResponse struct {
	OrchestratorID string  `json:"orchestratorId"`
	BatchID        string  `json:"batchId"`
	Status         string  `json:"status"` // "Running", "Paused", "Completed", "Failed", "Cancelled"
	TotalEnqueued  int     `json:"totalEnqueued"`
	TotalFailed    int     `json:"totalFailed"`
	Mode           string  `json:"mode"`
	PausedAtBatch  int     `json:"pausedAtBatch,omitempty"`
	FailureRate    float64 `json:"failureRate,omitempty"`
}

// BulkStatusResponse represents the status of a bulk execution
type BulkStatusResponse struct {
	OrchestratorID string        `json:"orchestratorId"`
	Status         string        `json:"status"`
	Progress       *BulkProgress `json:"progress,omitempty"`
	Metrics        *BulkMetrics  `json:"metrics,omitempty"`
}

// BulkProgress represents the progress of a bulk execution
type BulkProgress struct {
	TotalBatches        int `json:"totalBatches"`
	CompletedBatches    int `json:"completedBatches"`
	CurrentBatch        int `json:"currentBatch"`
	TotalExecutions     int `json:"totalExecutions"`
	CompletedExecutions int `json:"completedExecutions"`
}

// BulkMetrics represents metrics for a bulk execution
type BulkMetrics struct {
	SuccessRate     float64 `json:"successRate"`
	FailureRate     float64 `json:"failureRate"`
	AverageDuration float64 `json:"averageDuration"`
	PausedAtBatch   int     `json:"pausedAtBatch,omitempty"`
	LastUpdated     int64   `json:"lastUpdated"`
}

// BulkActionResponse represents the response for bulk actions (pause/resume/cancel)
type BulkActionResponse struct {
	OrchestratorID string `json:"orchestratorId"`
	Action         string `json:"action"`
	Success        bool   `json:"success"`
	Message        string `json:"message,omitempty"`
}

// ResumeOrchestratorResponse represents the response for resuming a stuck orchestrator
type ResumeOrchestratorResponse struct {
	BatchID      string `json:"batchId"`
	MicroBatchID string `json:"microBatchId"`
	Status       string `json:"status"` // "resumed", "not_found", "error"
	Message      string `json:"message,omitempty"`
}
