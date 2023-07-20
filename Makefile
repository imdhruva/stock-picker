APP_NAME = stock-picker
DOCKER_UPSTREAM = ghcr.io/imdhruva/
DOCKER_TAG ?= latest
DOCKER_IMAGE = ${DOCKER_UPSTREAM}${APP_NAME}:${DOCKER_TAG}
CHART_PATH = chart/${APP_NAME}/
NAMESPACE = default

.PHONY : help
help : Makefile ## print help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY:
all: lint fmt test docker-build docker-push helm-upgrade

.PHONY:
lint: ## Lint code
	golangci-lint run \
		--enable gocyclo \
		--enable revive \
		--enable depguard \
		--enable nakedret \
		--enable staticcheck \
		--enable errcheck \
		--enable gofmt \
		--exclude-use-default=false \
		--deadline=5m

.PHONY:
fmt: ## Format all files
	go fmt .
	helm lint ${CHART_PATH} --strict

.PHONY:
test: ## Test golang files
	go test . -v


.PHONY:
run: ## Run application
	go run .

.PHONY:
docker-build: ## Build docker container
	docker build -t ${DOCKER_IMAGE} .

.PHONY:
docker-push: ## Push docker container to repository
	docker push ${DOCKER_IMAGE}


.PHONY: helm-upgrade
helm-upgrade: helm-diff ## run helm upgrade
	@echo
	@echo -------------------------------------
	@echo running helm upgrade
	@echo -------------------------------------
	@echo
	helm upgrade ${APP_NAME} ${CHART_PATH} \
		--install \
		--wait \
		--set image.repository=${DOCKER_UPSTREAM}${APP_NAME} \
		--set image.tag=${DOCKER_TAG} \
		-n ${NAMESPACE}

.PHONY: helm-diff
helm-diff: ## check diff of k8s manifests for intended changes to be made
	@echo
	@echo -------------------------------------
	@echo running helm diff
	@echo -------------------------------------
	@echo
	helm diff upgrade ${APP_NAME} ${CHART_PATH} \
		--install \
		--set image.tag=${DOCKER_TAG} \
		--allow-unreleased \
		-n ${NAMESPACE}