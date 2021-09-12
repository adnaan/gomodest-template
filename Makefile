$(eval GOPATH=$(shell go env GOPATH))
install: install-air
	go get -v ./... && cd assets && npm install
install-x64: install-air
	go get -v ./... && cd assets && npm install --target_arch=x64
install-air:
	curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b ${GOPATH}/bin
watch: install
	echo "watching go files and assets directory..."; \
	${GOPATH}/bin/air -d -c .air.toml & \
	cd assets && npm run watch & \
	wait; \
	echo "bye!"
watch-x64: install-x64
	echo "watching go files and assets directory..."; \
	${GOPATH}/bin/air -d -c .air.toml & \
	cd assets && npm run watch & \
	wait; \
	echo "bye!"
watch-go:
	${GOPATH}/bin/air -c .air.toml
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