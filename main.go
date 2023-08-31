package main

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	// Use a Mutex to manage concurrent access to rooms
	roomMutex sync.Mutex
	rooms     = make(map[string]map[socketio.Conn]struct{})
)

func main() {
	lErr := godotenv.Load()
	if lErr != nil {
		log.Fatal(lErr)
	}

	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		log.Println("Client connected:", s.ID())
		return nil
	})

	//server.OnConnect("/", handleConnect)
	server.OnEvent("/", "chat message", handleChatMessage)
	server.OnDisconnect("/", handleDisconnect)

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.Handle("/socket.io/", server)

	port := os.Getenv("PORT")
	log.Printf("Server started on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleConnect(s socketio.Conn) error {
	log.Printf("Client connected: %s", s.ID())

	// You can implement authentication and authorization here
	// For example, validate tokens or user credentials

	return nil
}

func handleChatMessage(s socketio.Conn, msg string) {
	msg = sanitizeMessage(msg)

	log.Printf("Received message from %s: %s", s.ID(), msg)

	// Get the room associated with the socket
	room := s.Context().(string)

	// Broadcast the message to all clients in the room
	roomMutex.Lock()
	defer roomMutex.Unlock()
	if clients, exists := rooms[room]; exists {
		for client := range clients {
			client.Emit("chat message", fmt.Sprintf("%s says: %s", s.ID(), msg))
		}
	}
}

func handleDisconnect(s socketio.Conn, reason string) {
	log.Printf("Client disconnected: %s (Reason: %s)", s.ID(), reason)

	// Remove the socket from the associated room
	roomMutex.Lock()
	defer roomMutex.Unlock()
	room := s.Context().(string)
	if clients, exists := rooms[room]; exists {
		delete(clients, s)
		if len(clients) == 0 {
			delete(rooms, room)
		}
	}
}

func sanitizeMessage(msg string) string {
	// Remove leading and trailing whitespace
	msg = strings.TrimSpace(msg)

	// Limit message length
	if len(msg) > 200 {
		msg = msg[:200]
	}

	// Remove HTML tags
	msg = stripHTMLTags(msg)

	return msg
}

func stripHTMLTags(input string) string {
	// Regular expression to match HTML tags
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(input, "")
}
