# Wiki

## Use Cases

Users can add, edit and delete there goals. Goals are shown in a timeline view.
The user can share the timeline using a dedicated share link with stakeholders. Stakeholders can view the timeline without any registration. Users can only manage there own goals and timeline.

## System Design

### How many users?

We expect 1.000 Users in total and 100 active users at a time.

### How often do they open the app?

A user will login once a day and will requests the timeline once each hour. The shared timeline will be requests (in average) 10 times per user per second. There can be peaks up to 100 times per user per second.

Simplified formula: 10 ops/second _ 100 users/day = 1000 (ops _ users)/second => SQLite can handle aprox. 1000 isnerts/second.

## Database Design

```mermaid
erDiagram
    users {
        INTEGER id PK
        TEXT email
        TEXT password_hash
        DATETIME created_at
        DATETIME updated_at
    }

    branding {
        INTEGER id PK
        INTEGER user_id FK
        TEXT title
        TEXT description
    }

    goals {
        INTEGER id PK
        INTEGER user_id FK
        TEXT goal
        DATETIME due
        BOOL visible_to_public
        BOOL achieved
        DATETIME created_at
        DATETIME updated_at
    }

    share {
        INTEGER id PK
        INTEGER user_id FK
        TEXT public_id
    }

    users ||--o{ goals : "has (CASCADE)"
    users ||--o{ share : "creates (CASCADE)"
    users ||--|| branding : "has (CASCADE)"
```

## Scaling

Before scaling any services it is recommend to scale the hardware. Go can handle many requests with less CPU and RAM usage. After that, it is useful to identify the bottlenecks of the app. Which can be found in Issue [#16](https://github.com/bit8bytes/goalkeepr/issues/16). One predictable bottleneck will be SQLite, particularly the write operations.

## Development

Prerequisites:

- [Go-task](https://taskfile.dev/docs/installation#go-modules)
- [Golang (latest)](https://go.dev)
- [Tailwindcss](https://tailwindcss.com/blog/standalone-cli) 
