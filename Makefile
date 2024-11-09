BINARY_NAME=bin/wills-race-dash-go
BINARY_DIR=./cmd/wills-race-dash-go

build:
	go build -o ${BINARY_NAME} ${BINARY_DIR}

run: build
	./${BINARY_NAME}

clean:
	go clean
	rm ./${BINARY_NAME}