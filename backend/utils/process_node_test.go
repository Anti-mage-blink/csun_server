package utils

import (
	"testing"

	"github.com/expr-lang/expr"
)

func TestExprEvaluation(t *testing.T) {
	tests := []struct {
		name       string
		exprStr    string
		env        map[string]interface{}
		wantResult bool
	}{
		{
			name:       "Equal string match",
			exprStr:    `main_category == "汽车涂层制动盘"`,
			env:        map[string]interface{}{"main_category": "汽车涂层制动盘"},
			wantResult: true,
		},
		{
			name:       "Not equal string match",
			exprStr:    `main_category != "汽车涂层制动盘"`,
			env:        map[string]interface{}{"main_category": "普通制动盘"},
			wantResult: true,
		},
		{
			name:       "Float comparison match",
			exprStr:    `quote_float_rate > 0.1`,
			env:        map[string]interface{}{"quote_float_rate": 0.15},
			wantResult: true,
		},
		{
			name:       "Float comparison not match",
			exprStr:    `quote_float_rate > 0.1`,
			env:        map[string]interface{}{"quote_float_rate": 0.05},
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.exprStr, expr.Env(tt.env))
			if err != nil {
				t.Fatalf("Compile error: %v", err)
			}

			output, err := expr.Run(program, tt.env)
			if err != nil {
				t.Fatalf("Run error: %v", err)
			}

			boolVal, ok := output.(bool)
			if !ok {
				t.Fatalf("Output is not bool: %v", output)
			}

			if boolVal != tt.wantResult {
				t.Errorf("Expr %s with env %v got %v, want %v", tt.exprStr, tt.env, boolVal, tt.wantResult)
			}
		})
	}
}
