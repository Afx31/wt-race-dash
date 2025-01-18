BINARY_NAME=bin/wt-race-dash
BINARY_DIR=./cmd/wt-race-dash
DATALOG_BINARY_NAME=bin/wt-datalogging
DATALOG_BINARY_DIR=../wt-datalogging

build:
	go build -o ${BINARY_NAME} ${BINARY_DIR}

run: build
	./${BINARY_NAME}

clean:
	go clean
	rm ./${BINARY_NAME}

buildall:
	make build
	cd ${DATALOG_BINARY_DIR} && go build -o ${DATALOG_BINARY_NAME}