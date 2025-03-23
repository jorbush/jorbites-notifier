.PHONY: docker run clean

docker:
	@if [ ! -f .env ]; then echo ".env file not found"; exit 1; fi
	$(eval API_KEY := $(shell grep API_KEY .env | cut -d '=' -f2))
	docker build -t jorbites-notifier .
	docker run -d --name jorbites-notifier -p 8080:8080 -e API_KEY="$(API_KEY)" jorbites-notifier

run:
	@if [ ! -f .env ]; then echo ".env file not found"; exit 1; fi
	$(eval export $(shell sed -e 's/=.*//' .env))
	$(eval export $(shell grep -v '^#' .env | xargs -0))
	go run cmd/server/main.go

clean:
	-docker stop jorbites-notifier
	-docker rm jorbites-notifier
