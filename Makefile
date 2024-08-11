NAME=github_project_prometheus_exporter
VERSION ?= 0.1.0
IMAGE=atishikawa/$(NAME)

.PHONY: generate \
	docker/build \
	docker/push

generate:
	go generate ./...

# https://lipanski.com/posts/speed-up-your-docker-builds-with-cache-from
docker/build:
	docker build --target builder --cache-from $(IMAGE):builder -t $(IMAGE):builder .
	docker build -t $(IMAGE):$(VERSION) .

docker/push:
	docker push $(IMAGE):$(VERSION)
