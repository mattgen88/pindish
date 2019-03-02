include .env
export $(shell sed 's/=.*//' .env)

all: test build deploy

dev: selfsigned.key selfsigned.crt
	gin --certFile=selfsigned.crt --keyFile=selfsigned.key main.go

selfsigned.key:
selfsigned.crt:
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout selfsigned.key -out selfsigned.crt

test:
	go test ./...

build:
	go build
	go install

deploy:
	git push heroku master
	heroku open

local: test build
	heroku local web

