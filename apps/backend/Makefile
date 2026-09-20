BIN := bin/backend
GIT_SHA := $(shell git rev-parse HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -X boot.dev/linko/internal/build.GitSHA=$(GIT_SHA) -X boot.dev/linko/internal/build.BuildTime=$(BUILD_TIME)

.PHONY: dev prod clean

dev:
	go build -gcflags='all=-N -l' -ldflags="$(LDFLAGS)" -o $(BIN) .

prod:
	go build -trimpath -ldflags="-s -w $(LDFLAGS)" -o $(BIN) .

clean:
	rm -f $(BIN)
