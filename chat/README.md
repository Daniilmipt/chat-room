# Chat Room Application
This is a decentralized chat-room console application built using the `go-libp2p` framework. It allows users to join chat rooms, send messages to other peers, and receive messages in real-time. The application logs all messages in the `/messages`
## Features
- Each launch create connection to `host` and `port`, which described in `config.yml`.
- Decentralized communication using `go-libp2p-pubsub`.
- mDNS-based peer discovery for connecting with other users on the same network.
- Message logging for each chat room.
## Directory Structure
- `chat/`: Contains the core logic of the chat application.
  - `pkg/chatroom.go`: Handles PubSub topic subscriptions, message publishing, and receiving.
  - `pkg/logger.go`: Configures logging using the Zap library.
- `service/`: Implements console appliation and peer discovery.
  - `discovery.go`: Sets up mDNS discovery for peer connections.
  - `service.go`: Manages peer connections and message handling.
- `messages/`: Directory where chat logs are stored for each room.
- `config.yml`: Configuration file for setting the host and port of the application.
- `main.go`: Entry point for the application.
## Why we save messages in file
We use it as an auxiliary application for the main `chat-rooms` application in folder `/server`. So, we use files as a `database storage`. This files are used by `chat-rooms` application.
## Prerequisites
Ensure the following are installed on your system:
- [Go](https://golang.org/) (version 1.23.4)
## Configuration
Modify the `config.yml` file to set the host and port for the application:
```yaml
host: "0.0.0.0"
port: "8082"
```
## Building the Application
1. Clone the repository:
   ```bash
   git clone <repository_url>
   cd <repository_name>
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Build the application:
   ```bash
   go build -o chat
   ```
Or you can run ready `.exe` file
## Running the Application
1. Start the application by specifying the nickname and room name:
   ```bash
   ./chat -nick=<nickname> -room=<room_name>
   ```
   - `nickname`: The name to identify yourself in the chat (default: `anonymous`).
   - `room_name`: The name of the chat room to join (default: `main`).
   Example:
   ```bash
   ./chat -nick=John -room=general
   ```
2. The application will connect to other peers in the specified room and start exchanging messages.
## Logs
- All messages are logged in the `/messages` directory.
- Logs are stored as files named `<room_name>.log`.
## Example Workflow
1. Start the first instance of the application:
   ```bash
   ./chat -nick=Alice -room=developers
   ```
2. Start another instance of the application on the same network:
   ```bash
   ./chat -nick=Bob -room=developers
   ```
3. Alice and Bob can now exchange messages in the `developers` chat room.
## Notes
- Make sure the peers are on the same network for mDNS discovery to work.
- The application automatically creates the `/messages` directory and the necessary log files for each room.
## Future
- Add authentification mechanism. Where you can create private rooms.
- Transform to social network. Send not only messages, but posts and other.
## Contribution
Feel free to submit issues or pull requests for improvements or bug fixes. Contributions are always welcome!