# Notification Queue System

The notification queue is the core component of Jorbites Notifier. It manages the processing of notifications in a FIFO (First In, First Out) manner.

## Queue Features

- **In-memory Storage**: Queue is currently stored in memory
- **FIFO Processing**: Notifications are processed in the order they are received
- **Status Tracking**: Each notification has a status that is updated during processing
- **Automatic Processing**: Queue processor runs in the background
- **Thread-safe**: Concurrent access to the queue is handled with mutex locks

## Notification Lifecycle

1. **Creation**: Notification is created by a client through the API with initial status "pending"
2. **Queuing**: Notification is added to the end of the queue
3. **Processing**: When the notification reaches the front of the queue, its status changes to "processing"
4. **Removal**: Processed notifications are removed from the queue

## Queue Status

The queue status can be checked via the API at any time. It provides:
- Total count of notifications in the queue
- List of all notifications with their current status

## Notification Statuses

| Status | Description |
|--------|-------------|
| `pending` | Notification is waiting in the queue to be processed |
| `processing` | Notification is currently being processed |

## Future Improvements

- **Persistence**: Store the queue in a database to survive service restarts
- **Retry Mechanism**: Implement automatic retry for failed notifications
- **Multiple Workers**: Add support for parallel processing with multiple workers
- **Prioritization**: Add priority levels for different notification types
