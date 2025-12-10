# Input Validation & Security Remediation Plan

**Date Created:** 2025-12-09
**Status:** Pending Implementation
**Priority:** High - Security vulnerabilities identified

## Executive Summary

A comprehensive security audit has identified several input validation gaps in the Backgammon application. While the codebase has a strong foundation with parameterized SQL queries and secure session management, there are critical server-side validation issues that need immediate attention.

**Overall Security Assessment:** MODERATE - Good foundation with specific gaps requiring remediation

---

## Critical Vulnerabilities (HIGH PRIORITY)

### 1. Combined Move Validation Bypass 🔴

**Severity:** High
**Location:** `service/game.go:564-576`
**CVSS Score:** 6.5 (Medium-High)

#### Problem

Server-side validation is skipped for combined game moves with this comment:

```go
// For combined moves, validate by simulating each step
// The frontend already validated this is a legal combined move
```

The server trusts frontend validation, allowing authenticated attackers to potentially execute invalid moves by crafting malicious requests.

#### Current Code

```go
if req.IsCombinedMove && len(req.DiceIndices) > 0 {
    // Only validates dice indices, NOT the move coordinates
    for _, idx := range req.DiceIndices {
        if idx < 0 || idx >= len(state.DiceUsed) {
            util.ErrorResponse(w, http.StatusBadRequest, "Invalid dice index")
            return
        }
        if state.DiceUsed[idx] {
            util.ErrorResponse(w, http.StatusBadRequest, "Die already used")
            return
        }
    }
    diceIndicesToMark = req.DiceIndices
    // ⚠️ NO VALIDATION - goes directly to ExecuteMove()
}
```

#### Fix Required

```go
if req.IsCombinedMove && len(req.DiceIndices) > 0 {
    // Validate dice indices
    for _, idx := range req.DiceIndices {
        if idx < 0 || idx >= len(state.DiceUsed) {
            util.ErrorResponse(w, http.StatusBadRequest, "Invalid dice index")
            return
        }
        if state.DiceUsed[idx] {
            util.ErrorResponse(w, http.StatusBadRequest, "Die already used")
            return
        }
    }
    diceIndicesToMark = req.DiceIndices

    // ✅ ADD THIS: Validate move coordinates
    if req.FromPoint < 0 || req.FromPoint > 25 || req.ToPoint < 0 || req.ToPoint > 25 {
        util.ErrorResponse(w, http.StatusBadRequest, "Invalid point values")
        return
    }

    // ✅ ADD THIS: Validate die value
    if req.DieUsed < 1 || req.DieUsed > 6 {
        util.ErrorResponse(w, http.StatusBadRequest, "Invalid die value")
        return
    }

    // ✅ ADD THIS: Always call server-side validation
    err = business.ValidateMove(state.BoardState, req.FromPoint, req.ToPoint, req.DieUsed, color, barCount)
    if err != nil {
        util.ErrorResponse(w, http.StatusBadRequest, err.Error())
        return
    }
}
```

#### Files to Modify

-   `service/game.go` (lines 564-576)

---

### 2. No Input Length Limits 🔴

**Severity:** High
**Location:** Multiple files
**CVSS Score:** 5.3 (Medium)

#### Problem

No maximum length validation on user inputs, allowing:

-   Database performance degradation
-   Potential denial of service
-   Memory exhaustion
-   Storage waste

#### Missing Validations

**Registration (`service/user.go:76-84`)**

```go
// Current code - only minimum checks
if len(req.Username) < 3 {
    util.ErrorResponse(w, http.StatusBadRequest, "Username must be at least 3 characters")
    return
}
if len(req.Password) < 6 {
    util.ErrorResponse(w, http.StatusBadRequest, "Password must be at least 6 characters")
    return
}
// ❌ No maximum length checks!
```

#### Fix Required

```go
// ✅ Add maximum length validation
const (
    MinUsernameLength = 3
    MaxUsernameLength = 255  // Matches DB VARCHAR(255)
    MinPasswordLength = 8    // Increased from 6 for better security
    MaxPasswordLength = 256  // Reasonable limit for bcrypt
)

// Username validation
if len(req.Username) < MinUsernameLength {
    util.ErrorResponse(w, http.StatusBadRequest,
        fmt.Sprintf("Username must be at least %d characters", MinUsernameLength))
    return
}
if len(req.Username) > MaxUsernameLength {
    util.ErrorResponse(w, http.StatusBadRequest,
        fmt.Sprintf("Username must not exceed %d characters", MaxUsernameLength))
    return
}

// Password validation
if len(req.Password) < MinPasswordLength {
    util.ErrorResponse(w, http.StatusBadRequest,
        fmt.Sprintf("Password must be at least %d characters", MinPasswordLength))
    return
}
if len(req.Password) > MaxPasswordLength {
    util.ErrorResponse(w, http.StatusBadRequest,
        fmt.Sprintf("Password must not exceed %d characters", MaxPasswordLength))
    return
}
```

#### Files to Modify

-   `service/user.go` (registration and login handlers)
-   Create new file: `util/validation.go` for shared validation constants

---

### 3. No Rate Limiting 🔴

**Severity:** Critical
**Location:** All endpoints
**CVSS Score:** 7.5 (High)

#### Problem

No rate limiting on any endpoints allows:

-   Brute force attacks on login/passwords
-   Registration token spam
-   Invitation spam
-   Resource exhaustion (DoS)

#### Vulnerable Endpoints

1. `POST /api/v1/auth/login` - Brute force attack vector
2. `POST /api/v1/auth/register` - Account creation spam
3. `POST /api/v1/auth/register/token` - Token generation spam
4. `POST /api/v1/invitations` - Invitation spam
5. `POST /api/v1/games/{id}/move` - Move spam
6. `PUT /api/v1/lobby/presence/heartbeat` - Heartbeat spam

#### Fix Required

**Step 1: Install rate limiting library**

```bash
go get golang.org/x/time/rate
```

**Step 2: Create rate limiting middleware (`middleware/ratelimit.go`)**

```go
package middleware

import (
    "net/http"
    "sync"
    "time"
    "golang.org/x/time/rate"
    "yourproject/util"
)

type visitor struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

type RateLimiter struct {
    visitors map[string]*visitor
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    rl := &RateLimiter{
        visitors: make(map[string]*visitor),
        rate:     r,
        burst:    b,
    }

    // Cleanup old visitors every 3 minutes
    go rl.cleanupVisitors()

    return rl
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    v, exists := rl.visitors[ip]
    if !exists {
        limiter := rate.NewLimiter(rl.rate, rl.burst)
        rl.visitors[ip] = &visitor{limiter, time.Now()}
        return limiter
    }

    v.lastSeen = time.Now()
    return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
    for {
        time.Sleep(3 * time.Minute)

        rl.mu.Lock()
        for ip, v := range rl.visitors {
            if time.Since(v.lastSeen) > 3*time.Minute {
                delete(rl.visitors, ip)
            }
        }
        rl.mu.Unlock()
    }
}

func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ip := util.GetRealIP(r)
        limiter := rl.getVisitor(ip)

        if !limiter.Allow() {
            util.ErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded")
            return
        }

        next(w, r)
    }
}
```

**Step 3: Apply rate limiters to routes (`main.go` or router setup)**

```go
// Different rate limits for different endpoint types
var (
    // Strict rate limit for auth endpoints: 5 requests per minute
    authLimiter = middleware.NewRateLimiter(rate.Every(12*time.Second), 5)

    // Moderate rate limit for game actions: 30 requests per minute
    gameLimiter = middleware.NewRateLimiter(rate.Every(2*time.Second), 30)

    // Lenient rate limit for reads: 60 requests per minute
    readLimiter = middleware.NewRateLimiter(rate.Every(time.Second), 60)
)

// Apply to routes
mux.HandleFunc("/api/v1/auth/login", authLimiter.Limit(service.LoginHandler))
mux.HandleFunc("/api/v1/auth/register", authLimiter.Limit(service.RegisterHandler))
mux.HandleFunc("/api/v1/auth/register/token", authLimiter.Limit(service.GetRegistrationTokenHandler))
mux.HandleFunc("/api/v1/invitations", authLimiter.Limit(service.CreateInvitationHandler))
mux.HandleFunc("/api/v1/games/{id}/move", gameLimiter.Limit(service.GameMoveHandler))
```

#### Files to Create/Modify

-   Create: `middleware/ratelimit.go` (new file)
-   Modify: `main.go` or router configuration file
-   Ensure: `util/auth.go` has `GetRealIP()` function

---

## Medium Priority Issues (MEDIUM PRIORITY)

### 4. Weak Password Requirements ⚠️

**Severity:** Medium
**Location:** `service/user.go:80-84`
**CVSS Score:** 4.3 (Medium)

#### Problem

Current password requirements are too weak:

-   Minimum 6 characters (should be 8+)
-   No complexity requirements
-   Passwords like "password" or "123456" are accepted

#### Fix Required

**Create password validator (`util/validation.go`)**

```go
package util

import (
    "errors"
    "unicode"
)

const (
    MinPasswordLength = 8
    MaxPasswordLength = 256
)

type PasswordStrength struct {
    HasMinLength   bool
    HasUppercase   bool
    HasLowercase   bool
    HasNumber      bool
    HasSpecial     bool
    ComplexityMet  bool
}

func ValidatePasswordStrength(password string) (*PasswordStrength, error) {
    strength := &PasswordStrength{}

    if len(password) < MinPasswordLength {
        return strength, errors.New("password must be at least 8 characters")
    }
    if len(password) > MaxPasswordLength {
        return strength, errors.New("password must not exceed 256 characters")
    }
    strength.HasMinLength = true

    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            strength.HasUppercase = true
        case unicode.IsLower(char):
            strength.HasLowercase = true
        case unicode.IsNumber(char):
            strength.HasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            strength.HasSpecial = true
        }
    }

    // Require at least 2 of: uppercase, number, special character
    complexityCount := 0
    if strength.HasUppercase {
        complexityCount++
    }
    if strength.HasNumber {
        complexityCount++
    }
    if strength.HasSpecial {
        complexityCount++
    }

    if complexityCount < 2 {
        return strength, errors.New("password must contain at least 2 of: uppercase letter, number, special character")
    }

    strength.ComplexityMet = true
    return strength, nil
}
```

**Update registration handler (`service/user.go`)**

```go
// Replace password length check with complexity validation
strength, err := util.ValidatePasswordStrength(req.Password)
if err != nil {
    util.ErrorResponse(w, http.StatusBadRequest, err.Error())
    return
}
```

#### Files to Create/Modify

-   Create: `util/validation.go`
-   Modify: `service/user.go`

---

### 5. Username Validation Too Permissive ⚠️

**Severity:** Medium
**Location:** `service/user.go:76-79`
**CVSS Score:** 3.9 (Low-Medium)

#### Problem

Username validation only checks minimum length. Allows:

-   Control characters
-   Whitespace
-   Confusable Unicode characters
-   Special characters that could break UI

#### Fix Required

**Add username validator (`util/validation.go`)**

```go
import "regexp"

const (
    MinUsernameLength = 3
    MaxUsernameLength = 255
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func ValidateUsername(username string) error {
    if len(username) < MinUsernameLength {
        return errors.New("username must be at least 3 characters")
    }
    if len(username) > MaxUsernameLength {
        return errors.New("username must not exceed 255 characters")
    }

    if !usernameRegex.MatchString(username) {
        return errors.New("username must contain only letters, numbers, underscores, and hyphens")
    }

    // Optional: Check against reserved words
    reservedWords := []string{"admin", "root", "system", "api", "null", "undefined"}
    usernameLower := strings.ToLower(username)
    for _, reserved := range reservedWords {
        if usernameLower == reserved {
            return errors.New("username is reserved")
        }
    }

    return nil
}
```

**Update registration handler**

```go
if err := util.ValidateUsername(req.Username); err != nil {
    util.ErrorResponse(w, http.StatusBadRequest, err.Error())
    return
}
```

#### Files to Modify

-   `util/validation.go` (add function)
-   `service/user.go` (use validator)

---

### 6. Missing Die Value Validation ⚠️

**Severity:** Medium
**Location:** `service/game.go:594` (single move validation)
**CVSS Score:** 4.0 (Medium)

#### Problem

`req.DieUsed` is never explicitly validated to be in range 1-6 before being passed to business logic.

#### Fix Required

**In game move handler (`service/game.go`)**

```go
// Add this validation before ValidateMove call (line ~594)
if req.DieUsed < 1 || req.DieUsed > 6 {
    util.ErrorResponse(w, http.StatusBadRequest, "Die value must be between 1 and 6")
    return
}

err = business.ValidateMove(state.BoardState, req.FromPoint, req.ToPoint, req.DieUsed, color, barCount)
```

#### Files to Modify

-   `service/game.go` (add validation before ValidateMove calls)

---

### 7. No ID Range Validation ⚠️

**Severity:** Low-Medium
**Location:** `service/game.go:295-306`, `service/invitation.go:385-399`
**CVSS Score:** 3.1 (Low)

#### Problem

Game IDs and Invitation IDs parsed from URL paths are not validated to be positive integers.

#### Fix Required

**Update parseGameIDFromPath (`service/game.go`)**

```go
func parseGameIDFromPath(path string) (int, error) {
    trimmed := strings.TrimPrefix(path, "/api/v1/games/")
    id, err := strconv.Atoi(trimmed)
    if err != nil {
        return 0, err
    }

    // ✅ Add range validation
    if id <= 0 {
        return 0, errors.New("game ID must be a positive integer")
    }

    return id, nil
}
```

**Update parseInvitationIDFromPath (`service/invitation.go`)**

```go
func parseInvitationIDFromPath(path, prefix, suffix string) (int, error) {
    trimmed := strings.TrimPrefix(path, prefix)
    if suffix != "" {
        trimmed = strings.TrimSuffix(trimmed, suffix)
    }
    id, err := strconv.Atoi(trimmed)
    if err != nil {
        return 0, err
    }

    // ✅ Add range validation
    if id <= 0 {
        return 0, errors.New("invitation ID must be a positive integer")
    }

    return id, nil
}
```

#### Files to Modify

-   `service/game.go`
-   `service/invitation.go`

---

## Low Priority Issues (LOW PRIORITY)

### 8. CSRF Protection Enhancement

**Severity:** Low
**Current Status:** Cookies use `SameSite=Lax`

#### Recommendation

Consider upgrading to `SameSite=Strict` for sensitive operations:

-   Game moves
-   Account deletion
-   Password changes
