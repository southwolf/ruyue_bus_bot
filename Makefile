git_hash=$(shell git rev-parse HEAD)
tag?=$(git_hash)

.DEFAULT_GOAL := build

build: clean
	GOOS=linux GOARCH=amd64 go build .
	docker build -t southwolf/ruyue:$(tag) .
	docker push southwolf/ruyue:$(tag)

clean:
	rm -f ./ruyue