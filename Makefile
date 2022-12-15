# starts the dockerized environment
compose-up:
	docker-compose up &

# stops the dockerized environment
compose-down:
	docker-compose down &

# removes the created dockeri containers and images
docker-clean: compose-down
	docker container prune --force
	docker image rm faceit/userservice

# creates the docker image
package: build
	docker build -t faceit/userservice .

# generate
generate: generate-api generate-mocks update

# generate the Echo server type definitions and handler wrappers based on the OpenAPI definition
generate-api:
	oapi-codegen --config api/config/users_types.yaml api/users.yaml
	oapi-codegen --config api/config/users_server.yaml api/users.yaml

# generate mocks used by tests based on the defined interfaces
generate-mocks:
	cd pkg && find . -type f -name 'mock_*.go' -delete && mockery --inpackage --recursive --all
	cd internal && find . -type f -name 'mock_*.go' -delete && mockery --inpackage --recursive --all

# installs the required Go tools
install:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
	go install github.com/vektra/mockery/v2@latest

# updates the required Go libraries
update:
	go mod tidy
	go mod vendor

# cleans up the Go workspace
clean:
	rm -r bin/
	rm -r vendor/

# starts the server
run:
	go run main.go

# runs all the tests and requires DB and RabbitMQ connection
test:
	go fmt ./...
	go test -cover --tags=integration ./...

# runs only the unit tests
unittest:
	go fmt ./...
	go test -cover ./...

# builds the binary
build: unittest
	go build -o bin/userservice main.go
