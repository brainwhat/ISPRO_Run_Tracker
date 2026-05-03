#!/usr/bin/env bash
set -euo pipefail

BREW_PREFIX=$(brew --prefix 2>/dev/null || echo /usr/local)
P=$(pwd)

mkdir -p data/prometheus data/grafana

prometheus \
    --config.file=configs/prometheus.yml \
    --storage.tsdb.path=./data/prometheus \
    --web.listen-address=:9090 \
    --log.level=warn &
PID_PROM=$!

GF_PATHS_PROVISIONING="$P/configs/grafana/provisioning" \
GF_PATHS_DATA="$P/data/grafana" \
GF_SERVER_HTTP_PORT=3000 \
GF_AUTH_ANONYMOUS_ENABLED=true \
GF_AUTH_ANONYMOUS_ORG_ROLE=Admin \
GF_SECURITY_ADMIN_PASSWORD=admin \
GF_DASHBOARD_PATH="$P/configs/grafana/dashboards" \
grafana server --homepath "$BREW_PREFIX/share/grafana" >> data/grafana/server.log 2>&1 &
PID_GRAF=$!

go run ./cmd/server &
PID_APP=$!

wait "$PID_APP"
