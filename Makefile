build-refresh:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/refresh src/refresh/main.go

build-open:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/open src/open/main.go

build: build-refresh build-open
