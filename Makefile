.PHONY: docker run clean test

docker:
	@if [ ! -f .env ]; then echo ".env file not found"; exit 1; fi
	$(eval API_KEY := $(shell grep API_KEY .env | cut -d '=' -f2))
	$(eval SMTP_USER := $(shell grep SMTP_USER .env | cut -d '=' -f2))
	$(eval SMTP_PASSWORD := $(shell grep SMTP_PASSWORD .env | cut -d '=' -f2))
	$(eval MONGO_URI := $(shell grep MONGO_URI .env | cut -d '=' -f2))
	$(eval MONGO_DB := $(shell grep MONGO_DB .env | cut -d '=' -f2))
	docker build -t jorbites-notifier .
	docker run -d --name jorbites-notifier -p 8080:8080 -e API_KEY="$(API_KEY)" -e SMTP_USER=$(SMTP_USER) -e SMTP_PASSWORD=$(SMTP_PASSWORD) -e MONGO_URI=$(MONGO_URI) -e MONGO_DB=$(MONGO_DB) jorbites-notifier

run:
	@if [ ! -f .env ]; then echo ".env file not found"; exit 1; fi
	$(eval API_KEY := $(shell grep API_KEY .env | cut -d '=' -f2))
	$(eval SMTP_USER := $(shell grep SMTP_USER .env | cut -d '=' -f2))
	$(eval SMTP_PASSWORD := $(shell grep SMTP_PASSWORD .env | cut -d '=' -f2))
	$(eval MONGO_URI := $(shell grep MONGO_URI .env | cut -d '=' -f2))
	$(eval MONGO_DB := $(shell grep MONGO_DB .env | cut -d '=' -f2))
	API_KEY=$(API_KEY) \
	SMTP_USER=$(SMTP_USER) \
	SMTP_PASSWORD=$(SMTP_PASSWORD) \
	MONGO_URI=$(MONGO_URI) \
	MONGO_DB=$(MONGO_DB) \
	go run cmd/server/main.go


clean:
	-docker stop jorbites-notifier
	-docker rm jorbites-notifier

test:
	go test -v ./...
