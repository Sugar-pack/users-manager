.PHONY: docker-run
docker-run: vet lint
	@docker-compose up --build -d --remove-orphans

.PHONY: docker-restart
docker-restart:
	@docker-compose up -d

vet:  ## Run go vet
	go vet ./...

lint: ## Run go lint
	golangci-lint run

test:
	go test -cover -count 1 ./...