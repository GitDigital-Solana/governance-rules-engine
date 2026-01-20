### **governance-rules-engine/main.go**
```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Rule represents a governance rule
type Rule struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Condition string    `json:"condition"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	CreatedAt time.Time `json:"createdAt"`
}

// Policy represents a governance policy
type Policy struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Rules       []Rule    `json:"rules"`
	TargetType  string    `json:"targetType"`
	Enabled     bool      `json:"enabled"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Resource represents a resource to evaluate
type Resource struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Metadata   map[string]string      `json:"metadata"`
	Tags       map[string]string      `json:"tags"`
}

// EvaluationResult represents evaluation outcome
type EvaluationResult struct {
	ResourceID  string      `json:"resourceId"`
	EvaluatedAt time.Time   `json:"evaluatedAt"`
	Passed      bool        `json:"passed"`
	Violations  []Violation `json:"violations,omitempty"`
	Metrics     Metrics     `json:"metrics"`
}

type Violation struct {
	RuleID      string `json:"ruleId"`
	RuleName    string `json:"ruleName"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
	Condition   string `json:"condition"`
	ResourceID  string `json:"resourceId"`
	ResourceType string `json:"resourceType"`
}

type Metrics struct {
	TotalRules     int   `json:"totalRules"`
	EvaluatedRules int   `json:"evaluatedRules"`
	ExecutionTime  int64 `json:"executionTime"`
	MemoryUsage    int64 `json:"memoryUsage"`
}

// RulesEngine is the main engine
type RulesEngine struct {
	policies map[string]Policy
	cache    map[string]interface{}
}

func NewRulesEngine() *RulesEngine {
	return &RulesEngine{
		policies: make(map[string]Policy),
		cache:    make(map[string]interface{}),
	}
}

func (e *RulesEngine) RegisterPolicy(policy Policy) error {
	e.policies[policy.ID] = policy
	log.Printf("Registered policy: %s (v%s)", policy.Name, policy.Version)
	return nil
}

func (e *RulesEngine) Evaluate(resource Resource) EvaluationResult {
	start := time.Now()
	result := EvaluationResult{
		ResourceID:  resource.ID,
		EvaluatedAt: time.Now(),
		Passed:      true,
		Violations:  []Violation{},
		Metrics: Metrics{
			TotalRules: 0,
		},
	}

	for _, policy := range e.policies {
		if !policy.Enabled || policy.TargetType != resource.Type {
			continue
		}

		result.Metrics.TotalRules += len(policy.Rules)

		for _, rule := range policy.Rules {
			result.Metrics.EvaluatedRules++

			passed, err := e.evaluateRule(rule, resource)
			if err != nil {
				log.Printf("Error evaluating rule %s: %v", rule.Name, err)
				continue
			}

			if !passed {
				result.Passed = false
				result.Violations = append(result.Violations, Violation{
					RuleID:       rule.ID,
					RuleName:     rule.Name,
					Message:      rule.Message,
					Severity:     rule.Severity,
					Condition:    rule.Condition,
					ResourceID:   resource.ID,
					ResourceType: resource.Type,
				})
			}
		}
	}

	result.Metrics.ExecutionTime = time.Since(start).Milliseconds()
	return result
}

func (e *RulesEngine) evaluateRule(rule Rule, resource Resource) (bool, error) {
	// Use JSONPath evaluator
	evaluator := NewJSONPathEvaluator()
	return evaluator.Evaluate(rule.Condition, resource.Properties)
}

func (e *RulesEngine) EvaluateBatch(resources []Resource) []EvaluationResult {
	results := make([]EvaluationResult, len(resources))
	
	for i, resource := range resources {
		results[i] = e.Evaluate(resource)
	}
	
	return results
}

func main() {
	engine := NewRulesEngine()
	
	// Example usage
	policy := Policy{
		ID:          "s3-encryption",
		Name:        "S3 Encryption Required",
		Description: "Requires encryption on all S3 buckets",
		TargetType:  "aws_s3_bucket",
		Enabled:     true,
		Version:     "1.0.0",
		Rules: []Rule{
			{
				ID:        "rule-001",
				Name:      "require-encryption",
				Condition: "$.encryption != null",
				Message:   "S3 bucket must have encryption enabled",
				Severity:  "high",
			},
		},
	}
	
	engine.RegisterPolicy(policy)
	
	resource := Resource{
		ID:   "bucket-001",
		Type: "aws_s3_bucket",
		Properties: map[string]interface{}{
			"encryption": "AES256",
			"versioning": true,
		},
	}
	
	result := engine.Evaluate(resource)
	fmt.Printf("Evaluation Result: %+v\n", result)
}
