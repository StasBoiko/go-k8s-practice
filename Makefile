.PHONY: build run test lint swagger docker-up docker-down k8s-deploy

SWAG ?= $(shell which swag 2>/dev/null || echo $(shell go env GOPATH)/bin/swag)

build:
	CGO_ENABLED=0 go build -o ./bin/api ./cmd/api

run:
	go run ./cmd/api

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run ./...

swagger:
	$(SWAG) init -g cmd/api/main.go -o docs

docker-up:
	docker compose up -d

docker-down:
	docker compose down

k8s-deploy:
	@if [ -z "$(ENV)" ]; then echo "Error: ENV is required (e.g. make k8s-deploy ENV=dev)"; exit 1; fi
	docker build -t go-k8s-practice:latest .
	kind load docker-image go-k8s-practice:latest --name practice-cluster
	kubectl apply -f k8s/$(ENV)/
	kubectl rollout restart deployment/api-deployment -n $(ENV)
	kubectl rollout status deployment/api-deployment -n $(ENV)
