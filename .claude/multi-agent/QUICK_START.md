# Multi-Agent Quick Start Guide

**For User:** How to work with the 4-agent team for v2.0-beta development

---

## 🎯 Overview

You have 4 independent agent workspaces working in parallel:

```
streamspace/           → Agent 1 (Architect) - Coordination
streamspace-builder/   → Agent 2 (Builder) - Bug fixes
streamspace-validator/ → Agent 3 (Validator) - Testing
streamspace-scribe/    → Agent 4 (Scribe) - Documentation
```

---

## 🚀 Starting Work

### Step 1: Open Claude Code in Each Workspace

**Terminal 1 - Validator (START FIRST):**
```bash
cd /Users/s0v3r1gn/streamspace/streamspace-validator
# Open Claude Code here
```

**Paste this prompt:**
```
Act as Agent 3 (Validator) for StreamSpace v2.0-beta integration testing.

Read: .claude/multi-agent/agent3-validator-instructions.md
Read: .claude/multi-agent/MULTI_AGENT_PLAN.md (lines 91-132)

Current Task: Integration Testing & E2E Validation (P0 CRITICAL)

Your Mission:
1. Deploy v2.0-beta to local K8s cluster (use scripts/local-build.sh + local-deploy.sh)
2. Test 8 critical scenarios (see MULTI_AGENT_PLAN.md)
3. Create comprehensive test report
4. Report all bugs with P0/P1/P2 priority

Push work to: claude/v2-validator
```

---

**Terminal 2 - Scribe (START IN PARALLEL):**
```bash
cd /Users/s0v3r1gn/streamspace/streamspace-scribe
# Open Claude Code here
```

**Paste this prompt:**
```
Act as Agent 4 (Scribe) for StreamSpace v2.0-beta documentation.

Read: .claude/multi-agent/agent4-scribe-instructions.md
Read: .claude/multi-agent/MULTI_AGENT_PLAN.md (lines 136-169)

Current Task: v2.0-beta Documentation (P0 CRITICAL)

Your Mission:
Create 6 documentation files:
1. docs/V2_DEPLOYMENT_GUIDE.md
2. docs/V2_AGENT_GUIDE.md
3. docs/V2_ARCHITECTURE.md
4. docs/V2_MIGRATION_GUIDE.md
5. CHANGELOG.md (update)
6. README.md (update)

Push work to: claude/v2-scribe
```

---

**Terminal 3 - Builder (STANDBY):**
```bash
cd /Users/s0v3r1gn/streamspace/streamspace-builder
# Open Claude Code here
```

**Paste this prompt:**
```
Act as Agent 2 (Builder) for StreamSpace v2.0-beta.

Read: .claude/multi-agent/agent2-builder-instructions.md
Read: .claude/multi-agent/MULTI_AGENT_PLAN.md (lines 173-195)

Current Status: STANDBY - Waiting for bug reports from Validator

Your Role:
- Monitor for bug reports
- Fix bugs discovered during testing
- Push fixes to claude/v2-builder

Stand by until Validator reports bugs.
```

---

### Step 2: Let Agents Work

Each agent will:
1. Read their instructions
2. Execute their assigned tasks
3. Commit work to their branch
4. Push to remote

**You don't need to do anything** - just let them work independently!

---

## 🔄 Integration (When Agents Complete Work)

### When to Integrate

Integrate when you see:
- ✅ Validator pushes test report
- ✅ Scribe pushes documentation
- ✅ Builder pushes bug fixes

### How to Integrate (As Architect)

**Terminal 4 - Architect:**
```bash
cd /Users/s0v3r1gn/streamspace/streamspace
# Open Claude Code here
```

**Paste this prompt:**
```
Act as Agent 1 (Architect) for StreamSpace v2.0 coordination.

Pull updates from all agent branches and integrate:

git fetch origin claude/v2-builder claude/v2-validator claude/v2-scribe
git merge origin/claude/v2-scribe --no-edit
git merge origin/claude/v2-builder --no-edit
git merge origin/claude/v2-validator --no-edit

Update MULTI_AGENT_PLAN.md with integration summary, then push to feature/streamspace-v2-agent-refactor.
```

---

## 📊 Monitoring Progress

### Check What Agents Have Done

**See latest commits from each agent:**
```bash
cd /Users/s0v3r1gn/streamspace/streamspace
git fetch --all
git log --oneline origin/claude/v2-validator -5
git log --oneline origin/claude/v2-scribe -5
git log --oneline origin/claude/v2-builder -5
```

### Check Agent Workspace Status

**Visit each workspace:**
```bash
# Validator
cd /Users/s0v3r1gn/streamspace/streamspace-validator && git status

# Scribe
cd /Users/s0v3r1gn/streamspace/streamspace-scribe && git status

# Builder
cd /Users/s0v3r1gn/streamspace/streamspace-builder && git status
```

---

## 🎯 Expected Timeline

### Week 1 (Days 1-7)

**Validator:**
- Days 1-2: Setup testing environment
- Days 3-5: Execute 8 test scenarios
- Days 6-7: Write test report, log bugs

**Scribe:**
- Days 1-2: V2_DEPLOYMENT_GUIDE.md
- Days 3-4: V2_AGENT_GUIDE.md + V2_ARCHITECTURE.md
- Days 5-6: V2_MIGRATION_GUIDE.md
- Day 7: CHANGELOG.md + README.md updates

**Builder:**
- Days 1-7: Standby, fix bugs as reported

### Week 2 (Days 8-14)

**Validator:**
- Days 8-10: Retest after bug fixes
- Days 11-12: Performance testing
- Days 13-14: Final validation

**Scribe:**
- Days 8-10: Polish documentation
- Days 11-14: Review and updates based on feedback

**Builder:**
- Days 8-14: Final bug fixes and refinements

**Architect (You):**
- Days 10-14: Integration, release preparation

---

## ✅ Success Checklist

### Validator Complete
- [ ] Test report published
- [ ] All 8 scenarios tested
- [ ] Bug list created (P0/P1/P2)
- [ ] Performance metrics recorded
- [ ] Integration test suite created
- [ ] Pushed to `claude/v2-validator`

### Scribe Complete
- [ ] V2_DEPLOYMENT_GUIDE.md created
- [ ] V2_AGENT_GUIDE.md created
- [ ] V2_ARCHITECTURE.md created
- [ ] V2_MIGRATION_GUIDE.md created
- [ ] CHANGELOG.md updated
- [ ] README.md updated
- [ ] Pushed to `claude/v2-scribe`

### Builder Complete
- [ ] All P0 bugs fixed
- [ ] All P1 bugs fixed or documented
- [ ] Tests pass after fixes
- [ ] Pushed to `claude/v2-builder`

### Architect Complete
- [ ] All agent work integrated
- [ ] MULTI_AGENT_PLAN.md updated
- [ ] Release notes prepared
- [ ] v2.0-beta ready for deployment

---

## 🚨 Troubleshooting

### Agent Seems Stuck
**Solution:** Give more specific guidance
```
You seem stuck. Focus on [specific task].
Read [specific file/section] for guidance.
Break down into smaller steps if needed.
```

### Merge Conflicts
**Solution:** Architect resolves manually
```bash
# In architect workspace
git merge origin/claude/v2-[agent]
# Fix conflicts manually
git add .
git commit -m "merge: resolve conflicts from [agent]"
```

### Agent Asks Questions
**Answer directly or:**
```
Check [relevant documentation]
See [example in codebase]
Proceed with [your decision]
```

---

## 💡 Pro Tips

1. **Let agents work independently** - Don't micromanage
2. **Integrate frequently** - Don't wait until everything is done
3. **Clear deliverables** - Each agent knows what to produce
4. **Parallel work** - Validator and Scribe work simultaneously
5. **Trust the process** - Agents are designed to work autonomously

---

## 📞 Quick Reference

**Workspaces:**
- Architect: `/Users/s0v3r1gn/streamspace/streamspace`
- Builder: `/Users/s0v3r1gn/streamspace/streamspace-builder`
- Validator: `/Users/s0v3r1gn/streamspace/streamspace-validator`
- Scribe: `/Users/s0v3r1gn/streamspace/streamspace-scribe`

**Branches:**
- Integration: `feature/streamspace-v2-agent-refactor`
- Architect: `claude/v2-architect`
- Builder: `claude/v2-builder`
- Validator: `claude/v2-validator`
- Scribe: `claude/v2-scribe`

**Key Files:**
- Coordination Plan: `.claude/multi-agent/MULTI_AGENT_PLAN.md`
- Status Updates: `.claude/multi-agent/COORDINATION_STATUS.md`
- This Guide: `.claude/multi-agent/QUICK_START.md`

---

**Ready to start?** Open 3 terminals, paste the prompts, and let the agents work! 🚀
