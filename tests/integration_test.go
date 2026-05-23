package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/horlakz/jarakey.assessment/internal/app"
	"github.com/horlakz/jarakey.assessment/internal/config"
	"github.com/horlakz/jarakey.assessment/internal/database"
	"github.com/horlakz/jarakey.assessment/internal/entities"
	"gorm.io/gorm"
)

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

func TestAuthorizationPermissionDriftAndOverrides(t *testing.T) {
	application, seed := newTestApp(t)
	t.Cleanup(func() {
		_ = application.Fiber.Shutdown()
	})

	token := login(t, application, application.Config.DefaultEmail, application.Config.DefaultPassword)

	resp := performJSONRequest(t, application, http.MethodPost, "/gate/open", nil, token, seed.EstateID)
	assertStatus(t, resp, http.StatusOK)

	resp = performJSONRequest(t, application, http.MethodPost, "/debug/downgrade-role", nil, token, seed.EstateID)
	assertStatus(t, resp, http.StatusOK)

	resp = performJSONRequest(t, application, http.MethodPost, "/gate/open", nil, token, seed.EstateID)
	assertStatus(t, resp, http.StatusForbidden)

	if err := application.DB.Save(&entities.UserPermissionOverride{
		UserID:         seed.UserID,
		EstateID:       seed.EstateID,
		PermissionCode: database.GateOpenCode,
		Effect:         entities.OverrideAllow,
	}).Error; err != nil {
		t.Fatalf("create allow override: %v", err)
	}

	resp = performJSONRequest(t, application, http.MethodPost, "/gate/open", nil, token, seed.EstateID)
	assertStatus(t, resp, http.StatusOK)

	if err := application.DB.Save(&entities.UserPermissionOverride{
		UserID:         seed.UserID,
		EstateID:       seed.EstateID,
		PermissionCode: database.GateOpenCode,
		Effect:         entities.OverrideDeny,
	}).Error; err != nil {
		t.Fatalf("create deny override: %v", err)
	}

	resp = performJSONRequest(t, application, http.MethodPost, "/gate/open", nil, token, seed.EstateID)
	assertStatus(t, resp, http.StatusForbidden)
}

func TestContextAndMembershipFailures(t *testing.T) {
	application, seed := newTestApp(t)
	t.Cleanup(func() {
		_ = application.Fiber.Shutdown()
	})

	token := login(t, application, application.Config.DefaultEmail, application.Config.DefaultPassword)

	testCases := []struct {
		name       string
		estateID   string
		setup      func(t *testing.T, db *gorm.DB, seed *database.SeedResult) string
		wantStatus int
	}{
		{
			name:       "missing estate header",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong estate id without membership",
			estateID:   "0197d0bc-3e51-7dbf-a394-1a4ab33f8888",
			wantStatus: http.StatusForbidden,
		},
		{
			name: "estate exists but user lacks membership",
			setup: func(t *testing.T, db *gorm.DB, seed *database.SeedResult) string {
				estate := entities.Estate{Name: "No Membership Estate"}
				if err := db.Create(&estate).Error; err != nil {
					t.Fatalf("create estate: %v", err)
				}
				return estate.ID
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			estateID := tc.estateID
			if tc.setup != nil {
				estateID = tc.setup(t, application.DB, seed)
			}
			resp := performJSONRequest(t, application, http.MethodPost, "/gate/open", nil, token, estateID)
			assertStatus(t, resp, tc.wantStatus)
		})
	}
}

func TestLoginAndMe(t *testing.T) {
	application, _ := newTestApp(t)
	t.Cleanup(func() {
		_ = application.Fiber.Shutdown()
	})

	token := login(t, application, application.Config.DefaultEmail, application.Config.DefaultPassword)
	resp := performJSONRequest(t, application, http.MethodGet, "/me", nil, token, "")
	assertStatus(t, resp, http.StatusOK)
}

func newTestApp(t *testing.T) (*app.Application, *database.SeedResult) {
	t.Helper()

	cfg := config.Config{
		AppEnv:          "test",
		ServerPort:      ":0",
		DatabaseDSN:     filepath.Join(t.TempDir(), "test.db"),
		JWTSecret:       "test-secret",
		DefaultEmail:    "admin@jarakey.com",
		DefaultPassword: "Pa$$w0rd!",
	}

	application, err := app.Build(cfg)
	if err != nil {
		t.Fatalf("build app: %v", err)
	}

	var user entities.User
	if err := application.DB.Where("email = ?", cfg.DefaultEmail).First(&user).Error; err != nil {
		t.Fatalf("load seeded user: %v", err)
	}

	var estate entities.Estate
	if err := application.DB.Where("name = ?", database.DefaultEstate).First(&estate).Error; err != nil {
		t.Fatalf("load seeded estate: %v", err)
	}

	return application, &database.SeedResult{
		UserID:   user.ID,
		EstateID: estate.ID,
	}
}

func login(t *testing.T, application *app.Application, email, password string) string {
	t.Helper()

	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	resp := performJSONRequest(t, application, http.MethodPost, "/auth/login", payload, "", "")
	assertStatus(t, resp, http.StatusOK)

	var body loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if body.AccessToken == "" {
		t.Fatal("expected access token")
	}
	return body.AccessToken
}

func performJSONRequest(t *testing.T, application *app.Application, method, path string, body interface{}, token, estateID string) *http.Response {
	t.Helper()

	var buffer bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buffer).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buffer)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if estateID != "" {
		req.Header.Set("X-Estate-ID", estateID)
	}

	resp, err := application.Fiber.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	return resp
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		t.Fatalf("expected status %d, got %d", expected, resp.StatusCode)
	}
}
