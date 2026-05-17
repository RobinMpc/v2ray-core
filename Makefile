.PHONY: all prepare build copy-config

all: prepare build copy-config

prepare:
	mkdir -p bin conf

build: prepare
	go build -o bin/v2ray main/main.go

copy-config: prepare
	cp config.json conf/config.json
