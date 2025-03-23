# API Security with API Key Authentication

## Overview

Jorbites Notifier implements API Key authentication to secure access to its endpoints. This prevents unauthorized access to sensitive operations while maintaining a lightweight, dependency-free approach to security.

## How It Works

The API uses a static API Key mechanism where all protected endpoints require a valid API Key to be included in the request header. The API Key is configured through environment variables, allowing for easy management across different environments.

## Protected Endpoints

All API endpoints are protected by API Key authentication except for:
- `/health` - Health check endpoint (public)

Protected endpoints include:
- `/notifications` - Add notifications to the queue
- `/queue` - Get queue status

## Configuration

### Setting the API Key

The API Key is set through the `API_KEY` environment variable:

```bash
# In your terminal or .env file
API_KEY=your-secure-random-key
```

### Requirements

- The API Key should be a secure, random string
- Minimum recommended length is 32 characters
- The application will not start if the API Key is not configured

## Making Authenticated Requests

To access protected endpoints, include the API Key in the `X-API-Key` HTTP header:

```bash
# Example using curl
curl -X POST http://your-server/notifications \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secure-random-key" \
  -d '{
    "type": "NEW_COMMENT",
    "recipient": "user@example.com",
    "metadata": {
      "commentId": "12345",
      "authorName": "User1"
    }
  }'
```

### Authentication Responses

- Missing API Key: Returns `401 Unauthorized` with message "API key is missing"
- Invalid API Key: Returns `401 Unauthorized` with message "Invalid API key"
- Valid API Key: Proceeds with the requested operation

## Security Best Practices

1. **Generate Strong API Keys**
   - Use a secure random generator
   - Example: `openssl rand -hex 32`

2. **Protect Your API Key**
   - Never expose the API Key in client-side code
   - Don't commit API Keys to source control
   - Use environment variables or secure vaults

3. **Rotate Keys Periodically**
   - Change API Keys regularly, especially after team member changes
   - Implement a process for securely distributing new keys

4. **Transport Security**
   - Always use HTTPS in production to encrypt API Keys in transit

5. **Role-Based Keys (Future Enhancement)**
   - For more advanced scenarios, implement different keys for different permissions
   - Consider implementing key expiration

## Example Configuration Files

### .env.example
```
API_KEY=your-secure-random-key
PORT=8080
```
