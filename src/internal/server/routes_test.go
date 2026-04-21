package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"pubsub/internal/config"
	"pubsub/internal/handlers"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func setupAuthConfig(t *testing.T) func() {
	t.Helper()

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	config.Cfg = &config.Config{
		AuthCallbackURL:   authServer.URL,
		GoPubSubMasterKey: "test-master-key",
	}

	return func() {
		authServer.Close()
	}
}

func TestHelloWorldHandler(t *testing.T) {
	s := &Server{}
	r := gin.New()
	r.GET("/", s.HelloWorldHandler)
	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	// Serve the HTTP request
	r.ServeHTTP(rr, req)
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// Check the response body
	expected := "{\"message\":\"Hello World\"}"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestPublishHandler(t *testing.T) {
	cleanup := setupAuthConfig(t)
	defer cleanup()

	handler := handlers.NewHandler()
	hub := handlers.NewHub()
	go hub.Run()
	r := gin.New()

	r.POST("/publish", handler.PublishHandler(hub))

	reqBody := struct {
		Topic   string `json:"topic"`
		Message any    `json:"message"`
	}{
		Topic:   "test",
		Message: "test",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", "/publish", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	// Serve the HTTP request
	r.ServeHTTP(rr, req)
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "{\"status\":\"Message published\"}"

	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

}

func TestPublishDeliveredToSubscriber(t *testing.T) {
	cleanup := setupAuthConfig(t)
	defer cleanup()

	handler := handlers.NewHandler()
	hub := handlers.NewHub()
	go hub.Run()
	r := gin.New()

	r.POST("/publish", handler.PublishHandler(hub))
	r.GET("/subscribe/:topic", handler.SubscribeHandler(hub))

	ts := httptest.NewServer(r)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/subscribe/test"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer conn.Close()

	// Allow the subscribe handler to register the connection before publishing.
	time.Sleep(50 * time.Millisecond)

	reqBody := struct {
		Topic   string      `json:"topic"`
		Message interface{} `json:"message"`
	}{
		Topic: "test",
		Message: map[string]string{
			"hello": "world",
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(ts.URL+"/publish", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected publish status: got %d body %s", resp.StatusCode, string(respBody))
	}

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read websocket message: %v", err)
	}

	expected := `{"hello":"world"}`
	if string(msg) != expected {
		t.Fatalf("unexpected websocket payload: got %s want %s", string(msg), expected)
	}
}
