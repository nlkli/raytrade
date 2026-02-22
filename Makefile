APP=raytrade
LDFLAGS=-ldflags="-s -w"

.PHONY: build clean run

default: build

build:
	go build -o bin/$(APP) $(LDFLAGS) .

run: build
	./bin/$(APP)

clean:
	rm -rf bin/
