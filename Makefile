.PHONY: docker run clean

docker:
	@if [ ! -f .env ]; then echo ".env file not found"; exit 1; fi
	$(eval API_KEY := $(shell grep API_KEY .env | cut -d '=' -f2))
	$(eval SMTP_USER := $(shell grep SMTP_USER .env | cut -d '=' -f2))
	$(eval SMTP_PASSWORD := $(shell grep SMTP_PASSWORD .env | cut -d '=' -f2))
	docker build -t jorbites-notifier .
	docker run -d --name jorbites-notifier -p 8080:8080 -e API_KEY="$(API_KEY)" -e SMTP_USER=$(SMTP_USER) -e SMTP_PASSWORD=$(SMTP_PASSWORD) jorbites-notifier

run:
	@if [ ! -f .env ]; then echo ".env file not found"; exit 1; fi
	$(eval API_KEY := $(shell grep API_KEY .env | cut -d '=' -f2))
	$(eval SMTP_USER := $(shell grep SMTP_USER .env | cut -d '=' -f2))
	$(eval SMTP_PASSWORD := $(shell grep SMTP_PASSWORD .env | cut -d '=' -f2))
	API_KEY=$(API_KEY) \
	SMTP_USER=$(SMTP_USER) \
	SMTP_PASSWORD=$(SMTP_PASSWORD) \
	go run cmd/server/main.go


clean:
	-docker stop jorbites-notifier
	-docker rm jorbites-notifier
