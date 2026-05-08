run:
	go run ./cmd/server

build:
	go build -o bin/running-tracker ./cmd/server

run-all:
	@bash scripts/run-all.sh

stop:
	@pkill -f "prometheus --config.file" 2>/dev/null || true
	@pkill -f "grafana server"           2>/dev/null || true
	@pkill -f "loki -config.file"        2>/dev/null || true
	@pkill -f "tempo -config.file"       2>/dev/null || true
	@pkill -f "alloy run"                2>/dev/null || true
	@lsof -ti:8080,3200,4317,14317 | xargs kill 2>/dev/null || true

install-tools:
	brew tap grafana/grafana
	brew install prometheus grafana loki grafana-alloy grafana/grafana/tempo
