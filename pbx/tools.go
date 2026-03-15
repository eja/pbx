// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package pbx

import (
	"encoding/json"
	"fmt"
)

var Tools = map[string]LLMTool{
	"calculate": {
		Description: "Perform basic mathematical operations (add, subtract, multiply, divide) on two numbers",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"operation": map[string]any{
					"type":        "string",
					"description": "The operation to perform: 'add', 'subtract', 'multiply', or 'divide'",
					"enum":        []string{"add", "subtract", "multiply", "divide"},
				},
				"a": map[string]any{
					"type":        "number",
					"description": "The first number",
				},
				"b": map[string]any{
					"type":        "number",
					"description": "The second number",
				},
			},
			"required": []string{"operation", "a", "b"},
		},
		Callback: func(args string) (string, error) {
			type MathInput struct {
				Operation string  `json:"operation"`
				A         float64 `json:"a"`
				B         float64 `json:"b"`
			}

			var input MathInput
			if err := json.Unmarshal([]byte(args), &input); err != nil {
				return "", err
			}

			var result float64
			switch input.Operation {
			case "add":
				result = input.A + input.B
			case "subtract":
				result = input.A - input.B
			case "multiply":
				result = input.A * input.B
			case "divide":
				if input.B == 0 {
					return "Error: Cannot divide by zero", nil
				}
				result = input.A / input.B
			default:
				return "Error: Invalid operation", nil
			}

			return fmt.Sprintf("Result: %.2f", result), nil
		},
	},
}
