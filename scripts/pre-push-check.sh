#!/bin/bash
# scripts/pre-push-check.sh
# 
# Security verification script to catch secrets before pushing to GitHub
# Install as git hook: cp scripts/pre-push-check.sh .git/hooks/pre-push && chmod +x .git/hooks/pre-push
# 
# Usage:
#   ./scripts/pre-push-check.sh           # Check staged changes (pre-commit)
#   GIT_INDEX_FILE="" ./scripts/pre-push-check.sh  # Check working directory

set -u

RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

FAILED=0

echo "🔐 Running pre-push security checks..."
echo

# ============================================
# Check if we have staged changes (or are in a git repo)
# ============================================
if ! git rev-parse --git-dir > /dev/null 2>&1; then
  echo -e "${YELLOW}⚠️  Not in a git repository${NC}"
  echo "Skipping staged changes check"
  echo
else
  STAGED_FILES=$(git diff --cached --name-only 2>/dev/null || echo "")
  
  if [ -z "$STAGED_FILES" ]; then
    echo -e "${YELLOW}⚠️  No staged changes detected${NC}"
    echo "Running checks on current working directory..."
    CHECK_MODE="working"
  else
    echo "Found $(echo "$STAGED_FILES" | wc -l) staged files"
    CHECK_MODE="staged"
  fi
fi

# ============================================
# 1. Check for .env files in repo
# ============================================
echo "📋 Checking for .env files..."

if git ls-files 2>/dev/null | grep -E "\.env($|\.)" > /dev/null; then
  echo -e "${RED}✗ Found .env files tracked in git${NC}"
  git ls-files | grep -E "\.env"
  FAILED=1
else
  echo -e "${GREEN}✓ No .env files tracked in git${NC}"
fi

echo

# ============================================
# 2. Check for secrets/ directory
# ============================================
echo "📋 Checking for secrets/ directory..."

if git ls-files 2>/dev/null | grep -E "^secrets/" > /dev/null; then
  echo -e "${RED}✗ Found files in secrets/ directory tracked in git${NC}"
  git ls-files | grep -E "^secrets/"
  FAILED=1
else
  echo -e "${GREEN}✓ No files from secrets/ in tracked files${NC}"
fi

echo

# ============================================
# 3. Check .gitignore has critical patterns
# ============================================
echo "📋 Verifying .gitignore patterns..."

if [ -f .gitignore ]; then
  CRITICAL_PATTERNS=(".env" "secrets/" "gcp_credentials.json")
  for pattern in "${CRITICAL_PATTERNS[@]}"; do
    if grep -q "^${pattern}" .gitignore; then
      echo -e "${GREEN}✓ Pattern found: $pattern${NC}"
    else
      echo -e "${YELLOW}⚠️  Pattern missing: $pattern${NC}"
    fi
  done
else
  echo -e "${RED}✗ .gitignore not found${NC}"
  FAILED=1
fi

echo

# ============================================
# 4. Check for common credential patterns in code
# ============================================
echo "📋 Checking for hardcoded credential patterns..."

FORBIDDEN_PATTERNS=(
  "TIGER_DB_PASSWORD="
  "TIGER_SECRET_KEY="
  "password.*=.*[a-zA-Z0-9]"
  'api.?key.*=.*[a-zA-Z0-9]'
  "Bearer[[:space:]]+[A-Za-z0-9]"
)

EXCLUDE_PATHS=(
  ".gitignore"
  "SECURITY.md"
  "SETUP.md"
  ".env.example"
  "scripts/pre-push-check.sh"
  "HARDENING-CHECKLIST.md"
)

# Build exclude pattern
EXCLUDE_PATTERN=""
for path in "${EXCLUDE_PATHS[@]}"; do
  if [ -z "$EXCLUDE_PATTERN" ]; then
    EXCLUDE_PATTERN="--exclude=$path"
  else
    EXCLUDE_PATTERN="$EXCLUDE_PATTERN --exclude=$path"
  fi
done

FOUND_ISSUES=0
for pattern in "${FORBIDDEN_PATTERNS[@]}"; do
  if grep -r "$pattern" backend/ frontend/ --include="*.go" --include="*.ts" --include="*.tsx" $EXCLUDE_PATTERN 2>/dev/null | \
     grep -v "example\|template\|your_.*_here\|change_me" > /dev/null; then
    echo -e "${RED}✗ Possible credential match: $pattern${NC}"
    grep -r "$pattern" backend/ frontend/ --include="*.go" --include="*.ts" --include="*.tsx" $EXCLUDE_PATTERN 2>/dev/null | \
      grep -v "example\|template\|your_.*_here\|change_me" | head -3
    FOUND_ISSUES=1
  fi
done

if [ $FOUND_ISSUES -eq 0 ]; then
  echo -e "${GREEN}✓ No obvious hardcoded secrets found${NC}"
else
  FAILED=1
fi

echo

# ============================================
# 5. Summary
# ============================================

if [ $FAILED -eq 0 ]; then
  echo -e "${GREEN}✅ All security checks passed!${NC}"
  echo "Safe to push 🚀"
  exit 0
else
  echo -e "${RED}❌ SECURITY CHECKS FAILED${NC}"
  echo
  echo "Before pushing, ensure:"
  echo "  1. No .env files tracked"
  echo "  2. No files from secrets/ tracked"
  echo "  3. No real credentials in code"
  echo "  4. All passwords/keys are in .env.example (without real values)"
  echo
  echo "If you need to untrack files:"
  echo "  git rm --cached <file>"
  echo "  git commit -m 'Remove sensitive file from tracking'"
  exit 1
fi
