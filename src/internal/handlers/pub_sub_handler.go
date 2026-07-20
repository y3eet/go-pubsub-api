package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"pubsub/internal/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Publisher struct {
	Topic   string `json:"topic"`
	Message any    `json:"message"`
}

type Subscriber struct {
	conn  *websocket.Conn
	Topic string
	send  chan []byte
}

type Hub struct {
	subscribers map[string]map[*Subscriber]struct{}
	publish     chan Publisher
	subscribe   chan *Subscriber
	unsubscribe chan *Subscriber
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string]map[*Subscriber]struct{}),
		publish:     make(chan Publisher),
		subscribe:   make(chan *Subscriber),
		unsubscribe: make(chan *Subscriber),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case pub := <-h.publish:
			msg := fmt.Appendf(nil, "%v", pub.Message)
			for sub := range h.subscribers[pub.Topic] {
				select {
				case sub.send <- msg:
				default:
					close(sub.send)
					delete(h.subscribers[pub.Topic], sub)
				}
			}
		case sub := <-h.subscribe:
			if h.subscribers[sub.Topic] == nil {
				h.subscribers[sub.Topic] = make(map[*Subscriber]struct{})
			}
			h.subscribers[sub.Topic][sub] = struct{}{}

		case sub := <-h.unsubscribe:
			fmt.Printf("Unsubscribing from topic: %s\n", sub.Topic)
			delete(h.subscribers[sub.Topic], sub)
			if len(h.subscribers[sub.Topic]) == 0 {
				delete(h.subscribers, sub.Topic) // optional cleanup
			}
			if err := sub.conn.Close(); err != nil {
				fmt.Printf("Error closing WebSocket connection for topic %s: %v\n", sub.Topic, err)
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) SubscribeHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		topic := c.Param("topic")

		if status, err := authCallback(c, topic, "subscribe"); err != nil {
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("Failed to set websocket upgrade: ", err)
			return
		}
		subscriber := &Subscriber{conn: conn, Topic: topic, send: make(chan []byte, 16)}
		go subscriber.writePump()

		hub.subscribe <- subscriber
		// Keep connection alive until client disconnects
		go func() {
			defer func() {
				hub.unsubscribe <- subscriber
			}()

			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	}
}

func (h *Handler) PublishHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {

		body := struct {
			Topic   string `json:"topic"`
			Message any    `json:"message"`
		}{}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		msgBytes, err := json.Marshal(body.Message)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal message"})
			return
		}
		if status, err := authCallback(c, body.Topic, "publish"); err != nil {
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		hub.publish <- Publisher{Topic: body.Topic, Message: string(msgBytes)}
		c.JSON(http.StatusOK, gin.H{"status": "Message published"})
	}
}

func authCallback(c *gin.Context, topic string, action string) (int, error) {
	payload := map[string]string{
		"topic":  topic,
		"action": action,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to marshal auth payload:", err)
		return http.StatusForbidden, err
	}

	req, err := http.NewRequest("POST", config.Cfg.AuthCallbackURL, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Failed to create auth request:", err)
		return req.Response.StatusCode, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Go-Pub-Sub-Key", config.Cfg.GoPubSubMasterKey)
	if auth := c.GetHeader("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	if cookie := c.GetHeader("Cookie"); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send auth request:", err)
		return resp.StatusCode, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, fmt.Errorf("auth request failed with status: %d", resp.StatusCode)
	}

	return http.StatusOK, nil
}

func (s *Subscriber) writePump() {
	defer s.conn.Close()
	for msg := range s.send {
		s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := s.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
	// channel was closed (e.g. by hub eviction) — send a close frame
	s.conn.WriteMessage(websocket.CloseMessage, []byte{})
}
