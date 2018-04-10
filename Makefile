git_hash=$(shell git rev-parse HEAD)
tag?=$(git_hash)

.DEFAULT_GOAL := build

build: clean
    GOOS=linux GOARCH=amd64 go build .
	docker build -f Dockerfile.deploy -t southwolf/ruyue_bus_bot:$(tag) .
	docker tag southwolf/ruyue_bus_bot:$(tag) registry.heroku.com/ruyue-bot/web
	docker push southwolf/ruyue_bus_bot:$(tag)
	docker push registry.heroku.com/ruyue-bot/web

clean:
	rm -f ./ruyue