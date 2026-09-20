package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/system"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji siklus lengkap CRUD kebijakan Unified Policy Engine dan penolakan aturan berformat JSON tidak valid.
func TestPolicyCRUDAndValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	rdb, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer rdb.Close()

	repo := system.NewPolicyRepository(db)
	service := system.NewPolicyService(repo, rdb)
	ctrl := system.NewPolicyController(service)

	r := gin.New()
	r.POST("/policies", ctrl.Create)
	r.GET("/policies/:id", ctrl.GetByID)
	r.GET("/policies", ctrl.List)
	r.PUT("/policies/:id", ctrl.Update)
	r.DELETE("/policies/:id", ctrl.Delete)

	ctx := context.Background()

	// 1. Uji penolakan format JSON tidak valid
	invalidPayload := map[string]interface{}{
		"policy_domain":   "ATTENDANCE",
		"name":            "Kebijakan Malformed",
		"effective_date":  "2026-09-20",
		"condition_rules": "{malformed_json",
	}
	invalidBody, _ := json.Marshal(invalidPayload)
	reqInv, _ := http.NewRequest(http.MethodPost, "/policies", bytes.NewBuffer(invalidBody))
	reqInv.Header.Set("Content-Type", "application/json")
	wInv := httptest.NewRecorder()
	r.ServeHTTP(wInv, reqInv)
	assert.Equal(t, http.StatusBadRequest, wInv.Code)

	// 2. Uji pembuatan kebijakan valid
	validPayload := map[string]interface{}{
		"policy_domain":  "ATTENDANCE",
		"name":           "Kebijakan Uji Toleransi",
		"effective_date": "2026-09-20",
		"condition_rules": map[string]interface{}{
			"grace_period_minutes": 15,
			"late_tier_1_max":      20,
		},
		"action_definitions": map[string]interface{}{
			"deduct_xp": 1,
		},
		"priority_order": 2,
	}
	body, _ := json.Marshal(validPayload)
	req, _ := http.NewRequest(http.MethodPost, "/policies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp struct {
		Data system.Policy `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)
	policyID := createResp.Data.ID
	assert.NotEqual(t, uuid.Nil, policyID)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM policies WHERE id = $1", policyID)
		_ = rdb.Client.Del(ctx, "policy:domain:ATTENDANCE").Err()
	})

	// 3. Uji pengambilan detail kebijakan berdasarkan ID
	reqGet, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/policies/%s", policyID.String()), nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusOK, wGet.Code)

	// 4. Uji pembaruan data kebijakan
	updatePayload := map[string]interface{}{
		"name": "Kebijakan Toleransi Diperbarui",
	}
	updateBody, _ := json.Marshal(updatePayload)
	reqUpdate, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/policies/%s", policyID.String()), bytes.NewBuffer(updateBody))
	reqUpdate.Header.Set("Content-Type", "application/json")
	wUpdate := httptest.NewRecorder()
	r.ServeHTTP(wUpdate, reqUpdate)
	assert.Equal(t, http.StatusOK, wUpdate.Code)

	// 5. Uji penghapusan data kebijakan
	reqDel, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/policies/%s", policyID.String()), nil)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)
	assert.Equal(t, http.StatusOK, wDel.Code)
}

// Menguji evaluasi aturan kebijakan bisnis serta mekanisme caching dan invalidasi pada Redis.
func TestPolicyEvaluationAndRedisCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	rdb, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer rdb.Close()

	repo := system.NewPolicyRepository(db)
	service := system.NewPolicyService(repo, rdb)

	ctx := context.Background()

	// Buat kebijakan aktif untuk domain XP
	policyID := uuid.New()
	p := &system.Policy{
		ID:                policyID,
		PolicyDomain:      "XP",
		Name:              "Kebijakan Sanksi Penalti XP",
		EffectiveDate:     time.Now().UTC(),
		ConditionRules:    []byte(`{"penalty_rate": 2, "max_late": 30}`),
		ActionDefinitions: []byte(`{"notify_supervisor": true}`),
		PriorityOrder:     1,
		IsActive:          true,
	}
	err = repo.Create(ctx, p)
	require.NoError(t, err)

	cacheKey := "policy:domain:XP"
	_ = rdb.Client.Del(ctx, cacheKey).Err()

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM policies WHERE id = $1", policyID)
		_ = rdb.Client.Del(ctx, cacheKey).Err()
	})

	// 1. Evaluasi pertama (Cache Miss -> DB Query -> Set Cache)
	res1, err := service.EvaluatePolicy(ctx, "XP", map[string]interface{}{"late_minutes": 10})
	require.NoError(t, err)
	assert.True(t, res1.Allowed)
	assert.Equal(t, policyID, res1.PolicyID)

	// Verifikasi bahwa cache Redis telah terisi
	cachedVal, err := rdb.Client.Get(ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.NotEmpty(t, cachedVal)

	// 2. Evaluasi kedua (Cache Hit dari Redis)
	res2, err := service.EvaluatePolicy(ctx, "XP", map[string]interface{}{"late_minutes": 10})
	require.NoError(t, err)
	assert.True(t, res2.Allowed)
	assert.Equal(t, policyID, res2.PolicyID)

	// 3. Verifikasi invalidasi cache saat kebijakan dihapus
	err = service.DeletePolicy(ctx, policyID)
	require.NoError(t, err)

	exists, err := rdb.Client.Exists(ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}
