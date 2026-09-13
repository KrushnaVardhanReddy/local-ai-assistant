#!/bin/bash

set -e

# Navigate to the project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# If podman is installed, set the DOCKER_HOST to the user podman socket
if command -v podman &> /dev/null; then
    export DOCKER_HOST="unix://$XDG_RUNTIME_DIR/podman/podman.sock"
fi

echo "Starting local Supabase instance..."
cd "$PROJECT_ROOT/supabase"

# Run supabase start (we don't need to parse this output directly)
npx supabase start

echo "Fetching status..."
STATUS_ENV=$(npx supabase status -o env)

SUPABASE_URL=$(echo "$STATUS_ENV" | grep "^API_URL=" | cut -d'"' -f2)
SUPABASE_SERVICE_ROLE_KEY=$(echo "$STATUS_ENV" | grep "^SERVICE_ROLE_KEY=" | cut -d'"' -f2)

if [ -z "$SUPABASE_URL" ] || [ -z "$SUPABASE_SERVICE_ROLE_KEY" ]; then
    echo "Error: Failed to extract Supabase URL or Service Role Key from the output."
    exit 1
fi

echo "Successfully extracted Supabase configuration."
echo "SUPABASE_URL: $SUPABASE_URL"

# Create or overwrite cloud-worker/.dev.vars
DEV_VARS_FILE="$PROJECT_ROOT/cloud-worker/.dev.vars"
echo "Writing configuration to $DEV_VARS_FILE..."

cat <<EOF > "$DEV_VARS_FILE"
SUPABASE_URL=$SUPABASE_URL
SUPABASE_SERVICE_ROLE_KEY=$SUPABASE_SERVICE_ROLE_KEY
STRIPE_SECRET_KEY=sk_test_123
STRIPE_WEBHOOK_SECRET=whsec_123
EOF

echo "Done writing .dev.vars."

echo "Starting cloud-worker..."
cd "$PROJECT_ROOT/cloud-worker"
npm run dev
