# Jorbites Notifier

A lightweight notification service for Jorbites platform written in Go.

## Overview

Jorbites Notifier is a microservice designed to handle notifications for the Jorbites platform. It provides a simple FIFO queue for processing notifications and supports email delivery.

## Features

- Simple FIFO notification queue
- Email notifications
- RESTful API for notification management
- Lightweight implementation with minimal dependencies

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check endpoint |
| `/notifications` | POST | Add a notification to the queue |
| `/queue` | GET | Get the current queue status |

## Running the service

```bash
# Install dependencies
go mod tidy

# Run the service
go run cmd/server/main.go
```

Or using the Makefile:

```bash
make run
```

## Running the service in Docker

```bash
make docker
```

## Documentation

For detailed documentation, check the [docs](./docs) directory.

## Testing

_TODO_
