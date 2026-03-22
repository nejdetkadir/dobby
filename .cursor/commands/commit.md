# Git Commit

Analyze all staged and unstaged changes, then create a commit following this project's conventions:

## Commit Message Format

```
<type>(<scope>): <description>
```

**Types:** `feat` | `fix` | `refactor` | `chore` | `docs` | `test` | `perf`
**Scope:** Service or component name in lowercase: `order`, `payment`, `dealer`, `catalog`, `delivery`, `finance`, `identity`, `cms`, `notification`, `report`, `gateway`, `shared`, `infra`

## Rules

- Description should explain WHY, not WHAT
- If changes span multiple services, use `shared` as scope
- Never commit files from `Keys/`, `.env`, or credential files
- Stage only relevant files, not build artifacts

## Examples

```
feat(order): add gRPC client for catalog price validation
fix(dealer): correct PostGIS distance calculation in matching
refactor(shared): extract cache service into hybrid L1/L2 pattern
chore(infra): update docker-compose postgres version to 16
```
