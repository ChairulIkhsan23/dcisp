package system

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Policy struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	PolicyDomain      string          `json:"policy_domain" db:"policy_domain"`
	Name              string          `json:"name" db:"name"`
	EffectiveDate     time.Time       `json:"effective_date" db:"effective_date"`
	ConditionRules    json.RawMessage `json:"condition_rules" db:"condition_rules"`
	ActionDefinitions json.RawMessage `json:"action_definitions" db:"action_definitions"`
	PriorityOrder     int             `json:"priority_order" db:"priority_order"`
	IsActive          bool            `json:"is_active" db:"is_active"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

type CreatePolicyRequest struct {
	PolicyDomain      string          `json:"policy_domain" binding:"required"`
	Name              string          `json:"name" binding:"required"`
	EffectiveDate     string          `json:"effective_date" binding:"required"` // Format: YYYY-MM-DD
	ConditionRules    json.RawMessage `json:"condition_rules" binding:"required"`
	ActionDefinitions json.RawMessage `json:"action_definitions"`
	PriorityOrder     int             `json:"priority_order"`
	IsActive          *bool           `json:"is_active"`
}

type UpdatePolicyRequest struct {
	Name              string          `json:"name"`
	EffectiveDate     string          `json:"effective_date"`
	ConditionRules    json.RawMessage `json:"condition_rules"`
	ActionDefinitions json.RawMessage `json:"action_definitions"`
	PriorityOrder     *int            `json:"priority_order"`
	IsActive          *bool           `json:"is_active"`
}

type PolicyEvaluationRequest struct {
	Domain  string                 `json:"domain" binding:"required"`
	Context map[string]interface{} `json:"context" binding:"required"`
}

type PolicyEvaluationResult struct {
	Allowed    bool                   `json:"allowed"`
	PolicyID   uuid.UUID              `json:"policy_id"`
	Domain     string                 `json:"domain"`
	Actions    map[string]interface{} `json:"actions"`
	Conditions map[string]interface{} `json:"conditions"`
	Reason     string                 `json:"reason,omitempty"`
}
