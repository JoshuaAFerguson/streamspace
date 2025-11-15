package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupMFA(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Setup TOTP MFA",
			payload: map[string]interface{}{
				"type": "totp",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Setup SMS MFA",
			payload: map[string]interface{}{
				"type": "sms",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Setup Email MFA",
			payload: map[string]interface{}{
				"type": "email",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid MFA type",
			payload: map[string]interface{}{
				"type": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing type",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userID", "user1")

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/v1/security/mfa/setup", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			SetupMFA(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "mfa_id")

				if tt.payload["type"] == "totp" {
					assert.Contains(t, response, "secret")
					assert.Contains(t, response, "qr_code_url")
				}
			}
		})
	}
}

func TestVerifyMFASetup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mfaID          string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name:  "Verify with correct code",
			mfaID: "1",
			payload: map[string]interface{}{
				"code": "123456",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Verify with incorrect code",
			mfaID: "1",
			payload: map[string]interface{}{
				"code": "000000",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Verify with invalid code format",
			mfaID: "1",
			payload: map[string]interface{}{
				"code": "abc",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing code",
			mfaID:          "1",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userID", "user1")
			c.Params = gin.Params{
				{Key: "id", Value: tt.mfaID},
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/v1/security/mfa/"+tt.mfaID+"/verify", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			VerifyMFASetup(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "verified")
				assert.Contains(t, response, "backup_codes")
				assert.Equal(t, true, response["verified"])
			}
		})
	}
}

func TestListMFAMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "user1")

	req := httptest.NewRequest("GET", "/api/v1/security/mfa/methods", nil)
	c.Request = req

	ListMFAMethods(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "methods")
}

func TestDeleteMFAMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mfaID          string
		expectedStatus int
	}{
		{
			name:           "Delete existing MFA method",
			mfaID:          "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Delete non-existent MFA method",
			mfaID:          "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userID", "user1")
			c.Params = gin.Params{
				{Key: "id", Value: tt.mfaID},
			}

			req := httptest.NewRequest("DELETE", "/api/v1/security/mfa/"+tt.mfaID, nil)
			c.Request = req

			DeleteMFAMethod(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreateIPWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Add valid IP address",
			payload: map[string]interface{}{
				"ip_address":  "192.168.1.100",
				"description": "Office IP",
				"enabled":     true,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Add valid CIDR range",
			payload: map[string]interface{}{
				"ip_address":  "10.0.0.0/24",
				"description": "VPN subnet",
				"enabled":     true,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Add invalid IP address",
			payload: map[string]interface{}{
				"ip_address": "999.999.999.999",
				"enabled":    true,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Add invalid CIDR",
			payload: map[string]interface{}{
				"ip_address": "192.168.1.0/99",
				"enabled":    true,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing IP address",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userID", "user1")

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/v1/security/ip-whitelist", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			CreateIPWhitelist(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "id")
			}
		})
	}
}

func TestListIPWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "user1")

	req := httptest.NewRequest("GET", "/api/v1/security/ip-whitelist", nil)
	c.Request = req

	ListIPWhitelist(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "entries")
}

func TestDeleteIPWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		entryID        string
		expectedStatus int
	}{
		{
			name:           "Delete existing entry",
			entryID:        "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Delete non-existent entry",
			entryID:        "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userID", "user1")
			c.Params = gin.Params{
				{Key: "id", Value: tt.entryID},
			}

			req := httptest.NewRequest("DELETE", "/api/v1/security/ip-whitelist/"+tt.entryID, nil)
			c.Request = req

			DeleteIPWhitelist(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetSecurityAlerts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{
			name:           "Get all alerts",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get alerts by severity",
			query:          "?severity=high",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get alerts by status",
			query:          "?status=open",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userID", "user1")

			req := httptest.NewRequest("GET", "/api/v1/security/alerts"+tt.query, nil)
			c.Request = req

			GetSecurityAlerts(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "alerts")
		})
	}
}

func TestValidateIPAddress(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		isValid bool
	}{
		{"Valid IPv4", "192.168.1.1", true},
		{"Valid IPv4 - localhost", "127.0.0.1", true},
		{"Valid IPv4 - broadcast", "255.255.255.255", true},
		{"Invalid IPv4 - too many octets", "192.168.1.1.1", false},
		{"Invalid IPv4 - out of range", "256.1.1.1", false},
		{"Invalid IPv4 - letters", "192.168.a.1", false},
		{"Valid CIDR", "192.168.1.0/24", true},
		{"Valid CIDR - /32", "192.168.1.1/32", true},
		{"Invalid CIDR - bad prefix", "192.168.1.0/33", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := isValidIPOrCIDR(tt.ip)
			assert.Equal(t, tt.isValid, valid)
		})
	}
}

func TestGenerateBackupCodes(t *testing.T) {
	codes := generateBackupCodes(10)

	assert.Len(t, codes, 10)

	// Check all codes are unique
	uniqueCodes := make(map[string]bool)
	for _, code := range codes {
		assert.NotEmpty(t, code)
		assert.False(t, uniqueCodes[code], "Duplicate backup code found")
		uniqueCodes[code] = true

		// Check format: XXXXXX-XXXXXX
		assert.Len(t, code, 13) // 6 + 1 (dash) + 6
		assert.Equal(t, "-", string(code[6]))
	}
}

// Helper functions for validation
func isValidIPOrCIDR(ipStr string) bool {
	if ipStr == "" {
		return false
	}
	// Simple validation - in real implementation would use net.ParseIP and net.ParseCIDR
	// For testing purposes, basic validation
	return len(ipStr) >= 7 // Minimum "0.0.0.0"
}

func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// Generate format: ABCDEF-123456
		codes[i] = "ABC123-DEF456" // Mock implementation
	}
	return codes
}
