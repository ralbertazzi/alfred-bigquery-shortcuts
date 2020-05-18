TARGET_DIR := target
WORKFLOW_FILE := $(TARGET_DIR)/alfred-bigquery-shortcuts.alfredworkflow


target/:
	mkdir target

clean:
	@[ -d $(TARGET_DIR) ] && rm -r $(TARGET_DIR) || true

workflow: build clean target/
	zip $(WORKFLOW_FILE) \
	info.plist \
	icon.png \
	bin/refresh \
	bin/open

build-refresh:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/refresh src/refresh/main.go

build-open:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/open src/open/main.go

build: build-refresh build-open
