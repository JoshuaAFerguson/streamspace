# StreamSpace - Immediate Action Plan
**Priority: FIX BUILD ISSUES (Week 1)**

**Goal**: Make all components buildable within 3 hours of focused work

---

## Issue #1: UI TypeScript Compilation Errors ❌

### Problem
30+ TypeScript errors preventing UI from building, primarily in:
- `src/pages/admin/Quotas.tsx` (27 errors)
- `src/pages/admin/UserDetail.tsx` (3 errors)

### Root Cause
Type definition mismatch between `UserQuota` interface and API usage.

### Fix (Estimated: 2 hours)

#### Step 1: Update UserQuota Type Definition (30 min)

**File**: `ui/src/lib/api.ts`

Current (likely):
```typescript
export interface UserQuota {
  userId: string;
  maxSessions: number;
  maxMemory: string;
  maxCPU: string;
  maxStorage: string;
}
```

Should be:
```typescript
export interface UserQuota {
  userId: string;
  username: string;  // ADD THIS
  maxSessions: number;
  maxMemory: string;
  maxCPU: string;
  maxStorage: string;
  limits: {  // ADD THIS
    maxSessions: number;
    maxMemory: string;
    maxCPU: string;
    maxStorage: string;
  };
  used: {  // ADD THIS
    sessions: number;
    memory: string;
    cpu: string;
    storage: string;
  };
}

export interface SetQuotaRequest {
  userId: string;  // Not username
  limits: {
    maxSessions: number;
    maxMemory: string;
    maxCPU: string;
    maxStorage: string;
  };
}
```

#### Step 2: Fix Quotas.tsx (1 hour)

**File**: `ui/src/pages/admin/Quotas.tsx`

Changes needed:
1. Remove unused import:
   ```typescript
   - import { Chip, ... } from '@mui/material';
   + import { ... } from '@mui/material';  // Remove Chip
   ```

2. Update quota access pattern:
   ```typescript
   - {quota.limits.maxSessions}
   + {quota.limits?.maxSessions || quota.maxSessions}

   - {quota.used.sessions}
   + {quota.used?.sessions || 0}

   - username: selectedUsername,
   + userId: selectedUserId,
   ```

3. Update all property accesses to use optional chaining:
   ```typescript
   quota.username → quota.username
   quota.limits.maxSessions → quota.limits?.maxSessions
   quota.used.sessions → quota.used?.sessions
   ```

#### Step 3: Fix UserDetail.tsx (30 min)

**File**: `ui/src/pages/admin/UserDetail.tsx`

Remove unused imports:
```typescript
- import { Tooltip, ... } from '@mui/material';
+ import { ... } from '@mui/material';  // Remove Tooltip

- import { User, UserQuota } from '../lib/api';
+ import { User } from '../lib/api';  // Remove UserQuota if unused
```

### Verification

```bash
cd /home/user/streamspace/ui
npm run build

# Should output:
# vite v5.0.8 building for production...
# ✓ 1234 modules transformed.
# dist/index.html                   0.45 kB │ gzip:   0.30 kB
# dist/assets/index-[hash].css     12.34 kB │ gzip:   3.45 kB
# dist/assets/index-[hash].js     567.89 kB │ gzip: 123.45 kB
# ✓ built in 12.34s
```

---

## Issue #2: API Dependencies Missing ⚠️

### Problem
No `go.sum` file causes build failure.

### Root Cause
Dependencies not resolved (`go mod tidy` not run).

### Fix (Estimated: 5 minutes with network)

#### Step 1: Resolve Dependencies

```bash
cd /home/user/streamspace/api
go mod tidy
```

Expected output:
```
go: downloading github.com/gin-gonic/gin v1.9.1
go: downloading k8s.io/client-go v0.29.0
...
[Many downloads]
...
go: downloading github.com/crewjam/saml v0.5.1
```

This will create `go.sum` file with ~250 entries.

#### Step 2: Verify Build

```bash
cd /home/user/streamspace/api
go build -o /tmp/api-server ./cmd/main.go

# Should output:
# [Compilation messages]
# [No errors]

ls -lh /tmp/api-server
# Should show binary: -rwxr-xr-x 1 root root 45M Nov 14 10:00 /tmp/api-server
```

### Verification

```bash
cd /home/user/streamspace
make build-api

# Should output:
# Building API...
# ✓ API built successfully
```

---

## Issue #3: Controller Already Works ✅

### Status
Controller builds successfully - **no action needed**.

### Verification

```bash
cd /home/user/streamspace/controller
go build -o /tmp/controller ./cmd/main.go
echo $?  # Should output: 0

ls -lh /tmp/controller
# Should show binary: -rwxr-xr-x 1 root root 38M Nov 14 10:00 /tmp/controller
```

---

## Complete Workflow (3 Hours Total)

### Hour 1: UI Fixes

```bash
# 1. Update type definitions (30 min)
cd /home/user/streamspace/ui
vim src/lib/api.ts  # Update UserQuota interface

# 2. Fix Quotas.tsx (30 min)
vim src/pages/admin/Quotas.tsx  # Fix type errors
```

### Hour 2: UI Testing & API Dependencies

```bash
# 3. Fix UserDetail.tsx and test UI (30 min)
vim src/pages/admin/UserDetail.tsx  # Remove unused imports
npm run build  # Verify builds

# 4. Resolve API dependencies (5 min with network)
cd /home/user/streamspace/api
go mod tidy

# 5. Verify API build (5 min)
go build -o /tmp/api-server ./cmd/main.go

# 6. Run controller tests (20 min)
cd /home/user/streamspace/controller
make test
```

### Hour 3: Docker Builds & Verification

```bash
# 7. Build all Docker images (30 min)
cd /home/user/streamspace
make docker-build

# Expected output:
# Building controller image...
# ✓ Controller image built: streamspace/controller:latest
# Building API image...
# ✓ API image built: streamspace/api:latest
# Building UI image...
# ✓ UI image built: streamspace/ui:latest

# 8. Verify images (5 min)
docker images | grep streamspace
# Should show:
# streamspace/controller   latest   abc123   45MB
# streamspace/api          latest   def456   50MB
# streamspace/ui           latest   ghi789   25MB

# 9. Test local deployment (25 min)
docker-compose up -d
docker-compose ps  # Verify all running
docker-compose logs --tail=50  # Check for errors
docker-compose down
```

---

## Detailed Fix: UI Type Definitions

### Complete Updated `api.ts` (Replace UserQuota section)

```typescript
// User Management Types
export interface User {
  id: string;
  username: string;
  email: string;
  role: 'user' | 'admin';
  tier: 'free' | 'basic' | 'pro' | 'enterprise';
  enabled: boolean;
  createdAt: string;
  lastLoginAt?: string;
}

export interface UserQuota {
  userId: string;
  username: string;
  limits: {
    maxSessions: number;
    maxMemory: string;
    maxCPU: string;
    maxStorage: string;
  };
  used: {
    sessions: number;
    memory: string;
    cpu: string;
    storage: string;
  };
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  role?: 'user' | 'admin';
  tier?: 'free' | 'basic' | 'pro' | 'enterprise';
}

export interface UpdateUserRequest {
  email?: string;
  role?: 'user' | 'admin';
  tier?: 'free' | 'basic' | 'pro' | 'enterprise';
  enabled?: boolean;
}

export interface SetQuotaRequest {
  userId: string;  // Important: userId, not username
  limits: {
    maxSessions: number;
    maxMemory: string;
    maxCPU: string;
    maxStorage: string;
  };
}

// Group Management Types
export interface Group {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  memberCount: number;
}

export interface GroupMember {
  userId: string;
  username: string;
  role: 'member' | 'admin';
  joinedAt: string;
}

export interface CreateGroupRequest {
  name: string;
  description: string;
}

export interface UpdateGroupRequest {
  name?: string;
  description?: string;
}
```

### Example Fix in Quotas.tsx

**Before** (causes errors):
```typescript
<TableCell>{quota.username}</TableCell>
<TableCell>{quota.limits.maxSessions}</TableCell>
<TableCell>{quota.used.sessions} / {quota.limits.maxSessions}</TableCell>
```

**After** (fixed):
```typescript
<TableCell>{quota.username}</TableCell>
<TableCell>{quota.limits?.maxSessions ?? 'N/A'}</TableCell>
<TableCell>
  {quota.used?.sessions ?? 0} / {quota.limits?.maxSessions ?? 'N/A'}
</TableCell>
```

---

## Success Criteria

### Must Pass (All Components Build)

- [ ] `cd ui && npm run build` → Success (no errors)
- [ ] `cd api && go build ./cmd/main.go` → Success (binary created)
- [ ] `cd controller && go build ./cmd/main.go` → Success (binary created)
- [ ] `make docker-build` → Success (3 images created)
- [ ] `docker-compose up` → Success (all services start)

### Should Pass (Tests)

- [ ] `cd controller && make test` → Success (all tests pass)
- [ ] `cd ui && npm test` → Success (when tests added later)
- [ ] `cd api && go test ./...` → Success (when tests added later)

### Could Pass (CI/CD)

- [ ] GitHub Actions workflows run without errors
- [ ] Images pushed to GHCR successfully
- [ ] Helm chart lints successfully

---

## Commit Message Template

```
fix: resolve build issues for API and UI

## Changes

### UI Fixes
- Update UserQuota interface in lib/api.ts
  - Add username, limits, and used fields
  - Match backend API response structure
- Fix type errors in admin/Quotas.tsx
  - Update property access patterns
  - Add optional chaining for safety
  - Remove unused Chip import
- Fix unused imports in admin/UserDetail.tsx

### API Fixes
- Run go mod tidy to resolve dependencies
- Add go.sum file with all dependency hashes
- Verify build succeeds

### Verification
- ✅ Controller builds (already working)
- ✅ API builds (after go mod tidy)
- ✅ UI builds (after type fixes)
- ✅ Docker images build successfully
- ✅ docker-compose starts all services

## Testing
- Controller tests: All pass (780 lines)
- API tests: TODO (tracked in #issue-number)
- UI tests: TODO (tracked in #issue-number)

## Next Steps
- Write API tests (Priority 1)
- Write UI tests (Priority 1)
- Add integration tests (Priority 2)

Closes #build-issues
```

---

## Troubleshooting

### If UI Still Won't Build

Check exact error messages:
```bash
cd /home/user/streamspace/ui
npm run build 2>&1 | tee build-errors.log
grep "error TS" build-errors.log
```

Common issues:
1. **Missing imports**: Add them to the file
2. **Type mismatch**: Update interface definition
3. **Null safety**: Add optional chaining (`?.`) or nullish coalescing (`??`)

### If API Still Won't Build

Check Go module issues:
```bash
cd /home/user/streamspace/api
go mod verify  # Verify checksums
go list -m all  # List all dependencies
go mod graph | grep github.com/gin-gonic/gin  # Check specific dep
```

Common issues:
1. **Network issues**: Ensure internet connectivity
2. **Proxy issues**: Set GOPROXY if behind corporate proxy
3. **Module cache**: Clear with `go clean -modcache`

### If Docker Builds Fail

Check Dockerfile issues:
```bash
cd /home/user/streamspace
docker build --no-cache -t streamspace/ui:latest ui/
docker build --no-cache -t streamspace/api:latest api/
docker build --no-cache -t streamspace/controller:latest controller/
```

Common issues:
1. **Build context**: Ensure Dockerfile is in component directory
2. **Base image**: Verify `FROM` image is accessible
3. **Dependencies**: Ensure go.sum/package-lock.json exist

---

## After Fixes Complete

### Update Documentation

1. **PROJECT_STATUS.md**:
   ```markdown
   ## Current Status: BUILDABLE (Nov 14, 2025)

   All components now build successfully:
   - ✅ Controller: Builds and tests pass
   - ✅ API: Builds (dependencies resolved)
   - ✅ UI: Builds (type errors fixed)
   - ✅ Docker: All images build successfully

   Next phase: Testing (Weeks 2-4)
   ```

2. **NEXT_STEPS.md**:
   ```markdown
   ## ✅ Phase 0: Build Fixes (COMPLETE)

   All components now build successfully.

   ## → Phase 1: Testing (CURRENT)

   Focus: Write comprehensive tests for API and UI
   - API unit tests (1-2 weeks)
   - UI component tests (1-2 weeks)
   - Integration tests (1 week)

   Target: >80% code coverage
   ```

3. **Add BUILD_SUCCESS.md**:
   Document the fixes applied and how to reproduce builds.

---

## Timeline

| Time | Task | Status |
|------|------|--------|
| 0:00 - 0:30 | Update UI type definitions | ⏳ Pending |
| 0:30 - 1:00 | Fix Quotas.tsx errors | ⏳ Pending |
| 1:00 - 1:30 | Fix UserDetail.tsx & test UI build | ⏳ Pending |
| 1:30 - 1:35 | Run `go mod tidy` for API | ⏳ Pending |
| 1:35 - 1:40 | Verify API build | ⏳ Pending |
| 1:40 - 2:00 | Run controller tests | ⏳ Pending |
| 2:00 - 2:30 | Build all Docker images | ⏳ Pending |
| 2:30 - 2:35 | Verify images created | ⏳ Pending |
| 2:35 - 3:00 | Test docker-compose deployment | ⏳ Pending |

**Total**: 3 hours

---

## Success!

Once all checks pass:

```bash
# Create git branch for fixes
git checkout -b fix/build-issues

# Add changes
git add api/go.sum api/go.mod
git add ui/src/lib/api.ts
git add ui/src/pages/admin/Quotas.tsx
git add ui/src/pages/admin/UserDetail.tsx

# Commit with detailed message
git commit -F- <<EOF
fix: resolve build issues for API and UI

[Copy commit message from template above]
EOF

# Push and create PR
git push -u origin fix/build-issues
```

Then move to **Phase 1: Testing** 🎉

---

**Created**: November 14, 2025
**Priority**: P0 (Blocker)
**Estimated Effort**: 3 hours
**Dependencies**: Network connectivity for `go mod tidy`
