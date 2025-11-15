# Critical Security Fixes - Action Required

**Status:** 🚨 **BLOCKING PRODUCTION DEPLOYMENT**
**Timeline:** Must be completed before any production release
**Owner:** Development Team
**Review Date:** 2025-11-15

---

## 🔥 CRITICAL FIXES (Must Complete Immediately)

### Fix #1: WebSocket Origin Validation

**File:** `api/internal/handlers/websocket_enterprise.go`
**Line:** 44-46
**Current Code:**
```go
CheckOrigin: func(r *http.Request) bool {
    return true // ⚠️ ALLOWS ANY ORIGIN
},
```

**Fixed Code:**
```go
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")

    // Allow same-origin requests
    if origin == "" {
        return true
    }

    // Parse allowed origins from config
    allowedOrigins := []string{
        os.Getenv("ALLOWED_ORIGIN_1"), // e.g., "https://streamspace.example.com"
        os.Getenv("ALLOWED_ORIGIN_2"), // e.g., "https://app.streamspace.io"
    }

    for _, allowed := range allowedOrigins {
        if allowed != "" && origin == allowed {
            return true
        }
    }

    log.Warn().Str("origin", origin).Msg("WebSocket connection rejected: invalid origin")
    return false
},
```

**Test:**
```bash
# Should succeed
curl -i -N \
  -H "Origin: https://streamspace.example.com" \
  -H "Upgrade: websocket" \
  http://localhost:8080/api/v1/ws/enterprise

# Should fail (403)
curl -i -N \
  -H "Origin: https://malicious-site.com" \
  -H "Upgrade: websocket" \
  http://localhost:8080/api/v1/ws/enterprise
```

---

### Fix #2: Disable Incomplete MFA Methods

**File:** `api/internal/handlers/security.go`
**Line:** 62-141

**Add validation at start of SetupMFA:**
```go
func (h *Handler) SetupMFA(c *gin.Context) {
    userID := c.GetString("user_id")

    var req struct {
        Type        string `json:"type" binding:"required,oneof=totp sms email"`
        PhoneNumber string `json:"phone_number,omitempty"`
        Email       string `json:"email,omitempty"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // ⚠️ ADD THIS CHECK
    if req.Type == "sms" || req.Type == "email" {
        c.JSON(http.StatusNotImplemented, gin.H{
            "error": "SMS and Email MFA are not yet implemented",
            "message": "Please use TOTP (authenticator app) for multi-factor authentication",
        })
        return
    }

    // ... rest of function
}
```

**Also fix VerifyMFA at line 210:**
```go
func (h *Handler) VerifyMFA(c *gin.Context) {
    userID := c.GetString("user_id")

    var req struct {
        Code       string `json:"code" binding:"required"`
        MethodType string `json:"method_type,omitempty"`
        TrustDevice bool  `json:"trust_device,omitempty"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if req.MethodType == "" {
        req.MethodType = "totp"
    }

    // ⚠️ ADD THIS CHECK
    if req.MethodType == "sms" || req.MethodType == "email" {
        c.JSON(http.StatusNotImplemented, gin.H{
            "error": "SMS and Email MFA are not yet implemented",
        })
        return
    }

    // ... rest of function
}
```

**Frontend Fix:**
Update `ui/src/pages/SecuritySettings.tsx` to hide SMS/Email options:

```typescript
// Around line 98-100, replace MFA type selection
const handleStartMFASetup = (type: 'totp' | 'sms' | 'email') => {
    if (type === 'sms' || type === 'email') {
        alert('SMS and Email MFA are not yet available. Please use TOTP (Authenticator App).');
        return;
    }
    setMfaType(type);
    setMfaStep(0);
    setMfaDialog(true);
    // ... rest
};
```

---

### Fix #3: Add MFA Rate Limiting

**File:** `api/internal/handlers/security.go`
**Create new file:** `api/internal/middleware/ratelimit.go`

```go
package middleware

import (
    "sync"
    "time"
)

// Simple in-memory rate limiter (use Redis in production)
type RateLimiter struct {
    attempts map[string][]time.Time
    mu       sync.RWMutex
}

var globalRateLimiter = &RateLimiter{
    attempts: make(map[string][]time.Time),
}

func (rl *RateLimiter) CheckLimit(key string, maxAttempts int, window time.Duration) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()

    // Clean old attempts
    if attempts, exists := rl.attempts[key]; exists {
        validAttempts := []time.Time{}
        for _, t := range attempts {
            if now.Sub(t) < window {
                validAttempts = append(validAttempts, t)
            }
        }
        rl.attempts[key] = validAttempts
    }

    // Check if limit exceeded
    if len(rl.attempts[key]) >= maxAttempts {
        return false
    }

    // Record attempt
    rl.attempts[key] = append(rl.attempts[key], now)
    return true
}

func GetRateLimiter() *RateLimiter {
    return globalRateLimiter
}
```

**Update VerifyMFA in security.go:**

```go
import "your-project/api/internal/middleware"

func (h *Handler) VerifyMFA(c *gin.Context) {
    userID := c.GetString("user_id")

    // ⚠️ ADD RATE LIMITING
    rateLimitKey := fmt.Sprintf("mfa_verify:%s", userID)
    if !middleware.GetRateLimiter().CheckLimit(rateLimitKey, 5, 1*time.Minute) {
        c.JSON(http.StatusTooManyRequests, gin.H{
            "error": "Too many verification attempts",
            "message": "Please wait 1 minute before trying again",
            "retry_after": 60,
        })
        return
    }

    // ... rest of verification

    if !valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid MFA code"})
        return
    }

    // On success, clear rate limit
    middleware.GetRateLimiter().mu.Lock()
    delete(middleware.GetRateLimiter().attempts, rateLimitKey)
    middleware.GetRateLimiter().mu.Unlock()

    // ... rest
}
```

---

### Fix #4: Webhook SSRF Protection

**File:** `api/internal/handlers/integrations.go`
**Add validation function:**

```go
import (
    "net"
    "net/url"
    "strings"
)

func (h *Handler) validateWebhookURL(urlStr string) error {
    parsed, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("invalid URL format: %w", err)
    }

    // Must be http or https
    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        return fmt.Errorf("URL must use http or https")
    }

    host := parsed.Hostname()

    // Resolve hostname to IP
    ips, err := net.LookupIP(host)
    if err != nil {
        return fmt.Errorf("could not resolve hostname: %w", err)
    }

    // Check each resolved IP
    for _, ip := range ips {
        if ip.IsLoopback() {
            return fmt.Errorf("webhook URL cannot point to loopback address")
        }
        if ip.IsPrivate() {
            return fmt.Errorf("webhook URL cannot point to private IP address")
        }
        if ip.IsLinkLocalUnicast() {
            return fmt.Errorf("webhook URL cannot point to link-local address")
        }

        // Block cloud metadata endpoints
        if ip.String() == "169.254.169.254" {
            return fmt.Errorf("webhook URL is not allowed")
        }
    }

    // Block specific hostnames
    blockedHosts := []string{
        "metadata.google.internal",
        "169.254.169.254",
        "localhost",
    }
    for _, blocked := range blockedHosts {
        if strings.Contains(strings.ToLower(host), blocked) {
            return fmt.Errorf("webhook URL is not allowed")
        }
    }

    return nil
}
```

**Update CreateWebhook (line 125-129):**

```go
// Validate URL
if webhook.URL == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
    return
}

// ⚠️ ADD SSRF VALIDATION
if err := h.validateWebhookURL(webhook.URL); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "Invalid webhook URL",
        "details": err.Error(),
    })
    return
}
```

**Update deliverWebhook (line 578):**

```go
// Send request with restrictions
client := &http.Client{
    Timeout: 10 * time.Second, // Reduced from 30s
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        // Disable redirects to prevent SSRF bypass
        return http.ErrUseLastResponse
    },
}
```

---

### Fix #5: Remove Secrets from API Responses

**File:** `api/internal/handlers/integrations.go`

**Update Webhook struct:**

```go
// Webhook represents a webhook configuration
type Webhook struct {
    ID          int64                  `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    URL         string                 `json:"url"`
    Secret      string                 `json:"-"` // ⚠️ Never serialize to JSON
    Events      []string               `json:"events"`
    Headers     map[string]string      `json:"headers,omitempty"`
    Enabled     bool                   `json:"enabled"`
    RetryPolicy WebhookRetryPolicy     `json:"retry_policy"`
    Filters     WebhookFilters         `json:"filters,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedBy   string                 `json:"created_by"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// WebhookWithSecret - only used for creation response
type WebhookWithSecret struct {
    Webhook
    Secret string `json:"secret"`
}
```

**Update CreateWebhook response (line 167):**

```go
c.JSON(http.StatusCreated, WebhookWithSecret{
    Webhook: webhook,
    Secret:  webhook.Secret, // Only show secret on creation
})
```

**Same for security.go MFA secrets:**

```go
type MFAMethod struct {
    ID          int64     `json:"id"`
    UserID      string    `json:"user_id"`
    Type        string    `json:"type"`
    Enabled     bool      `json:"enabled"`
    Secret      string    `json:"-"` // ⚠️ Never expose
    PhoneNumber string    `json:"phone_number,omitempty"`
    Email       string    `json:"email,omitempty"`
    IsPrimary   bool      `json:"is_primary"`
    Verified    bool      `json:"verified"`
    CreatedAt   time.Time `json:"created_at"`
    LastUsedAt  time.Time `json:"last_used_at,omitempty"`
}

type MFASetupResponse struct {
    ID       int64  `json:"id"`
    Type     string `json:"type"`
    Secret   string `json:"secret,omitempty"`   // Only for TOTP setup
    QRCode   string `json:"qr_code,omitempty"`  // Only for TOTP setup
    Message  string `json:"message"`
}
```

---

### Fix #6: Fix WebSocket Race Condition

**File:** `api/internal/handlers/websocket_enterprise.go`
**Lines:** 87-99

**Replace with:**

```go
case message := <-h.Broadcast:
    // Collect clients to remove (don't modify map during iteration)
    clientsToRemove := make([]*WebSocketClient, 0)

    h.Mu.RLock()
    for _, client := range h.Clients {
        select {
        case client.Send <- message:
            // Message sent successfully
        default:
            // Client buffer full - mark for removal
            clientsToRemove = append(clientsToRemove, client)
        }
    }
    h.Mu.RUnlock()

    // Now safely remove disconnected clients
    if len(clientsToRemove) > 0 {
        h.Mu.Lock()
        for _, client := range clientsToRemove {
            if ch, exists := h.Clients[client.ID]; exists {
                close(ch.Send)
                delete(h.Clients, client.ID)
                log.Printf("Removed disconnected WebSocket client: %s", client.ID)
            }
        }
        h.Mu.Unlock()
    }
```

---

### Fix #7: Add Database Transactions

**File:** `api/internal/handlers/security.go`
**Function:** `VerifyMFASetup` (lines 143-208)

**Replace with:**

```go
func (h *Handler) VerifyMFASetup(c *gin.Context) {
    userID := c.GetString("user_id")
    mfaID := c.Param("mfaId")

    var req struct {
        Code string `json:"code" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // ⚠️ START TRANSACTION
    tx, err := h.DB.Begin()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
        return
    }
    defer tx.Rollback() // Rollback if not committed

    // Get MFA method using transaction
    var mfaMethod MFAMethod
    err = tx.QueryRow(`
        SELECT id, user_id, type, secret, phone_number, email
        FROM mfa_methods
        WHERE id = $1 AND user_id = $2
    `, mfaID, userID).Scan(&mfaMethod.ID, &mfaMethod.UserID, &mfaMethod.Type,
        &mfaMethod.Secret, &mfaMethod.PhoneNumber, &mfaMethod.Email)

    if err == sql.ErrNoRows {
        c.JSON(http.StatusNotFound, gin.H{"error": "MFA method not found"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
        return
    }

    // Verify code
    valid := false
    if mfaMethod.Type == "totp" {
        valid = totp.Validate(req.Code, mfaMethod.Secret)
    }

    if !valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid verification code"})
        return
    }

    // Enable and verify MFA method
    _, err = tx.Exec(`
        UPDATE mfa_methods
        SET verified = true, enabled = true
        WHERE id = $1
    `, mfaID)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enable MFA"})
        return
    }

    // Generate backup codes within transaction
    backupCodes := make([]string, 10)
    for i := 0; i < 10; i++ {
        code := generateRandomCode(8)
        backupCodes[i] = code

        hash := sha256.Sum256([]byte(code))
        hashStr := hex.EncodeToString(hash[:])

        _, err := tx.Exec(`
            INSERT INTO backup_codes (user_id, code)
            VALUES ($1, $2)
        `, userID, hashStr)

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate backup codes"})
            return
        }
    }

    // ⚠️ COMMIT TRANSACTION
    if err := tx.Commit(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit changes"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message":      "MFA enabled successfully",
        "backup_codes": backupCodes,
    })
}
```

---

## 📋 Testing Checklist

After implementing fixes, verify:

- [ ] **Fix #1:** WebSocket rejects connections from unauthorized origins
- [ ] **Fix #2:** SMS/Email MFA returns 501 Not Implemented
- [ ] **Fix #3:** MFA rate limiting blocks after 5 attempts
- [ ] **Fix #4:** Webhook creation rejects private IPs and metadata endpoints
- [ ] **Fix #5:** GET /webhooks does not return secrets
- [ ] **Fix #6:** WebSocket hub handles 1000+ concurrent connections without panic
- [ ] **Fix #7:** Failed backup code generation rolls back MFA enable

---

## 🚀 Deployment Plan

1. **Create feature branch:** `fix/critical-security-issues`
2. **Implement all 7 fixes** (estimated 4-6 hours)
3. **Run security tests** (see SECURITY_REVIEW.md)
4. **Code review** with security team
5. **Merge to main** after approval
6. **Deploy to staging** and verify
7. **Production deployment** only after all tests pass

---

## 📞 Need Help?

- Questions about fixes: Contact development lead
- Security concerns: Contact security team
- Testing support: Contact QA team

**Status:** ⚠️ **IN PROGRESS** - Do not deploy until all fixes verified
