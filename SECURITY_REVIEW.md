# Security and Code Review - Enterprise Features

**Review Date:** 2025-11-15
**Scope:** Enterprise Features (Integration Hub, Security/MFA, Scheduling, Scaling, Compliance)
**Reviewer:** AI Code Analysis
**Status:** ⚠️ CRITICAL ISSUES FOUND - DO NOT DEPLOY TO PRODUCTION

---

## 🚨 CRITICAL SECURITY ISSUES (Must Fix Before Production)

### 1. SQL Injection Vulnerabilities ⚠️ **SEVERITY: CRITICAL**

**Location:** `api/internal/handlers/integrations.go`

**Lines 183-186:**
```go
if enabled != "" {
    query += fmt.Sprintf(" AND enabled = $%d", argCount)
    args = append(args, enabled == "true")
    argCount++
}
```

**Issue:** While parameterized queries are used for values, dynamic query building with string concatenation is error-prone.

**Also Found In:**
- `integrations.go:456-465` (ListIntegrations)
- `compliance.go:382-401` (ListViolations)

**Fix Required:**
- Use query builder library (e.g., `squirrel`)
- OR carefully validate all query construction
- Add SQL injection tests

---

### 2. WebSocket Origin Validation Bypass ⚠️ **SEVERITY: CRITICAL**

**Location:** `api/internal/handlers/websocket_enterprise.go:44-46`

```go
CheckOrigin: func(r *http.Request) bool {
    return true // Configure appropriately for production
},
```

**Issue:** Allows **ANY** website to connect to WebSocket endpoints. Enables Cross-Site WebSocket Hijacking (CSWSH) attacks.

**Attack Scenario:**
1. User visits malicious website while logged into StreamSpace
2. Malicious JS connects to `wss://streamspace.example.com/api/v1/ws/enterprise`
3. Attacker receives all real-time security alerts, webhook data, compliance violations

**Fix Required:**
```go
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    allowedOrigins := []string{
        "https://streamspace.example.com",
        "https://app.streamspace.io",
    }
    for _, allowed := range allowedOrigins {
        if origin == allowed {
            return true
        }
    }
    return false
},
```

---

### 3. Incomplete MFA Implementation - Security Bypass ⚠️ **SEVERITY: CRITICAL**

**Location:** `api/internal/handlers/security.go:179-182`

```go
} else {
    // TODO: Verify SMS/Email code from cache/temporary storage
    valid = true // Placeholder
}
```

**Issue:** SMS and Email MFA verification **ALWAYS RETURNS TRUE**. Any code is accepted as valid!

**Attack Scenario:**
1. Attacker sets up SMS/Email MFA on compromised account
2. MFA verification always succeeds with any 6-digit code
3. MFA provides zero security

**Also Missing:**
- Line 133-138: SMS sending not implemented (TODO)
- Line 135: Email sending not implemented (TODO)

**Fix Required:**
- Implement Redis/cache storage for verification codes
- Add 5-minute expiration
- Add rate limiting (max 5 attempts)
- OR **remove SMS/Email MFA options until implemented**

**Temporary Mitigation:**
```go
if mfaMethod.Type == "sms" || mfaMethod.Type == "email" {
    c.JSON(http.StatusNotImplemented, gin.H{
        "error": "SMS/Email MFA not yet implemented"
    })
    return
}
```

---

### 4. Secrets Exposure in API Responses ⚠️ **SEVERITY: HIGH**

**Location:** `api/internal/handlers/integrations.go:25`

```go
Secret      string                 `json:"secret,omitempty"`
```

**Issue:** Webhook secrets are returned in API responses when listing/retrieving webhooks.

**Also Found:**
- `security.go:129`: TOTP secrets returned to client
- `security.go:29`: MFA secrets stored in plaintext

**Fix Required:**
1. **Never return secrets in GET requests:**
```go
type Webhook struct {
    // ... other fields
    Secret string `json:"-"` // Never serialize
}

type WebhookWithSecret struct {
    Webhook
    Secret string `json:"secret"` // Only for POST response
}
```

2. **Hash MFA secrets in database:**
```go
// Store: bcrypt(secret)
// Verify: bcrypt.Compare(providedSecret, storedHash)
```

3. **Use encryption for stored secrets:**
```go
import "crypto/aes"
// Encrypt secrets before storing
// Decrypt only when needed for verification
```

---

### 5. Server-Side Request Forgery (SSRF) ⚠️ **SEVERITY: HIGH**

**Location:** `api/internal/handlers/integrations.go:550-590` (deliverWebhook)

**Issue:** No validation that webhook URLs don't point to internal services.

**Attack Scenario:**
1. Attacker creates webhook with URL: `http://169.254.169.254/latest/meta-data/`
2. Webhook fired, server makes request to AWS metadata service
3. Attacker receives cloud credentials in webhook delivery logs

**Fix Required:**
```go
func (h *Handler) validateWebhookURL(urlStr string) error {
    parsed, err := url.Parse(urlStr)
    if err != nil {
        return err
    }

    // Block private IP ranges
    host := parsed.Hostname()
    ip := net.ParseIP(host)
    if ip != nil {
        if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
            return errors.New("webhook URL cannot point to private IP")
        }
    }

    // Block cloud metadata endpoints
    blockedHosts := []string{
        "169.254.169.254",  // AWS/Azure metadata
        "metadata.google.internal", // GCP metadata
    }
    for _, blocked := range blockedHosts {
        if strings.Contains(host, blocked) {
            return errors.New("webhook URL is not allowed")
        }
    }

    return nil
}
```

**Also Add Timeout:**
```go
client := &http.Client{
    Timeout: 10 * time.Second, // Currently 30s, reduce
    // Disable redirects to prevent bypass
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    },
}
```

---

### 6. Missing Rate Limiting on MFA Verification ⚠️ **SEVERITY: HIGH**

**Location:** `api/internal/handlers/security.go:210-278` (VerifyMFA)

**Issue:** No rate limiting on MFA code attempts. Allows brute force attacks.

**Attack Scenario:**
1. TOTP codes are 6 digits (000000-999999 = 1 million combinations)
2. Valid for 30 seconds, so ~3 codes valid at any time
3. Without rate limiting, attacker can try all codes in minutes

**Fix Required:**
```go
// Add to VerifyMFA function
func (h *Handler) VerifyMFA(c *gin.Context) {
    userID := c.GetString("user_id")

    // Check rate limit: max 5 attempts per minute
    attempts := h.getRateLimitAttempts(userID, "mfa_verify")
    if attempts >= 5 {
        c.JSON(http.StatusTooManyRequests, gin.H{
            "error": "Too many verification attempts. Try again in 1 minute.",
        })
        return
    }

    // ... rest of verification

    if !valid {
        h.incrementRateLimitAttempts(userID, "mfa_verify", 60) // 60 second TTL
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid MFA code"})
        return
    }

    // Success - reset counter
    h.resetRateLimitAttempts(userID, "mfa_verify")
}
```

---

### 7. Race Condition in WebSocket Hub ⚠️ **SEVERITY: MEDIUM**

**Location:** `api/internal/handlers/websocket_enterprise.go:87-99`

```go
case message := <-h.Broadcast:
    h.Mu.RLock()
    for _, client := range h.Clients {
        select {
        case client.Send <- message:
        default:
            // Client send buffer full, close connection
            close(client.Send)
            delete(h.Clients, client.ID) // ⚠️ RACE: Modifying map during RLock
        }
    }
    h.Mu.RUnlock()
```

**Issue:** Using `RLock` (read lock) but modifying map with `delete`. This violates lock semantics and can cause panics.

**Fix Required:**
```go
case message := <-h.Broadcast:
    h.Mu.RLock()
    clientsToRemove := []string{}
    for id, client := range h.Clients {
        select {
        case client.Send <- message:
        default:
            close(client.Send)
            clientsToRemove = append(clientsToRemove, id)
        }
    }
    h.Mu.RUnlock()

    // Now acquire write lock to remove clients
    if len(clientsToRemove) > 0 {
        h.Mu.Lock()
        for _, id := range clientsToRemove {
            delete(h.Clients, id)
        }
        h.Mu.Unlock()
    }
```

---

## ⚠️ SECURITY CONCERNS (Should Fix)

### 8. Missing CSRF Protection

**Issue:** No CSRF tokens visible in any handlers. All state-changing operations vulnerable.

**Recommendation:**
- Add CSRF middleware to Gin router
- Use `gorilla/csrf` package
- Require `X-CSRF-Token` header on all POST/PUT/DELETE requests

---

### 9. Weak Device Fingerprinting

**Location:** `api/internal/handlers/security.go:785-791`

```go
func (h *Handler) getDeviceFingerprint(c *gin.Context) string {
    data := c.Request.UserAgent() + c.ClientIP()
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

**Issue:** User-Agent + IP is easily spoofed. Not reliable for security decisions.

**Recommendation:** Use more robust fingerprinting or don't rely on it for security.

---

### 10. Missing Database Transactions

**Issue:** No transactions seen anywhere. Multi-step operations can leave inconsistent state.

**Example Risk:** `VerifyMFASetup` (security.go:143-208)
1. Updates MFA method to verified=true
2. Generates backup codes
3. If step 2 fails, MFA is enabled but user has no backup codes

**Fix Required:**
```go
tx, err := h.DB.Begin()
if err != nil {
    return err
}
defer tx.Rollback() // Rolled back if not committed

// ... operations using tx instead of h.DB

if err := tx.Commit(); err != nil {
    return err
}
```

---

### 11. Authorization After Fetch (Information Disclosure)

**Pattern Found In:** Multiple handlers

**Example:** `security.go:595-612` (DeleteIPWhitelist)

```go
// Check ownership
var ownerID sql.NullString
err := h.DB.QueryRow(`SELECT user_id FROM ip_whitelist WHERE id = $1`, entryID).Scan(&ownerID)
if err == sql.ErrNoRows {
    c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
    return
}

// Only admins can delete org-wide rules or other users' rules
if ownerID.Valid && ownerID.String != userID && role != "admin" {
    c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete other users' IP rules"})
    return
}
```

**Issue:** Returns "not found" vs "forbidden" - allows attacker to enumerate valid IDs.

**Fix:**
```go
// Combine ownership check with fetch
err := h.DB.QueryRow(`
    SELECT user_id FROM ip_whitelist
    WHERE id = $1 AND (user_id = $2 OR $3 = 'admin')
`, entryID, userID, role).Scan(&ownerID)

if err == sql.ErrNoRows {
    // Could be not found OR not authorized - don't reveal which
    c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
    return
}
```

---

### 12. IP Whitelist Logic Flaw

**Location:** `api/internal/handlers/security.go:502-545`

```go
// If rules exist but no match found, deny
return !hasRules
```

**Issue:** Logic seems inverted:
- If NO rules exist → hasRules=false → return true (ALLOW)
- If rules exist but no match → hasRules=true → return false (DENY)

This means "no whitelist = allow all" which might not be intended.

**Clarification Needed:** Is this the intended behavior? Document it clearly.

---

### 13. No Webhook Signature Verification for Incoming Webhooks

**Location:** Missing entirely

**Issue:** While outgoing webhooks sign payloads (line 572-575), there's no verification for incoming webhooks if the system supports them.

**Recommendation:** If receiving webhooks, implement signature verification.

---

## 🔧 INCOMPLETE IMPLEMENTATIONS

### 14. Calendar OAuth Not Implemented

**Location:** `api/internal/handlers/scheduling.go`

**TODOs:**
- Line 446-448: OAuth token exchange
- Line 559-562: Calendar sync implementation
- Line 719-727: Placeholder OAuth URLs

**Status:** Feature advertised but non-functional

**Recommendation:**
- Remove calendar integration UI until implemented
- OR add "Coming Soon" banner
- OR implement using `golang.org/x/oauth2`

---

### 15. Compliance Violation Actions Not Implemented

**Location:** `api/internal/handlers/compliance.go:355`

```go
// TODO: Implement violation actions (notify, create ticket, etc.)
```

**Impact:** Violations recorded but no automated response occurs.

**Recommendation:** Implement notification system or remove violation action configuration from UI.

---

### 16. Email Integration Testing Returns False Success

**Location:** `api/internal/handlers/integrations.go:647-648`

```go
case "email":
    // Would integrate with SMTP
    return true, "Email integration configured (SMTP test not implemented)"
```

**Issue:** Returns success without actually testing email sending.

---

### 17. Missing Error Logging

**Pattern:** Throughout all handlers

```go
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
    return
}
```

**Issue:** No structured logging, no error tracking, no alerts.

**Recommendation:**
```go
if err != nil {
    log.Error().
        Err(err).
        Str("user_id", userID).
        Str("handler", "CreateWebhook").
        Msg("Failed to create webhook")

    c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
    return
}
```

Use: `github.com/rs/zerolog` or `go.uber.org/zap`

---

### 18. Frontend Token Storage in localStorage

**Location:** `ui/src/hooks/useEnterpriseWebSocket.ts:68`

```typescript
const token = localStorage.getItem('token');
```

**Issue:** JWT tokens in localStorage vulnerable to XSS attacks.

**Recommendation:**
- Use httpOnly cookies for auth tokens
- OR use sessionStorage (cleared on tab close)
- Add Content Security Policy headers

---

## 📋 CODE QUALITY ISSUES

### 19. Ignored JSON Unmarshal Errors

**Location:** `api/internal/handlers/integrations.go:206-222`

```go
if err == nil {
    if events.Valid && events.String != "" {
        json.Unmarshal([]byte(events.String), &w.Events) // Error ignored
    }
    // ... more unmarshals with ignored errors
    webhooks = append(webhooks, w)
}
```

**Fix:**
```go
if err != nil {
    log.Warn().Err(err).Msg("Failed to unmarshal events")
    continue // Skip malformed records
}
```

---

### 20. Magic Numbers and Hardcoded Values

**Examples:**
- `websocket_enterprise.go:169`: `54 * time.Second` (why 54?)
- `websocket_enterprise.go:41`: `1024` buffer sizes
- `websocket_enterprise.go:61`: `256` channel buffer

**Recommendation:** Extract to constants with explanatory names.

---

### 21. Missing Input Validation

**Examples:**
- Webhook name length not checked (could be 1MB string)
- Email format not validated
- Phone number format not validated
- CIDR notation not fully validated

**Recommendation:** Use validation library:
```go
import "github.com/go-playground/validator/v10"

type Webhook struct {
    Name   string `json:"name" validate:"required,min=1,max=200"`
    URL    string `json:"url" validate:"required,url"`
    Events []string `json:"events" validate:"required,min=1,dive,required"`
}
```

---

### 22. No Request Size Limits

**Issue:** No limits on:
- Webhook payload size
- Number of events subscribed
- Number of backup codes requested
- WebSocket message size

**Recommendation:**
```go
router.Use(gin.Recovery())
router.Use(middleware.RequestSizeLimiter(10 * 1024 * 1024)) // 10MB max
```

---

### 23. Sensitive Data in Logs

**Potential Issue:** Ensure logs don't contain:
- TOTP secrets
- Webhook secrets
- Backup codes
- IP addresses (may be PII under GDPR)

**Recommendation:** Audit all log statements.

---

## 🎯 RECOMMENDATIONS

### Priority 1 (Critical - Fix Before Any Deployment):
1. ✅ Fix WebSocket CheckOrigin validation
2. ✅ Disable SMS/Email MFA or implement verification
3. ✅ Add rate limiting on MFA verification
4. ✅ Fix race condition in WebSocket hub
5. ✅ Add SSRF protection to webhook delivery
6. ✅ Remove secrets from API responses

### Priority 2 (High - Fix Before Production):
7. ✅ Implement database transactions for multi-step operations
8. ✅ Add CSRF protection
9. ✅ Add comprehensive error logging
10. ✅ Fix authorization checks to prevent enumeration
11. ✅ Add input validation with size limits
12. ✅ Move JWT tokens from localStorage to httpOnly cookies

### Priority 3 (Medium - Security Hardening):
13. ✅ Implement calendar OAuth or remove feature
14. ✅ Implement compliance violation actions
15. ✅ Add SQL injection tests
16. ✅ Hash/encrypt stored secrets
17. ✅ Add request size limits
18. ✅ Improve device fingerprinting or remove reliance on it

### Priority 4 (Low - Code Quality):
19. ✅ Add structured logging throughout
20. ✅ Extract magic numbers to constants
21. ✅ Handle JSON unmarshal errors
22. ✅ Add comprehensive unit tests for security functions
23. ✅ Add integration tests for webhooks

---

## 🔒 Security Testing Checklist

Before production deployment, test:

- [ ] SQL injection attempts on all filter endpoints
- [ ] SSRF via webhook URLs (try 169.254.169.254, localhost, etc.)
- [ ] Cross-Site WebSocket Hijacking
- [ ] MFA brute force (should be rate-limited)
- [ ] Webhook signature verification bypass attempts
- [ ] XSS in webhook/integration names and descriptions
- [ ] CSRF on all state-changing operations
- [ ] Authorization bypass (user A accessing user B's resources)
- [ ] Race conditions (concurrent requests to same resources)
- [ ] Large payload attacks (10GB webhook payload, etc.)
- [ ] WebSocket connection exhaustion (open 10,000 connections)
- [ ] Calendar OAuth flow security
- [ ] Compliance data export security
- [ ] IP whitelist bypass attempts

---

## 📊 Summary

| Category | Count | Severity |
|----------|-------|----------|
| Critical Security Issues | 7 | 🔴 CRITICAL |
| Security Concerns | 6 | 🟡 HIGH |
| Incomplete Features | 5 | 🟠 MEDIUM |
| Code Quality Issues | 5 | 🔵 LOW |
| **Total Issues** | **23** | |

**Recommendation:** **DO NOT DEPLOY TO PRODUCTION** until at minimum all Priority 1 (Critical) issues are resolved.

---

## 📝 Next Steps

1. **Immediate:** Disable SMS/Email MFA options in UI
2. **Week 1:** Fix all Priority 1 critical issues
3. **Week 2:** Implement Priority 2 high-priority fixes
4. **Week 3:** Security testing and penetration testing
5. **Week 4:** Code review with security team
6. **Week 5+:** Address remaining medium/low priority issues

---

**Review Completed:** 2025-11-15
**Recommended Re-Review After:** Fixes implemented
**Security Sign-off Required:** Yes ✅
