.PHONY: build all
all: build
	faas-cli up --build-arg GO111MODULE=on -f dedofeup.yml -g https://func.skmobi.com

build: template
	faas-cli build --build-arg GO111MODULE=on -f dedofeup.yml

template:
	faas-cli template pull https://github.com/bmcstdio/golang-http-template.git --overwrite

