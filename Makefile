export BINARY_NAME=terminator
export DC=docker-compose -f docker-compose-build.yaml
export IMAGE_NAME=vevo/$(BINARY_NAME):0.1.0-b$(GO_PIPELINE_COUNTER)
export LATEST_TAG=vevo/$(BINARY_NAME):latest


login:
ifeq ($(DOCKER_USER),vevocd)
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"
endif

clean:
	$(DC) down --rmi local --remove-orphans
	$(DC) rm -f

build: clean
	$(DC) run binary
	$(DC) build
	docker tag $(IMAGE_NAME) $(LATEST_TAG)

build_and_publish: build push

push: login
	docker push $(IMAGE_NAME)
	docker push $(LATEST_TAG)
