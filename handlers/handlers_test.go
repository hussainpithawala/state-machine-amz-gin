package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hussainpithawala/state-machine-amz-gin/models"
	"github.com/stretchr/testify/assert"
)

// Helper functions for tests
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createRequest(method, path string, body interface{}) *http.Request {
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	return req
}

// ==================== State Machine Handler Tests ====================

func TestCreateStateMachine_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.POST("/state-machines", CreateStateMachine)

	reqBody := map[string]interface{}{
		"id":   "test-sm",
		"name": "Test",
	}
	req := createRequest("POST", "/state-machines", reqBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetStateMachine_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/state-machines/:stateMachineId", GetStateMachine)

	req := httptest.NewRequest("GET", "/state-machines/test-sm", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListStateMachines_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/state-machines", ListStateMachines)

	req := httptest.NewRequest("GET", "/state-machines", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== Execution Handler Tests ====================

func TestStartExecution_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.POST("/state-machines/:stateMachineId/executions", StartExecution)

	reqBody := map[string]interface{}{
		"name":  "test-execution",
		"input": map[string]interface{}{"key": "value"},
	}
	req := createRequest("POST", "/state-machines/test-sm/executions", reqBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetExecution_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/executions/:executionId", GetExecution)

	req := httptest.NewRequest("GET", "/executions/test-exec", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListExecutions_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/state-machines/:stateMachineId/executions", ListExecutions)

	req := httptest.NewRequest("GET", "/state-machines/test-sm/executions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCountExecutions_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/state-machines/:stateMachineId/executions/count", CountExecutions)

	req := httptest.NewRequest("GET", "/state-machines/test-sm/executions/count?status=RUNNING", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStopExecution_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.DELETE("/executions/:executionId", StopExecution)

	req := httptest.NewRequest("DELETE", "/executions/test-exec", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetExecutionHistory_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/executions/:executionId/history", GetExecutionHistory)

	req := httptest.NewRequest("GET", "/executions/test-exec/history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== Batch Handler Tests ====================

func TestEnqueueExecution_NoQueue(t *testing.T) {
	router := setupTestRouter()
	router.POST("/queue/enqueue", EnqueueExecution)

	reqBody := map[string]interface{}{
		"stateMachineId": "test-sm",
		"executionName":  "test-exec",
	}
	req := createRequest("POST", "/queue/enqueue", reqBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetBatchStatus_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/batch/:batchId/status", GetBatchStatus)

	req := httptest.NewRequest("GET", "/batch/batch-123/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPauseBatchExecution_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.POST("/batch/:batchId/pause", PauseBatchExecution)

	req := httptest.NewRequest("POST", "/batch/batch-123/pause", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestResumeBatchExecution_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.POST("/batch/:batchId/resume", ResumeBatchExecution)

	req := httptest.NewRequest("POST", "/batch/batch-123/resume", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCancelBatchExecution_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.DELETE("/batch/:batchId", CancelBatchExecution)

	req := httptest.NewRequest("DELETE", "/batch/batch-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListBatches_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/batch", ListBatches)

	req := httptest.NewRequest("GET", "/batch", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== Bulk Handler Tests ====================

func TestGetBulkStatus_NoOrchestrator(t *testing.T) {
	router := setupTestRouter()
	router.GET("/bulk/:orchestratorId/status", GetBulkStatus)

	req := httptest.NewRequest("GET", "/bulk/bulk-123/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPauseBulkExecution_NoOrchestrator(t *testing.T) {
	router := setupTestRouter()
	router.POST("/bulk/:orchestratorId/pause", PauseBulkExecution)

	req := httptest.NewRequest("POST", "/bulk/bulk-123/pause", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestResumeBulkExecution_NoOrchestrator(t *testing.T) {
	router := setupTestRouter()
	router.POST("/bulk/:orchestratorId/resume", ResumeBulkExecution)

	req := httptest.NewRequest("POST", "/bulk/bulk-123/resume", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCancelBulkExecution_NoOrchestrator(t *testing.T) {
	router := setupTestRouter()
	router.DELETE("/bulk/:orchestratorId", CancelBulkExecution)

	req := httptest.NewRequest("DELETE", "/bulk/bulk-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListBulkExecutions_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/bulk", ListBulkExecutions)

	req := httptest.NewRequest("GET", "/bulk", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== Message Handler Tests ====================

func TestResumeExecution_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.POST("/executions/:executionId/resume", ResumeExecution)

	reqBody := map[string]interface{}{
		"output": map[string]interface{}{"result": "approved"},
	}
	req := createRequest("POST", "/executions/test-exec/resume", reqBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestResumeByCorrelation_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.POST("/state-machines/:stateMachineId/resume-by-correlation", ResumeByCorrelation)

	reqBody := map[string]interface{}{
		"correlationKey":   "orderId",
		"correlationValue": "12345",
	}
	req := createRequest("POST", "/state-machines/test-sm/resume-by-correlation", reqBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFindWaitingExecutions_MissingRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/state-machines/:stateMachineId/waiting", FindWaitingExecutions)

	req := httptest.NewRequest("GET", "/state-machines/test-sm/waiting?correlationKey=orderId&correlationValue=12345", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== Monitoring & OpenAPI Handler Tests ====================

func TestHealthCheck_NoRepo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/health", HealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", resp["status"])
	services, ok := resp["services"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "not_configured", services["database"])
}

func TestGetOpenAPISpec_Success(t *testing.T) {
	router := setupTestRouter()
	router.GET("/openapi.json", GetOpenAPISpec)

	req := httptest.NewRequest("GET", "/openapi.json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 if openapi.json exists in the project
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListTransformers_NoRegistry(t *testing.T) {
	router := setupTestRouter()
	router.GET("/transformers", ListTransformers)

	req := httptest.NewRequest("GET", "/transformers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ListTransformersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Empty(t, response.Transformers)
}
