package main

import (
	"context"
	"fmt"

	"github.com/hussainpithawala/state-machine-amz-go/pkg/executor"
)

func registerGlobalFunctions(baseExecutor *executor.BaseExecutor) *executor.StateRegistry {
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
	return nil
}
