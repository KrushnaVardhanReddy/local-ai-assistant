#!/bin/bash

set -e

# Navigate to the project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Starting local Supabase instance..."
cd "$PROJECT_ROOT/supabase"

# Run supabase start and capture output
# Use npx supabase to ensure we use the local version if available, or fetch it
SUPABASE_OUTPUT=$(npx supabase start)

echo "$SUPABASE_OUTPUT"

# Extract API URL and service_role key
# The output format is typically:
#          API URL: http://127.0.0.1:54321
# service_role key: eyJhb...

SUPABASE_URL=$(echo "$SUPABASE_OUTPUT" | grep -i "API URL:" | awk '{print $NF}' | tr -d '\r')
SUPABASE_SERVICE_ROLE_KEY=$(echo "$SUPABASE_OUTPUT" | grep -i "service_role key:" | awk '{print $NF}' | tr -d '\r')

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
