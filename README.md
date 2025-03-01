# Chatroom Application
## Overview
This is a decentralized (децентрализованное) chatroom application built with Golang and `go-libp2p-pubsub` library for peer-to-peer communication.

## Example Workflow
1. Start the first instance of the application:
   ```bash
   ./chatroom
   ```
2. Login:

![alt text](./images/login.png)


3. Send message:
![alt text](./images/room.png)

4. Start another instance of the application on the same network and send message:
   ```bash
   ./chatroom
   ```
![alt text](./images/room_view.png)

5. If we click on `Chat List` we go on page: 

![alt text](./images/room_list.png)


### Features
- Decentralized chatroom communication using `go-libp2p-pubsub`.
- Front-end built with HTML, CSS, and JavaScript.
- User session management via cookies.
- Live updates for chatroom lists and messages.
- File upload and showing.


## Future
- Add authentification mechanism. Where you can create private rooms.
- Transform to social network. Send not only messages, but posts and other.
- Add monetization and own currency.