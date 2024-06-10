SRC = $(shell find . -name '*.go')

rdo: $(SRC)
	go mod tidy
	go build -o rdo ./cmd/rdo
	go install github.com/hillside-labs/rdo/cmd/rdo
