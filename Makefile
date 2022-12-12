VERSION := "v0.0.1"
TARGET=_build/azurefileshare
PKG=$(shell go list)
LDFLAGS=-ldflags "-X '$(PKG)/cmd/azurefileshare.version=$(VERSION)'"

all: build

clean:
	rm -f $(TARGET)

build:
	go build -o $(TARGET) $(LDFLAGS) main.go

install:
	cp $(TARGET) ${GOPATH}/bin/