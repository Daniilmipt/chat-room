# Chat Room Application
This is a decentralized chat-room console application built using the `go-libp2p` framework. It allows users to join chat rooms, send messages to other peers, and receive messages in real-time. The application logs all messages in the `/messages`

## Features
- Each launch create connection to `host` and `port`, which we pass in command line arguments.
- Decentralized communication using `go-libp2p-pubsub`.
- mDNS-based peer discovery for connecting with other users on the same network.
- Message logging for each chat room.

## Why we save messages in file
We use it as an auxiliary application for the main `chat-rooms` application in folder `/frontend`. So, we use files as a `database storage`. This files are used by `chat-rooms` application.


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

## Running the Application
1. Start the application by specifying the nickname, room name:
   ```bash
   ./chat -nick=<nickname> -room=<room_name> -host=<host> -port=<port>
   ```
   - `nickname`: The name to identify yourself in the chat (default: `anonymous`).
   - `room_name`: The name of the chat room to join (default: `main`).
   - `host`(Optional): The host where we listen p2p.
   -   `port`(Optional): The port where we listen p2p.

   Example:
   ```bash
   ./chat -nick=John -room=general -host=0.0.0.0 -port=8082
   ```
2. The application will connect to other peers in the specified room and start exchanging messages.

It not neccessary to pass host and port. By default it will be `0.0.0.0` and `0`. Zero port means random port.

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
