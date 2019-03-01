include .env
export $(shell sed 's/=.*//' .env)

all: test build deploy

dev:
	gin main.go

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

