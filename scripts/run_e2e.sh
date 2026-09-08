#!/bin/bash
set -e
echo "=== Layer 1: Backend API Tests ==="
cd backend && PYTHONPATH=$(pwd) .venv/bin/python3 -m pytest ../tests/e2e/test_backend_e2e.py -v
echo "=== Layer 2: Frontend Playwright Tests ==="
cd ../frontend && npx playwright test
echo "=== Layer 3: Tauri Backend Tests ==="
cd src-tauri && cargo test
echo "=== All E2E tests passed! ==="
