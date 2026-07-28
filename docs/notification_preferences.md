# Notification Preferences System

Jorbites allows users to granularly enable and disable notifications based on the service/category. This document details how notification preferences are stored, processed, and enforced.

## Architecture & Data Flow

```
+------------------------------------+
| Jorbites Main App (Next.js)        |
| - Settings UI (Preferences Toggles)|
| - PATCH /api/notificationPreferences|
+-----------------+------------------+
                  | Updates
                  v
+------------------------------------+
| MongoDB (User Document)            |
| - emailNotifications (master flag) |
| - notificationPreferences (object) |
+-----------------+------------------+
                  | Reads
                  v
+------------------------------------+
| Jorbites Notifier (Go Service)     |
| - IsNotificationCategoryEnabled()  |
| - Filters Email & Push Delivery    |
+------------------------------------+
```

## Preference Categories & Mapping

Rather than exposing raw notification types to users, notifications are grouped into 6 user-friendly categories:

| Category | Field Name | Notification Types Included | Default |
|---|---|---|---|
| **Social** | `social` | `NEW_LIKE`, `NEW_COMMENT`, `MENTION_IN_COMMENT` | `true` |
| **New Content** | `newContent` | `NEW_RECIPE`, `NEW_BLOG` | `true` |
| **Events & Challenges** | `eventsAndChallenges` | `NEW_EVENT`, `EVENT_ENDING_SOON`, `NEW_CHALLENGE` | `true` |
| **Quests** | `quests` | `NEW_QUEST`, `QUEST_FULFILLED` | `true` |
| **Voting** | `voting` | `NEW_VOTATION`, `VOTATION_RESULT` | `true` |
| **Achievements** | `achievements` | `NEW_BADGE`, `VERIFIED` | `true` |

### System & Non-Configurable Notifications

The following notification types bypass category checks and are **always sent** when triggered:
- `FORGOT_PASSWORD`
- `NOTIFICATIONS_ACTIVATED`

## Master Toggle vs. Category Preferences

1. **`emailNotifications` (Boolean)**: Master toggle. When set to `false`, no email notifications are sent under any circumstance (except system-critical ones like `FORGOT_PASSWORD`).
2. **`notificationPreferences` (Object)**: Embedded document containing boolean flags per category. Controls both email and push notification channels.

If a user has `notificationPreferences: null` (legacy users), all categories default to `true` (opt-out model).

## Implementation in Go (`jorbites-notifier`)

The evaluation logic is encapsulated in `User.IsNotificationCategoryEnabled(notifType)` (`internal/models/user.go`):

```go
func (u *User) IsNotificationCategoryEnabled(notifType NotificationType) bool
```

This method is invoked prior to delivering both email and push notifications in `queue.go`.
