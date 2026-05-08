#!/usr/bin/env bash
set -euo pipefail

BREW_PREFIX=$(brew --prefix 2>/dev/null || echo /usr/local)
P=$(pwd)

mkdir -p data/prometheus data/grafana data/loki/{chunks,rules} data/app data/alloy data/tempo

prometheus \
    --config.file=configs/prometheus.yml \
    --storage.tsdb.path=./data/prometheus \
    --web.listen-address=:9090 \
    --log.level=warn &
PID_PROM=$!

loki \
    -config.file=configs/loki.yml \
    >> data/loki/server.log 2>&1 &
PID_LOKI=$!

docker rm -f running-tracker-tempo >/dev/null 2>&1 || true
docker run --rm --name running-tracker-tempo \
    -p 3200:3200 -p 4317:4317 \
    -v "$P/configs/tempo.yml:/etc/tempo/tempo.yml:ro" \
    grafana/tempo:2.10.5 -config.file=/etc/tempo/tempo.yml \
    >> data/tempo/server.log 2>&1 &
PID_TEMPO=$!

alloy run \
    --storage.path=./data/alloy \
    --server.http.listen-addr=127.0.0.1:12345 \
    configs/alloy.alloy \
    >> data/alloy/alloy.log 2>&1 &
PID_ALLOY=$!

GF_PATHS_PROVISIONING="$P/configs/grafana/provisioning" \
GF_PATHS_DATA="$P/data/grafana" \
GF_SERVER_HTTP_PORT=3000 \
GF_AUTH_ANONYMOUS_ENABLED=true \
GF_AUTH_ANONYMOUS_ORG_ROLE=Admin \
GF_SECURITY_ADMIN_PASSWORD=admin \
GF_DASHBOARD_PATH="$P/configs/grafana/dashboards" \
grafana server --homepath "$BREW_PREFIX/share/grafana" >> data/grafana/server.log 2>&1 &
PID_GRAF=$!

LOG_FILE=./data/app/app.log go run ./cmd/server &
PID_APP=$!

wait "$PID_APP"
