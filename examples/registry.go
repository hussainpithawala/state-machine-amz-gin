package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hussainpithawala/state-machine-amz-gin/middleware"
	"github.com/hussainpithawala/state-machine-amz-go/pkg/executor"
)

func RegisterGlobalFunctions(baseExecutor *executor.BaseExecutor) *executor.StateRegistry {
	// 4. Create executor and register handlers
	baseExecutor.RegisterGoFunction("initial-task", func(ctx context.Context, input interface{}) (interface{}, error) {
		fmt.Println("  → Executing initial task...")
		inputMap := input.(map[string]interface{})
		return map[string]interface{}{
			"orderId": inputMap["orderId"],
			"status":  "INITIAL_DONE",
		}, nil
	})

	baseExecutor.RegisterGoFunction("final-task", func(ctx context.Context, input interface{}) (interface{}, error) {
		fmt.Println("  → Executing final task...")
		return map[string]interface{}{
			"status": "COMPLETED",
		}, nil
	})

	baseExecutor.RegisterGoFunction("ingest:data", func(ctx context.Context, input interface{}) (interface{}, error) {
		data := input.(map[string]interface{})
		fmt.Printf("\n[Ingest] Processing: %v\n", data["orderId"])

		return map[string]interface{}{
			"orderId":     data["orderId"],
			"rawData":     data,
			"ingestedAt":  time.Now().Format(time.RFC3339),
			"ingestionID": fmt.Sprintf("ing-%v", data["orderId"]),
		}, nil
	})

	baseExecutor.RegisterGoFunction("process:order", func(ctx context.Context, input interface{}) (interface{}, error) {
		data := input.(map[string]interface{})
		orderId := data["orderId"]
		fmt.Printf("\n[Process] Processing order: %v\n", orderId)

		// Simulate processing
		time.Sleep(100 * time.Millisecond)

		return map[string]interface{}{
			"orderId":        orderId,
			"originalData":   data,
			"processedData":  fmt.Sprintf("Processed-%v", orderId),
			"processingTime": time.Now().Format(time.RFC3339),
			"status":         "processed",
		}, nil
	})

	baseExecutor.RegisterGoFunction("validate:order", func(ctx context.Context, input interface{}) (interface{}, error) {
		data := input.(map[string]interface{})
		orderId := data["orderId"]
		fmt.Printf("[Validate] Validating order: %v\n", orderId)

		return map[string]interface{}{
			"orderId":      orderId,
			"validated":    true,
			"validatedAt":  time.Now().Format(time.RFC3339),
			"originalData": data,
		}, nil
	})

	return nil
}

func RegisterTransformerFunctions() *middleware.TransformerRegistry {
	return &middleware.TransformerRegistry{
		"csv2Json": func(output interface{}) (interface{}, error) {
			fmt.Println("[Transformer] Transforming input from Execution A...")
			data := output.(map[string]interface{})

			// Extract only specific fields and add metadata
			transformed := map[string]interface{}{
				"validatedData": data["validationResult"],
				"source":        "execution-A-001",
				"transformedAt": "2024-01-01T12:05:00Z",
			}

			fmt.Printf("[Transformer] Transformed: %v\n", transformed)
			return transformed, nil
		},
	}
}
