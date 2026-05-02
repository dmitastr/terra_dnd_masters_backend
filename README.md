# 🐉 DnD Event Scheduling Server (Go)

A backend service for managing and scheduling tabletop RPG sessions like Dungeons & Dragons.  
Built with a focus on clean architecture, simplicity, and scalability using Go.

---

## 📌 Overview

This project is a RESTful server that allows users to create, manage, and schedule DnD events (sessions, campaigns, game nights). It is designed as a lightweight, extensible backend that can later be integrated with a frontend or expanded with persistent storage.

---

## 🚀 Features

- Create and manage DnD events
- Schedule sessions with date/time handling
- Uses Postgres for data storage (via docker container)
- RESTful JSON API
- Clean and modular project structure
- Dockerized for easy setup and deployment

---

## 🛠️ Tech Stack

- **Language:** Go (Golang)
- **HTTP Server:** `gin` (standard library)
- **Architecture:** Layered (handlers → services → repository)
- **Data Storage:** PostgreSQL
- **Containerization:** Docker
- **Documentation:** Swagger

---

## 🧠 Architecture & Design

This project follows a modular and testable architecture:

```
/cmd # Application entrypoint
/internal
    /handler # HTTP handlers (request/response logic)
    /service # Business logic
    /repository # Data access layer (in-memory implementation)
```



### Key design decisions:

- **Using powerful but light library**  
  No heavy frameworks — uses Go’s `gin` for better control and understanding of HTTP internals.

- **Repository pattern**  
  Abstracts storage layer to allow easy migration to databases like PostgreSQL.

- **Separation of concerns**  
  Clear boundaries between transport, business logic, and data layers.

- **Stateless API design**  
  Makes the service easy to scale and containerize.

---

## 🐳 Running with Docker

```bash
docker-compose build 
docker-compose up -d
```

The server will be available at: http://localhost:8080/api/v1/dnd

Documentation: http://localhost:8080/api/v1/dnd/swagger/index.html