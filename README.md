# Chatroom Application
## Overview
This is a decentralized (децентрализованное) chatroom application built with Golang and `go-libp2p-pubsub` library for peer-to-peer communication. Also you can join `AI bot` in your chat

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

6. If you want to add AI bot, then go to `cd ai/model` and run `python3.9 main.py`. After ypu can clikc on `Create Bot` and set bot nickname. After each message in chat he will answer you


### Features
- AI bot
- Decentralized chatroom communication using `go-libp2p-pubsub`.
- Front-end built with HTML, CSS, and JavaScript.
- User session management via cookies.
- Live updates for chatroom lists and messages.
- File uploading and showing.


## Future

### Technical
- Display last file messages. Now it will display only text messages.
- Add video files. Now you can not watch it, because we often refresh all current messages.
- Configure network for production. When users from different devices can chatting.
- Add authentification mechanism. Where you can create private rooms.
- Do fasthttp.
- Replace REST api calls between frontend and backend.

### Others
- Transform to social network. Send not only messages, but posts and other.
- Add monetization and own currency.