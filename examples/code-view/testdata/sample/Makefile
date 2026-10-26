.PHONY: run test fmt lint

run:
	cd backend && go run . -config ../config/services.yaml

test:
	cd backend && go test ./...

fmt:
	gofmt -w backend
	python3 -m black tools

lint:
	cd backend && go vet ./...
	python3 -m ruff check tools
