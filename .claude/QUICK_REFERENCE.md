# Multi-Agent Orchestration - Quick Reference

## Setup (One Time)

```bash
cd /path/to/streamspace
mkdir -p .claude/multi-agent
cp /path/to/streamspace-multi-agent/* .claude/multi-agent/
git add .claude/ && git commit -m "Add multi-agent setup"
```

## Starting Agents (Every Session)

Open 4 terminals, run `claude` in each, then paste:

**Terminal 1 (Architect):**
```
Act as Agent 1 (Architect) for StreamSpace.
Read: .claude/multi-agent/agent1-architect-instructions.md
Read: .claude/multi-agent/MULTI_AGENT_PLAN.md
CRITICAL: Documentation is aspirational. Audit actual code vs claims.
Begin comprehensive codebase audit.
```

**Terminal 2 (Builder):**
```
Act as Agent 2 (Builder) for StreamSpace.
Read: .claude/multi-agent/agent2-builder-instructions.md
Read: .claude/multi-agent/MULTI_AGENT_PLAN.md
Wait for assignments. Check plan every 30 min.
```

**Terminal 3 (Validator):**
```
Act as Agent 3 (Validator) for StreamSpace.
Read: .claude/multi-agent/agent3-validator-instructions.md
Read: .claude/multi-agent/MULTI_AGENT_PLAN.md
Monitor for testing assignments.
```

**Terminal 4 (Scribe):**
```
Act as Agent 4 (Scribe) for StreamSpace.
Read: .claude/multi-agent/agent4-scribe-instructions.md
Read: .claude/multi-agent/MULTI_AGENT_PLAN.md
Monitor for documentation requests.
```

## Common Commands

### Check Plan Status
```bash
cat .claude/multi-agent/MULTI_AGENT_PLAN.md | grep -A 3 "### Task:"
```

### View Recent Messages
```bash
tail -50 .claude/multi-agent/MULTI_AGENT_PLAN.md
```

### Check Agent Branches
```bash
git branch -a | grep agent
```

### View Agent Activity
```bash
git log --graph --all --oneline | head -20
```

## Task Status Format

```markdown
### Task: [Name]
- **Assigned To:** [Agent]
- **Status:** [Not Started | In Progress | Blocked | Review | Complete]
- **Priority:** [Low | Medium | High | Critical]
- **Dependencies:** [List or "None"]
- **Notes:** [Details]
- **Last Updated:** [Date] - [Agent]
```

## Message Format

```markdown
## [From] → [To] - [Time]
[Message content]
```

## Git Branch Strategy

- `agent1/planning` - Architect work
- `agent2/implementation` - Builder work
- `agent3/testing` - Validator work
- `agent4/documentation` - Scribe work
- `develop` - Integration branch

## Typical Workflow

1. **Architect** researches and creates tasks
2. **Architect** assigns to Builder/Validator/Scribe
3. **Builder** implements and notifies Validator
4. **Validator** tests and reports bugs
5. **Builder** fixes bugs
6. **Scribe** documents
7. **Architect** reviews and approves merge

## Emergency Commands

### Agent Lost Context
```
Re-read: .claude/multi-agent/agent[X]-[role]-instructions.md
Re-read: .claude/multi-agent/MULTI_AGENT_PLAN.md
```

### Check What Changed
```bash
git diff develop agent2/implementation
```

### Resolve Conflicts
```bash
# Coordinate through Architect
# Use separate files when possible
git status
```

## Key Files

- `.claude/multi-agent/MULTI_AGENT_PLAN.md` - **THE SOURCE OF TRUTH**
- `.claude/multi-agent/agent*-instructions.md` - Role definitions
- `.claude/multi-agent/SETUP_GUIDE.md` - Detailed instructions

## Remember

✅ Check plan every 30 minutes
✅ Update status after completing tasks
✅ Leave clear messages for other agents
✅ Use descriptive commit messages
✅ Let Architect coordinate merges

## Current Priority: Implementation Gap Analysis

**Reality:** Documentation describes ambitious vision, but many features aren't actually implemented yet.

**First Mission:** 
1. Audit codebase vs documentation
2. Identify what actually works
3. Create honest feature matrix
4. Prioritize core functionality
5. Build working foundation before adding enterprise features

**Success Criteria:**
- Honest documentation
- Working core features (sessions, templates, basic auth)
- Clear roadmap based on reality
- Solid foundation to build on

## Need Help?

1. Check MULTI_AGENT_PLAN.md for agent messages
2. Read SETUP_GUIDE.md
3. Review agent instruction files
4. Ask in StreamSpace Discord
5. Reference blog post: https://sjramblings.io/multi-agent-orchestration-claude-code-when-ai-teams-beat-solo-acts/
