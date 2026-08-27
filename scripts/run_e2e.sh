#!/bin/bash
set -e
echo "=== Layer 1: Backend API Tests ==="
cd backend && PYTHONPATH=$(pwd) python3 -m pytest ../tests/ -v
echo "=== Layer 2: Frontend Playwright Tests ==="
cd ../frontend && npx playwright test
echo "=== All E2E tests passed! ==="
