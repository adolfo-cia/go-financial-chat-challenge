# go-financial-chat-challenge
A simple multi chat-room, where participants can request stock values through the chat command ``/stock=appl.us`` where ``appl.us``is the stock ticker.

The arquitecture consists on 4 docker containers running with the following services:

- The chat.
- A simple bot that listens to a queue, hits an API to get the stock value and return it through the other queue.
- RabbitMQ to support the two queues that provides communication between the services.
- PostgreSQL for user management.

## Try it with Docker
#### Remove existing containers and images with:
- ``docker compose down -v --rmi all``
#### Start:
- ``docker compose up``

#### Create User:
- ``POST localhost:8080/users``
```
{
    "username": "a username",
    "password": "a password"
}
```
#### Login:
- ``POST localhost:8080/login``
```
{
    "username": "a username",
    "password": "a password"
}
```
Get the access token from the response and use it in the Authorization header like ``Bearer <accessToken>``

#### Create a chat room:
- ``POST localhost:8080/rooms``
```
{
    "title": "a title"
}
```
Get the roomId from the response and use it as a query param to join a room

#### Join a chat room (websocket):
- ``GET ws://localhost:8080/ws?roomId=307027e6-6768-4e2d-a9c2-3ff8bf5dcc0e``

In the payload you send messages with:
```
{
    "msg": "Hello World!"
}
```

To trigger the bot to fetch a stock value, you send a message like:
```
{
    "msg": "/stock=appl.us"
}
```

