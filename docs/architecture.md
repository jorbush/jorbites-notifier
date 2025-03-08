## Architecture Overview

Jorbites Notifier is designed as a lightweight microservice that provides notification capabilities for the Jorbites platform. The service uses a simple in-memory FIFO queue to process notifications in the order they are received.

### Key Components

- **Notification Queue**: In-memory FIFO queue for processing notifications
- **HTTP Server**: RESTful API for managing notifications
- **Email Sender**: Component for sending email notifications

### Flow

1. Clients send notification requests to the API
2. Notifications are added to the queue with "pending" status
3. The queue processor picks up notifications in FIFO order
4. Notifications are processed according to their type
