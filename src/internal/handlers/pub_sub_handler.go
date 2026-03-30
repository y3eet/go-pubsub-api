package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-pubsub-api/internal/config"
	"net/http"
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
}

type Hub struct {
	subscribers map[string][]*Subscriber
	publish     chan Publisher
	subscribe   chan *Subscriber
	unsubscribe chan *Subscriber
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string][]*Subscriber),
		publish:     make(chan Publisher),
		subscribe:   make(chan *Subscriber),
		unsubscribe: make(chan *Subscriber),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case pub := <-h.publish:
			for _, sub := range h.subscribers[pub.Topic] {
				sub.conn.WriteMessage(websocket.TextMessage, fmt.Appendf(nil, "%v", pub.Message))
			}
		case sub := <-h.subscribe:
			h.subscribers[sub.Topic] = append(h.subscribers[sub.Topic], sub)

		case sub := <-h.unsubscribe:
			fmt.Printf("Unsubscribing from topic: %s\n", sub.Topic)
			subs := h.subscribers[sub.Topic]
			for i, s := range subs {
				if s == sub {
					h.subscribers[sub.Topic] = append(subs[:i], subs[i+1:]...)
					break
				}
			}
			sub.conn.Close()
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

		if !authCallback(c, topic, "subscribe") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("Failed to set websocket upgrade: ", err)
			return
		}
		subscriber := &Subscriber{conn: conn, Topic: topic}

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
		if !authCallback(c, body.Topic, "publish") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		hub.publish <- Publisher{Topic: body.Topic, Message: string(msgBytes)}
		c.JSON(http.StatusOK, gin.H{"status": "Message published"})
	}
}

func authCallback(c *gin.Context, topic string, action string) bool {
	payload := map[string]string{
		"topic":  topic,
		"action": action,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to marshal auth payload:", err)
		return false
	}

	req, err := http.NewRequest("POST", config.Cfg.AuthCallbackURL, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Failed to create auth request:", err)
		return false
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
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
