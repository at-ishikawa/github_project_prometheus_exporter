NAME=github_project_prometheus_exporter
VERSION=0.1.0
IMAGE_TAG=atishikawa/$(NAME):$(VERSION)

.PHONY: generate \
	docker/build \
	docker/push

generate:
	go generate ./...

docker/build:
	docker build -t $(IMAGE_TAG) .

docker/push:
	docker push $(IMAGE_TAG)
