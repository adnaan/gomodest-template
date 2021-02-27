install:
	go get -u -v ./... && cd assets && yarn install
watch: install
	echo "watching go files and assets directory..."; \
	air -d -c .air.toml & \
	cd assets && yarn watch & \
	wait; \
	echo "bye!"
watch-go:
	air -c .air.toml
watch-assets:
	cd assets && yarn watch
run-go:
	go run main.go
build-assets:
	cd assets && yarn build
build-docker:
	docker build -t gomodest-template .
run-docker:
	docker run -it --rm -p 3000:3000 gomodest-template:latest
generate-todos-models:
	go generate ./samples/todos/generator