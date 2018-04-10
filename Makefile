git_hash=$(shell git rev-parse HEAD)
tag?=$(git_hash)

.DEFAULT_GOAL := build

build: clean
	docker build -t southwolf/ruyue_bus_bot:$(tag) .
	docker push southwolf/ruyue_bus_bot:$(tag)

clean:
	rm -f ./ruyue