.PHONY: install dev-backend dev-frontend dev-all stop e2e clean

# Install all dependencies (Frontend + Backend)
install:
	@echo "Installing backend dependencies..."
	cd backend && python3 -m venv .venv && .venv/bin/pip install -r requirements.txt
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Install complete!"

# Run the FastAPI backend
dev-backend:
	@echo "Starting backend..."
	cd backend && .venv/bin/python3 app.py

# Run the Svelte/Tauri frontend
dev-frontend:
	@echo "Starting frontend..."
	cd frontend && npm run tauri dev

# Run both backend and frontend concurrently
dev-all:
	@echo "Starting backend and frontend... (Press CTRL+C to stop both)"
	@bash -c 'trap "echo \"Cleaning up...\"; pkill -f \"python3 app.py\" || true" EXIT; \
	cd backend && .venv/bin/python3 app.py & \
	echo "Waiting 5 seconds for backend to initialize..." && \
	sleep 5 && \
	echo "Starting frontend..." && \
	cd frontend && npm run tauri dev'

# Force kill all running instances of the frontend and backend
stop:
	@echo "Stopping all backend and frontend processes..."
	-pkill -f "python3 app.py"
	-pkill -f "tauri"
	@echo "All processes stopped!"

# Run the E2E tests (from P7-T6)
e2e:
	@echo "Running E2E tests..."
	bash scripts/run_e2e.sh

# Clean up caches and node_modules
clean:
	rm -rf backend/.venv backend/__pycache__
	rm -rf frontend/node_modules frontend/dist
	@echo "Clean complete!"
