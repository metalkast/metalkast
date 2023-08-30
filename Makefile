OUT_DIR = _out

.PHONY: all clean
all: build

clean:
	rm -rf $(OUT_DIR)

.PHONY: build
build:
	mkdir -p $(OUT_DIR)
	go build -o $(OUT_DIR)/kast ./cmd/kast
