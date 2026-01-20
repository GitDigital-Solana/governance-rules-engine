governance-rules-engine/evaluator.go

```go
package main

import (
	"github.com/PaesslerAG/jsonpath"
)

// JSONPathEvaluator evaluates JSONPath expressions
type JSONPathEvaluator struct{}

func NewJSONPathEvaluator() *JSONPathEvaluator {
	return &JSONPathEvaluator{}
}

func (e *JSONPathEvaluator) Evaluate(expression string, data map[string]interface{}) (bool, error) {
	compiled, err := jsonpath.Compile(expression)
	if err != nil {
		return false, err
	}

	result, err := compiled.Execute(data)
	if err != nil {
		return false, err
	}

	// Convert result to boolean
	switch v := result.(type) {
	case bool:
		return v, nil
	case nil:
		return false, nil
	default:
		// Non-nil, non-bool values are considered truthy
		return true, nil
	}
}

// CompositeEvaluator evaluates multiple conditions
type CompositeEvaluator struct {
	evaluators []Evaluator
}

type Evaluator interface {
	Evaluate(expression string, data map[string]interface{}) (bool, error)
}

func NewCompositeEvaluator() *CompositeEvaluator {
	return &CompositeEvaluator{
		evaluators: []Evaluator{
			NewJSONPathEvaluator(),
			NewRegExEvaluator(),
		},
	}
}

func (c *CompositeEvaluator) Evaluate(expression string, data map[string]interface{}) (bool, error) {
	for _, evaluator := range c.evaluators {
		if result, err := evaluator.Evaluate(expression, data); err == nil {
			return result, nil
		}
	}
	return false, fmt.Errorf("no evaluator could process expression")
}
