run:
	@go run ./cmd/server

test:
	@go test ./... -count=1

integration:
	@echo "Integration tests placeholder. Run 'docker compose up -d' before integration tests."

ci-test:
	@go test ./... -count=1 -v


