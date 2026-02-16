VERSION = 1.12.0

build:
	docker buildx build --platform linux/amd64,linux/arm64 -t onrik/gaws:$(VERSION) .
