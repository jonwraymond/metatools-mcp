#!/bin/bash
# PRD-180 Migration Validation Script
# Validates that metatools-mcp has migrated from 7 standalone repos to 3 consolidated repos

set -e

echo "=== PRD-180 Migration Validation ==="
echo ""

echo "1. Checking compilation..."
GOWORK=off go build ./...
echo "   ✓ All packages compile"
echo ""

echo "2. Running unit tests..."
GOWORK=off go test ./... -short > /dev/null
echo "   ✓ All tests pass"
echo ""

echo "3. Running with toolsearch tag..."
GOWORK=off go test -tags=toolsearch ./internal/bootstrap/... > /dev/null
echo "   ✓ toolsearch tests pass"
echo ""

echo "4. Running with toolruntime tag..."
GOWORK=off go test -tags=toolruntime ./cmd/metatools/... > /dev/null
echo "   ✓ toolruntime tests pass"
echo ""

echo "5. Checking for old imports..."
OLD_IMPORTS=$(grep -rE "github\.com/jonwraymond/tool(model|index|search|docs|run|runtime|code)[^/]" --include="*.go" . 2>/dev/null | grep -v "go.sum" | grep -v "go.mod" || true)
if [ -n "$OLD_IMPORTS" ]; then
    echo "   ERROR: Found old imports:"
    echo "$OLD_IMPORTS"
    exit 1
fi
echo "   ✓ No old imports found"
echo ""

echo "6. Verifying go.mod..."
ERRORS=""
if grep -q "jonwraymond/toolmodel[^/]" go.mod 2>/dev/null; then ERRORS="$ERRORS toolmodel"; fi
if grep -q "jonwraymond/toolindex[^/]" go.mod 2>/dev/null; then ERRORS="$ERRORS toolindex"; fi
if grep -q "jonwraymond/toolsearch[^/]" go.mod 2>/dev/null; then ERRORS="$ERRORS toolsearch"; fi
if grep -q "jonwraymond/tooldocs[^/]" go.mod 2>/dev/null; then ERRORS="$ERRORS tooldocs"; fi
if grep -q "jonwraymond/toolrun[^/]" go.mod 2>/dev/null; then ERRORS="$ERRORS toolrun"; fi
if grep -q "jonwraymond/toolruntime[^/]" go.mod 2>/dev/null; then ERRORS="$ERRORS toolruntime"; fi
if grep -q "jonwraymond/toolcode[^/]" go.mod 2>/dev/null; then ERRORS="$ERRORS toolcode"; fi

if [ -n "$ERRORS" ]; then
    echo "   ERROR: Old dependencies in go.mod:$ERRORS"
    exit 1
fi
echo "   ✓ Old dependencies removed"

# Verify new dependencies present
if ! grep -q "jonwraymond/toolfoundation" go.mod; then echo "   ERROR: toolfoundation missing"; exit 1; fi
if ! grep -q "jonwraymond/tooldiscovery" go.mod; then echo "   ERROR: tooldiscovery missing"; exit 1; fi
if ! grep -q "jonwraymond/toolexec" go.mod; then echo "   ERROR: toolexec missing"; exit 1; fi
echo "   ✓ New consolidated dependencies present"
echo ""

echo "=== Validation Complete - PRD-180 SUCCESS ==="
echo ""
echo "Summary:"
echo "  - 7 old dependencies removed (toolmodel, toolindex, toolsearch, tooldocs, toolrun, toolruntime, toolcode)"
echo "  - 3 new consolidated dependencies added (toolfoundation, tooldiscovery, toolexec)"
echo "  - All tests passing"
echo "  - All builds successful"
