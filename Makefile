.PHONY: clean
clean:
	rm -rf bin/

.PHONY: lint
lint:
	golangci-lint run --timeout 300s -v ./...

.PHONY: build
build:
	go build  -o bin/ogre -ldflags="-X 'main.buildDateTime=$$(date +%Y-%m-%dT%H:%M:%S%z)' -X 'main.gitCommit=$$CI_COMMIT_SHORT_SHA' -X 'main.versionTag=$$CI_COMMIT_TAG'  -X 'main.buildAuthor=$$GITLAB_USER_LOGIN'" cmd/ogre/ogre.go

.PHONY: fmt
fmt:
	go fmt ./...