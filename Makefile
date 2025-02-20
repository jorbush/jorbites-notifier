.PHONY: docker

docker:
	docker build -t jorbites-notifier .
	docker run -d --name jorbites-notifier -p 8080:8080 jorbites-notifier

run:
	go run cmd/server/main.go
