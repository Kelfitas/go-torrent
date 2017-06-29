install:
	go get

build: install
	go build main.go

run: install
	go run main.go

clean:
	go clean
	git clean -idx