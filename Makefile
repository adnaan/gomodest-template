install:
	go get -v ./... && cd assets && npm install
install-x64:
	go get -v ./... && cd assets && npm install --target_arch=x64
watch: install
	echo "watching go files and assets directory..."; \
	air -d -c .air.toml & \
	cd assets && npm run watch & \
	wait; \
	echo "bye!"
watch-go:
	air -c .air.toml
watch-assets:
	cd assets && npm run watch
run-go:
	go run main.go
build-assets:
	cd assets && npm run build
build-docker:
	docker build -t gomodest-template .
run-docker:
	docker run -it --rm -p 3000:3000 gomodest-template:latest
generate-todos-models:
	go generate ./samples/todos/generator