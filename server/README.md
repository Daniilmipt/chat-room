# Chatroom Application

## Overview
This is a decentralized chatroom application built with Golang and leveraging the `chat` application which uses `go-libp2p-pubsub` library for peer-to-peer communication. The `server` package provides an HTTP interface for managing chatrooms, handling user sessions, and serving the front-end views for interacting with the chat system.

### Features
- Decentralized chatroom communication using `go-libp2p-pubsub`.
- REST API for chatroom operations.
- Front-end built with HTML, CSS, and JavaScript.
- User session management via cookies.
- Live updates for chatroom lists and messages.

## Example Workflow

1. Start the first instance of the application:
   ```bash
   ./chatroom
   ```
2. Login there, choose room

![alt text](./images/login.png)

3. Send message

![alt text](./images/login.png)

2. Start another instance of the application on the same network:
   ```bash
   ./chat -nick=Bob -room=developers
   ```

## Folder Structure

### **server**

#### **pkg**
- **models**
  - `messages.go`: Defines the `Messages` struct for representing chat messages and includes a `Validate` method.
- `file.go`: Utility function `GetLastLine` to retrieve the last line of a log file.
- `logger.go`: Configures the logger using Zap with output to the console and log files.

#### **router**
- `handler.go`: Main handler for chat-related API endpoints.
- `handler_internal.go`: Implements core chat operations such as joining rooms, sending messages, and reading logs.
- `handler_view.go`: Handles front-end view routing, including loading the room list and messages.

#### **Main Files**
- `main.go`: Initializes the Gin server, configures middleware, and starts the application.

### **Front-End Files**
- `room_list.html`: Interactive UI for viewing and joining chatrooms.
- `login.html`: To set nick or room and join to last.
- `room.html`: UI for room, where you can see messages and send by yourself.

## Installation

### Prerequisites
- [Golang](https://golang.org/) (version 1.23.4)

### Steps
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd chatroom/server
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Build the application:
   ```bash
   go build -o chatroom
   ```

4. Run the server:
   ```bash
   ./chatroom
   ```

Or you can run ready `.exe` file


5. Open `http://localhost:8080` in your browser.

## Usage

1. Open the application in your browser.
2. Use the "Join a New Room" button to join or create a chatroom.
3. View messages in real-time.
4. Log out to clear your session and return to the room list.

## Future

- Add authentification mechanism. Where you can create private rooms.
- Transform to social network. Send not only messages, but posts and other.

## Contribution
Feel free to submit issues or pull requests for improvements or bug fixes. Contributions are always welcome!
