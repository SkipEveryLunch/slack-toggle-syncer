IMAGE_NAME = slack-toggle-syncer

.PHONY: build run

build:
	docker build -t $(IMAGE_NAME) .

run: build
	docker run --rm --env-file .env $(IMAGE_NAME)
