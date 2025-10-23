# Socket Package

Package socket provides WebSocket functionality with room management, broadcasting, and real-time communication features.

## Features

- **WebSocket Hub**: Centralized hub for managing WebSocket connections
- **Room Management**: Join/leave rooms for group communication
- **Broadcasting**: Send messages to all clients, specific rooms, or users
- **Message Types**: Support for different message types and metadata
- **Thread Safety**: Thread-safe operations with mutex protection
- **Connection Management**: Automatic connection cleanup and error handling

## Usage

### Basic Setup

```go
import (
    "net/http"
    "anhnq/api-core/pkg/socket"
)

func main() {
    // Create a new hub
    hub := socket.NewHub()

    // Start the hub in a goroutine
    go hub.Run()

    // Set up WebSocket route
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        socket.ServeWS(hub, w, r)
    })

    // Start HTTP server
    http.ListenAndServe(":8080", nil)
}
```

### Broadcasting Messages

```go
// Broadcast to all clients
message := socket.Message{
    Type:      "notification",
    Data:      "Hello from server!",
    Timestamp: time.Now().Unix(),
}
hub.BroadcastToAll(message)

// Broadcast to specific room
message := socket.Message{
    Type:      "room_message",
    Data:      "Hello room!",
    Timestamp: time.Now().Unix(),
}
hub.BroadcastToRoom("room1", message)

// Send to specific user
message := socket.Message{
    Type:      "private_message",
    Data:      "Hello user!",
    Timestamp: time.Now().Unix(),
}
hub.BroadcastToUser("user123", message)
```

### Client-Side JavaScript

```javascript
const ws = new WebSocket("ws://localhost:8080/ws?user_id=user123");

ws.onopen = function () {
  console.log("Connected to WebSocket");

  // Join a room
  ws.send(
    JSON.stringify({
      type: "join_room",
      data: "room1",
    })
  );
};

ws.onmessage = function (event) {
  const message = JSON.parse(event.data);
  console.log("Received:", message);

  switch (message.type) {
    case "notification":
      showNotification(message.data);
      break;
    case "room_message":
      showRoomMessage(message.data);
      break;
    case "private_message":
      showPrivateMessage(message.data);
      break;
  }
};

ws.onclose = function () {
  console.log("Disconnected from WebSocket");
};

// Send a message to room
function sendRoomMessage(room, message) {
  ws.send(
    JSON.stringify({
      type: "room_message",
      data: {
        room: room,
        message: message,
      },
    })
  );
}
```

## Message Types

### Server to Client Messages

| Type              | Description              | Data Format           |
| ----------------- | ------------------------ | --------------------- |
| `notification`    | General notification     | String or Object      |
| `room_message`    | Message to specific room | Object with room info |
| `private_message` | Private message to user  | Object with user info |
| `system_message`  | System message           | String or Object      |

### Client to Server Messages

| Type           | Description      | Data Format                  |
| -------------- | ---------------- | ---------------------------- |
| `join_room`    | Join a room      | String (room name)           |
| `leave_room`   | Leave a room     | String (room name)           |
| `broadcast`    | Broadcast to all | String or Object             |
| `room_message` | Send to room     | Object with room and message |

## Hub Methods

### Connection Management

```go
// Get total number of clients
count := hub.GetClientCount()

// Get number of rooms
roomCount := hub.GetRoomCount()

// Get clients in a specific room
clients := hub.GetRoomClients("room1")
```

### Broadcasting

```go
// Broadcast to all clients
hub.BroadcastToAll(message)

// Broadcast to specific room
hub.BroadcastToRoom("room1", message)

// Send to specific user
hub.BroadcastToUser("user123", message)
```

## Message Structure

```go
type Message struct {
    Type      string                 `json:"type"`
    Data      interface{}            `json:"data"`
    Room      string                 `json:"room,omitempty"`
    UserID    string                 `json:"user_id,omitempty"`
    Timestamp int64                  `json:"timestamp"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
```

## Room Management

### Automatic Room Management

- Clients automatically join rooms when sending `join_room` messages
- Clients automatically leave rooms when sending `leave_room` messages
- Empty rooms are automatically cleaned up
- Clients are removed from all rooms when disconnected

### Manual Room Management

```go
// Join a client to a room
hub.JoinRoom(client, "room1")

// Remove a client from a room
hub.LeaveRoom(client, "room1")
```

## Error Handling

The WebSocket connection automatically handles:

- Connection errors
- Unexpected disconnections
- Invalid message formats
- Client cleanup on disconnect

## Integration with HTTP Handlers

```go
func setupWebSocketRoutes(r *chi.Router, hub *socket.Hub) {
    r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
        socket.ServeWS(hub, w, r)
    })

    // Example: Broadcast message via HTTP endpoint
    r.Post("/broadcast", func(w http.ResponseWriter, r *http.Request) {
        var message socket.Message
        if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
            http.Error(w, "Invalid message", http.StatusBadRequest)
            return
        }

        hub.BroadcastToAll(message)
        w.WriteHeader(http.StatusOK)
    })
}
```

## Authentication Integration

```go
func ServeWSWithAuth(hub *socket.Hub, w http.ResponseWriter, r *http.Request) {
    // Extract and validate JWT token
    token := r.URL.Query().Get("token")
    userID, err := validateJWTToken(token)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Set user ID in query params for ServeWS
    r.URL.RawQuery = "user_id=" + userID

    socket.ServeWS(hub, w, r)
}
```

## Best Practices

1. **Use Rooms**: Organize clients into rooms for better message targeting
2. **Handle Disconnections**: Always handle client disconnections gracefully
3. **Validate Messages**: Validate incoming messages from clients
4. **Rate Limiting**: Implement rate limiting for message sending
5. **Authentication**: Always authenticate WebSocket connections
6. **Error Handling**: Handle WebSocket errors appropriately
7. **Resource Cleanup**: Ensure proper cleanup of resources

## Example Use Cases

### Real-time Chat

```go
// Broadcast chat message to room
message := socket.Message{
    Type: "chat_message",
    Data: map[string]interface{}{
        "user":    "john_doe",
        "message": "Hello everyone!",
        "room":    "general",
    },
    Timestamp: time.Now().Unix(),
}
hub.BroadcastToRoom("general", message)
```

### Live Notifications

```go
// Send notification to specific user
message := socket.Message{
    Type: "notification",
    Data: map[string]interface{}{
        "title": "New Message",
        "body":  "You have a new message",
        "type":  "info",
    },
    UserID:    "user123",
    Timestamp: time.Now().Unix(),
}
hub.BroadcastToUser("user123", message)
```

### Live Updates

```go
// Broadcast system update to all clients
message := socket.Message{
    Type: "system_update",
    Data: map[string]interface{}{
        "version": "1.2.0",
        "message": "System updated successfully",
    },
    Timestamp: time.Now().Unix(),
}
hub.BroadcastToAll(message)
```
