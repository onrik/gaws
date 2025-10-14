VERSION = 1.11.1

build:
	docker buildx build --platform linux/amd64,linux/arm64 -t onrik/gaws:$(VERSION) .
