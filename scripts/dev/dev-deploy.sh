#!/bin/bash
# Idempotent local dev deploy script for kind + Helm, with auto kind cluster detection/creation
# NOTE: If you restart the db deployment, you MUST also trigger the migration job because data is not persistent.
# Use the restart_db_and_migrate helper in this script for that purpose.
set -euo pipefail

# Config
CHART_DIR="charts/inbox-whisperer"
NAMESPACE="default"
BACKEND_IMAGE="backend:latest"
FRONTEND_IMAGE="frontend:latest"
MIGRATE_IMAGE="inbox-whisperer-migrate:latest"
KIND_CLUSTER=""
KIND_DEFAULT_NAME="inbox-whisperer"

# Helper: Restart DB and trigger migration job
restart_db_and_migrate() {
  banner "[STEP X] Restarting DB deployment"
  kubectl rollout restart deployment/db || true
  banner "[STEP X] Waiting for DB to become ready"
  kubectl wait --for=condition=available --timeout=60s deployment/db || true
  banner "[STEP X] Triggering migration job"
  kubectl delete job migration-job --ignore-not-found
  # Helm should recreate the migration job if configured as a Helm hook, otherwise manually create it here if needed.
  banner "[DONE] DB restarted and migration job triggered"
}

# Banner helper
banner() {
  echo ""
  echo "==================================================="
  echo "$1"
  echo "==================================================="
  echo ""
}

# Parse arguments
BUILD_ALL=false
BUILD_BACKEND=false
BUILD_FRONTEND=false
BUILD_MIGRATE=false
DEPLOY_ALL=false
CREATE_SECRETS=false

for arg in "$@"; do
  case $arg in
    --build-all) BUILD_ALL=true ;;
    --build-backend) BUILD_BACKEND=true ;;
    --build-frontend) BUILD_FRONTEND=true ;;
    --build-migrate) BUILD_MIGRATE=true ;;
    --deploy-all) DEPLOY_ALL=true ;;
    --create-secrets) CREATE_SECRETS=true ;;
  esac
  shift || true
done

# Detect or create kind cluster
banner "[STEP 1] Detect or create kind cluster"
CLUSTERS=$(kind get clusters)
if [ -z "$CLUSTERS" ]; then
  echo "[INFO] No kind clusters found. Creating cluster named '$KIND_DEFAULT_NAME'..."
  kind create cluster --name "$KIND_DEFAULT_NAME" || { echo "[ERROR] Failed to create kind cluster"; exit 1; }
  KIND_CLUSTER="$KIND_DEFAULT_NAME"
else
  # Use the first cluster found
  KIND_CLUSTER=$(echo "$CLUSTERS" | head -n1)
  echo "[INFO] Using kind cluster: $KIND_CLUSTER"
fi

# Helper: Build and load backend
build_backend() {
  banner "[STEP 2] Build backend image"
  DOCKER_BUILDKIT=1 docker build -t $BACKEND_IMAGE -f cmd/backend/Dockerfile . || { echo "[ERROR] Backend image build failed"; exit 1; }
  kind load docker-image $BACKEND_IMAGE --name $KIND_CLUSTER || { echo "[ERROR] Failed to load backend image into kind"; exit 1; }
  kubectl rollout restart deployment/backend || true
  banner "[DONE] Backend built, loaded, and deployment restarted"
}

# Helper: Build and load frontend
build_frontend() {
  banner "[STEP 3] Build frontend image"
  DOCKER_BUILDKIT=1 docker build -t $FRONTEND_IMAGE -f web/Dockerfile ./web || { echo "[ERROR] Frontend image build failed"; exit 1; }
  kind load docker-image $FRONTEND_IMAGE --name $KIND_CLUSTER || { echo "[ERROR] Failed to load frontend image into kind"; exit 1; }
  kubectl rollout restart deployment/frontend || true
  banner "[DONE] Frontend built, loaded, and deployment restarted"
}

# Helper: Build and load migration image
build_migrate() {
  banner "[STEP 4] Build migration image"
  DOCKER_BUILDKIT=1 docker build -t $MIGRATE_IMAGE -f migrations/image/Dockerfile migrations/image || { echo "[ERROR] Migration image build failed"; exit 1; }
  kind load docker-image $MIGRATE_IMAGE --name $KIND_CLUSTER || { echo "[ERROR] Failed to load migration image into kind"; exit 1; }
  banner "[DONE] Migration image built and loaded"
}

# Helper: Create secrets
create_secrets() {
  banner "[STEP 6] Ensure db-secret exists for local dev (idempotent)"
  kubectl create secret generic db-secret \
    --namespace $NAMESPACE \
    --from-literal=postgresUser="${POSTGRES_USER:-inboxwhisperer}" \
    --from-literal=postgresPassword="${POSTGRES_PASSWORD:-changeme}" \
    --from-literal=postgresDb="${POSTGRES_DB:-inboxwhisperer}" \
    --dry-run=client -o yaml | kubectl apply -f - || { echo "[ERROR] Failed to create/apply db-secret"; exit 1; }
  banner "[DONE] DB secret created/applied"

  banner "[STEP 7] Generate backend config.json from environment or defaults"
  BACKEND_CONFIG_TEMP=$(mktemp)
  cat > $BACKEND_CONFIG_TEMP <<EOF
{
  "google": {
    "client_id": "${GOOGLE_CLIENT_ID:-}",
    "client_secret": "${GOOGLE_CLIENT_SECRET:-}",
    "redirect_url": "http://localhost:3000/api/auth/callback"
  },
  "server": {
    "port": "8080",
    "db_url": "postgres://${POSTGRES_USER:-inboxwhisperer}:${POSTGRES_PASSWORD:-changeme}@${POSTGRES_HOST:-db}:5432/${POSTGRES_DB:-inboxwhisperer}?sslmode=disable",
    "frontend_url": "http://localhost:3000",
    "sessionCookieSecure": false
  }
}
EOF
  echo "[STEP 8] Create backend-config secret from config.json"
  kubectl create secret generic backend-config --namespace $NAMESPACE --from-file=config.json=$BACKEND_CONFIG_TEMP --dry-run=client -o yaml | kubectl apply -f - || { echo "[ERROR] Failed to create/apply backend-config secret";  }
  rm $BACKEND_CONFIG_TEMP
  banner "[DONE] Backend config secret created/applied"
}

# Helper: Helm deploy
helm_deploy() {
  banner "[STEP 9] Helm upgrade/install (timeout 60s)"
  helm upgrade --install inbox-whisperer $CHART_DIR \
    --namespace $NAMESPACE --create-namespace \
    --set backend.image=$BACKEND_IMAGE \
    --set frontend.image=$FRONTEND_IMAGE \
    --wait --atomic --timeout 60s || { echo "[ERROR] Helm upgrade/install failed"; exit 1; }
  banner "[DONE] Helm upgrade/install complete"
}

# Helper: Show status
show_status() {
  banner "[STEP 11] Show deployment status"
  kubectl get deployments -n $NAMESPACE || { echo "[ERROR] Failed to get deployments"; exit 1; }
  kubectl get pods -n $NAMESPACE || { echo "[ERROR] Failed to get pods"; exit 1; }
}

# Load environment variables if dev.env exists
if [ -f dev.env ]; then
  echo "[INFO] Loading environment variables from dev.env"
  set -o allexport
  source dev.env
  set +o allexport
else
  echo "[INFO] dev.env file not found, skipping env load"
fi

# Main logic
if $DEPLOY_ALL; then
  build_backend
  build_frontend
  build_migrate
  create_secrets
  helm_deploy
  show_status
  exit 0
fi

if $BUILD_ALL; then
  build_backend
  build_frontend
  build_migrate
  exit 0
fi

if $BUILD_BACKEND; then
  build_backend
  exit 0
fi

if $BUILD_FRONTEND; then
  build_frontend
  exit 0
fi

if $BUILD_MIGRATE; then
  build_migrate
  restart_db_and_migrate
  helm_deploy
  exit 0
fi

if $CREATE_SECRETS; then
  create_secrets
  exit 0
fi

# Default: full deploy
build_backend
build_frontend
build_migrate
create_secrets
helm_deploy
show_status
exit 0
