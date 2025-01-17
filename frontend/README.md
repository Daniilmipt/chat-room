# Chatroom Application
## Overview
This is a decentralized chatroom application built with Golang and leveraging the `chat` application which uses `go-libp2p-pubsub` library for peer-to-peer communication. The `server` package provides an HTTP interface for managing chatrooms, handling user sessions, and serving the front-end views for interacting with the chat system.
### Features
- Decentralized chatroom communication using `go-libp2p-pubsub`.
- REST API for chatroom operations.
- Front-end built with HTML, CSS, and JavaScript.
- User session management via cookies.
- Live updates for chatroom lists and messages.


## Installation

### Steps
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd chatroom/server
   ```
2. Build the application:
   ```bash
   go build -o chatroom
   ```
3. Run the server:
   ```bash
   ./chatroom
   ```

Open `http://localhost:8080` in your browser. By default we use port `8080`, but you can set it in `/config/config.yml`

