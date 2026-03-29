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
	Name                   string      `json:"name" binding:"required"`
	Input                  interface{} `json:"input"`
	SourceExecutionID      string      `json:"sourceExecutionId,omitempty"`      // ID of execution whose output will be used as input
	SourceStateName        string      `json:"sourceStateName,omitempty"`        // Optional: specific state's output to use from source execution
	SourceInputTransformer string      `json:"sourceInputTransformer,omitempty"` // Optional: JSONPath or transformation expression to apply
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
	Filter            *BatchExecutionFilterRequest `json:"filter"`
	NamePrefix        string                       `json:"namePrefix"`
	Concurrency       int                          `json:"concurrency"`
	Mode              string                       `json:"mode"` // "distributed", "concurrent", "sequential"
	StopOnError       bool                         `json:"stopOnError"`
	ExecutionNameList []string                     `json:"executionNameList"` // Explicit list of execution names
	DoMicroBatch      bool                         `json:"doMicroBatch"`
	MicroBatchSize    int                          `json:"microBatchSize"`
}

// BatchExecutionFilterRequest represents filter parameters for listing executions
type BatchExecutionFilterRequest struct {
	SourceStateMachineId   string   `json:"sourceStateMachineId"`
	CurrentState           string   `json:"currentState"`
	SourceStateName        string   `json:"sourceStateName,omitempty"`        // Optional: specific state's output to use from source execution
	SourceInputTransformer string   `json:"sourceInputTransformer,omitempty"` // Optional: JSONPath or transformation expression to apply
	ApplyUnique            bool     `json:"applyUnique,omitempty"`
	Status                 string   `json:"status,omitempty"`
	StartTimeFrom          int64    `json:"startTimeFrom"`
	StartTimeTo            int64    `json:"startTimeTo"`
	NamePattern            string   `json:"namePattern"`
	Limit                  int      `json:"limit"`
	Offset                 int      `json:"offset"`
	States                 []string `json:"states"`
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

// ExecuteBulkRequest represents a request to execute a bulk operation with orchestration
type ExecuteBulkRequest struct {
	NamePrefix     string        `json:"namePrefix"`
	Concurrency    int           `json:"concurrency"`
	Mode           string        `json:"mode"` // "distributed", "concurrent", "sequential"
	StopOnError    bool          `json:"stopOnError"`
	Inputs         []interface{} `json:"inputs"` // Raw JSON array of inputs
	DoMicroBatch   bool          `json:"doMicroBatch"`
	MicroBatchSize int           `json:"microBatchSize"`
	OrchestratorID string        `json:"orchestratorId"` // Optional: custom orchestrator ID
	PauseThreshold float64       `json:"pauseThreshold"` // Optional: failure rate threshold for auto-pause (0.0-1.0)
	ResumeStrategy string        `json:"resumeStrategy"` // Optional: "manual", "automatic", "timeout"
	TimeoutSeconds int           `json:"timeoutSeconds"` // Optional: timeout for automatic resume
}

// ResumeOrchestratorRequest represents a request to resume a stuck orchestrator
// Used to push orchestrators stuck at WaitForMicroBatchCompletion/WaitForBulkMicroBatchCompletion
// to handle exceptional conditions
type ResumeOrchestratorRequest struct {
	BatchID          string `json:"batchId" binding:"required"`
	MicroBatchID     string `json:"microBatchId" binding:"required"`
	OrchestratorSMID string `json:"orchestratorSmId" binding:"required"` // "orchestrator" or "bulk-orchestrator"
}

// SignalResumeRequest represents a request to signal a paused batch to resume
// Used for human-in-the-loop pause/resume cycle via Redis signaling
type SignalResumeRequest struct {
	BatchID  string `json:"batchId" binding:"required"`
	Operator string `json:"operator" binding:"required"` // Operator name/ID
	Notes    string `json:"notes"`                       // Optional notes about the resume
}

// RevokeResumeRequest represents a request to revoke an unconsumed resume signal
type RevokeResumeRequest struct {
	BatchID string `json:"batchId" binding:"required"`
}

// CheckResumeRequest represents a request to check for a resume signal
type CheckResumeRequest struct {
	BatchID string `json:"batchId" binding:"required"`
}
