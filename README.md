# Chatroom Application

## Overview
This is a decentralized chatroom application built with Golang and leveraging the `chat` application which uses `go-libp2p-pubsub` library for peer-to-peer communication. The `server` package provides an HTTP interface for managing chatrooms, handling user sessions, and serving the front-end views for interacting with the chat system.

### Features
- Decentralized chatroom communication using `go-libp2p-pubsub`.
- Front-end built with HTML, CSS, and JavaScript.
- User session management via cookies.
- Live updates for chatroom lists and messages.

### Submoduls

- See more about application frontent in [frontend README](./frontend/README.md)
- See more about application backend (work with go-libp2p-pubsub) in [backend README](./chat/README.md)

## Future

- Add authentification mechanism. Where you can create private rooms.
- Transform to social network. Send not only messages, but posts and other.
- Add monetization and own currency