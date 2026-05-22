## Go Pub/Sub API

A simple Pub/Sub API built with Go, and WebSockets. It allows clients to subscribe to channels and receive real-time messages published to those channels.

### Quick Start

```bash
git clone https://github.com/y3eet/go-pubsub-api.git
cd go-pubsub-api/src &&
cp .env.example .env &&
make prepare &&
make watch
```

Dev ui is available at http://localhost:8080/ui

### API Endpoints

- `POST /publish/<topic>`: Publish a message to a channel.
- `GET /subscribe/<topic>`: Subscribe to a channel and receive messages in real-time via WebSockets.

### Environment Variables

- `PORT`: The port on which the server will run (default: 8080).
- `APP_ENV`: The application environment (default: "local").
- `GO_PUB_SUB_MASTER_KEY`: Master key for authentication (default: "your_master_key").
- `AUTH_CALLBACK_URL`: URL for authentication callback (default: `http://localhost:8080/auth/callback`).
  - Called every time a client tries to subscribe to a channel.
  - Server sends a POST request with the client's Cookies and Headers, Verify the client's credentials and the header X-Go-Pub-Sub-Key on your server.
  - Expected responses: `200` for authorized clients, `401` for unauthorized clients.
  - **WARNING:** Do not use the default callback URL in production, only use it for development purposes.

- `ALLOWED_ORIGINS`: Comma-separated list of allowed origins for CORS (default: "http://localhost:8080").
