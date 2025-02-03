# Chat Room Application
This is a decentralized chat-room console application built using the `go-libp2p` framework. It allows users to join chat rooms, send messages to other peers, and receive messages in real-time. The application logs all messages in the `/messages`

## Features
- Run chat as `node`. In this case you run server instance locally and other users can connect to you by address `/ip4/<host>/tcp/9090/p2p/<peer_id>`. Exact address can be found out in log file `chatlog`.

- If you connect to ready node, you create connection to `host` and `port`, which we pass in command line arguments.

- Decentralized communication using `go-libp2p-pubsub`.

- mDNS-based peer discovery for connecting with other users on the same network.


## Why we save messages in file
We use it as an auxiliary application for the main `chat-rooms` application in folder `/frontend`. So, we use files as a `database storage`. This files are used by `chat-rooms` application.


## Launching Application
1. Clone repositorty on go to chat folder:
   ```bash
   git clone <repo_url>
   cd chat
   ```
3. Build application with taskfile:
   ```bash
   task build-chat
   ```

4. Run application:
   ```bash
   task run-chat-linux-node
   ```

## Running the Application
1. Start the application by specifying the nickname, room name:
   ```bash
   ./chat -nick=<nickname> -room=<roomname> -host=<host> -port=<port> -node
   ```
   - `nick`: The name to identify yourself in the chat (default: `anonymous`).
   - `room`: The name of the chat room to join (default: `main`).
   - `host` (Optional): The host where we listen p2p (default: `0.0.0.0`).
   -   `port` (Optional): The port where we listen p2p (default: `9090`).
   - `node` (Optional): Flag which says if we run in node mode (default: `false`)

   Example:
   ```bash
   ./chat -nick=John -room=general -host=0.0.0.0 -port=8082
   ```
2. The application will connect to other peers in the specified room and start exchanging messages.


## Logs
- All messages are logged in the `/messages` directory.
- Logs are stored as files named `<roomname>.log`.

## Example Workflow
1. Start the first instance of the application as node:
   ```bash
   ./chat -nick=Alice -room=developers -node
   ```
2. Start another instance of the application on the same network:
   ```bash
   ./chat -nick=Bob -room=developers -peerid=12D3Kootttttttttttttttttttttt
   ```
3. Alice and Bob can now exchange messages in the `developers` chat room.


## Future
- Add authentification mechanism. Where you can create private rooms.
- Transform to social network. Send not only messages, but posts and others.
