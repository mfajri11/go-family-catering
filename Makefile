SHELL := /bin/bash
.DEFAULT_GOAL := help


help:
	@echo -e Usage: make [TARGET] [...ARGUMENTS]\\n;
	@echo -e Description:;
	@printf '%40s\n\n' "make utility for family-catering app"
	@echo -e TARGETS:;
	@grep -E '[a-z\-]:.*?##.*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "%4s\033[36m%-15s\033[0m %s\n", " ", $$1, $$2}';
	@echo
	@echo Arguments:
	@printf '%5s %59s\n' "n" "number of step (used by 'step' recipe/target)";
	@printf '%13s %146s\n' "container" "if true will used docker version, otherwise just use go server without docker (optional default 'true', used by all target except 'version')";
.PHONY: help

install: ## install the application or if use container will build the containers needed
	@if [ "$(container)" = 'false' ]; then \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    	go build -ldflags="-s -w" -o ./bin/fcat ./cmd/main.go; \
	else \
		make compose-up; \
	fi
.PHONY: start


version: ## print the current version of the family-catering app
	@ ./bin/fcat version || go run ./cmd/main.go version
.PHONY: version

start: ## running all containers built by 'make start'
	@if [[ '$(container)' = 'false' || '$(container)' = 'False' || '$(container)' = 'FALSE' ]]; then \
		./bin/fcat run || go run ./cmd/main.go run; \
	else \
		make compose-start; \
	fi
.PHONY: start

stop: ## stop running containers by 'make start'
	@make compose-stop
	# TODO: stop running fcat process (not a docker container)
.PHONY: stop

down: ## stopping and removing built containers
	@if [[ '$(container)' = 'false' || '$(container)' = 'False' || '$(container)' = 'FALSE' ]]; then \
		sudo rm -f ./bin/fcat; \
	else  \ 
		@make compose-down; \
	fi
.PHONY: down

migrate: ## running migration all the way up from active version of the schema (up migrations)
	@if [ '$(container)' = 'false' ]; then \
	 	go run ./cmd/main.go migrate; \
	 else \
		docker start --attach fcat \
		|| docker run fcat migrate; \
	fi
.PHONY: migrate
	

rollbacks: ## running migration all the way down from active version of the schema (down migrations)
	@if [ '$(container)' = 'false' ]; then \
		./bin/fcat rollbacks; \
		|| go run ./cmd/main.go rollbacks \
		\
	else \
		docker start fcat-rollbacks \
		|| docker run --name fcat-rollbacks fcat rollbacks; \
	fi 
.PHONY: rollbacks

step: ## running migration n step up/down relatively from active version of the schema (if n > 0 it will migrate up, otherwise is down)
	@if [ -z '$(n)' ]; then \
		echo missing required argument \'n\'; \
		\
	elif [[ '$(container)' = 'false' || '$(container)' = 'False' || '$(container)' = 'FALSE' ]]; then \
			./bin/fcat step -n $(n) \
			|| go run ./cmd/main.go step -n $(n); \
		\
	else \
		docker start fcat-step -n $(n) \
		|| docker run --name fcat-step fcat step -n $(n); \
	fi
.PHONY: step

test:
	@ go test -count 1 -coverprofile coverage ./...;
	 | cat coverage \
	 | grep -v mock > coverage;
	@ go tool cover -html=coverage
	@ sudo rm -f ./coverage
.PHONY: test

compose-up: 
	@docker-compose up --build -d
.PHONY: compose.up

compose-start: 
	@docker-compose start && docker attach fcat
.PHONY: compose-start

compse-stop:
	@docker-compose stop
.PHONY: compse.stop

compose-down:
	@docker-compose down -f -v --remove-orphans
.PHONY: compose-down
