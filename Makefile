APP_NAME=go-torrent

install:
	go get $(INSTALL_ARGS)

run: install
	go run $(RUN_ARGS) main.go

build: install
	GOOS=$(GOOS) GOARCH=$(GOARCH) \
    go build $(BUILD_ARGS) main.go

build-win32: GOOS=windows
build-win32: GOARCH=386
build-win32: BUILD_ARGS=-o $(APP_NAME)-x32.exe
build-win32: build

build-win64: GOOS=windows
build-win64: GOARCH=amd64
build-win64: BUILD_ARGS=-o $(APP_NAME)-x64.exe
build-win64: build

build-mac32: GOOS=darwin
build-mac32: GOARCH=386
build-mac32: BUILD_ARGS=-o $(APP_NAME)-x32.app
build-mac32: build

build-mac64: GOOS=darwin
build-mac64: GOARCH=amd64
build-mac64: BUILD_ARGS=-o $(APP_NAME)-x64.app
build-mac64: build

build-linux32: GOOS=linux
build-linux32: GOARCH=386
build-linux32: BUILD_ARGS=-o $(APP_NAME)-x32
build-linux32: build

build-linux64: GOOS=linux
build-linux64: GOARCH=amd64
build-linux64: BUILD_ARGS=-o $(APP_NAME)-x64
build-linux64: build

clean:
	go clean
	rm -f ./$(APP_NAME)*
	git clean -idx