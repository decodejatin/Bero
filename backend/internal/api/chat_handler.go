package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// ChatHandler handles chat REST and WebSocket endpoints
type ChatHandler struct {
	chatService service.ChatService
	authService service.AuthService
	hub         *WebSocketHub
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService service.ChatService, authService service.AuthService) *ChatHandler {
	h := &ChatHandler{
		chatService: chatService,
		authService: authService,
		hub:         NewWebSocketHub(),
	}
	go h.hub.Run()
	return h
}

// ==================== REST Endpoints ====================

// CreateOrGetConversationRequest is the request body
type CreateOrGetConversationRequest struct {
	ParticipantID string  `json:"participant_id" validate:"required"`
	JobID         *string `json:"job_id,omitempty"`
}

// SendMessageRequest is the request body
type SendMessageRequest struct {
	Content     string `json:"content" validate:"required"`
	MessageType string `json:"message_type"`
}

// GetConversations returns all conversations for the authenticated user
func (h *ChatHandler) GetConversations(c echo.Context) error {
	userID := c.Get("user_id").(string)

	conversations, err := h.chatService.GetConversations(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get conversations"})
	}

	if conversations == nil {
		conversations = []service.ConversationResponse{}
	}

	return c.JSON(http.StatusOK, conversations)
}

// CreateOrGetConversation creates a new conversation or returns existing
func (h *ChatHandler) CreateOrGetConversation(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req CreateOrGetConversationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.ParticipantID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "participant_id is required"})
	}

	conv, err := h.chatService.GetOrCreateConversation(c.Request().Context(), userID, req.ParticipantID, req.JobID)
	if err != nil {
		if err == service.ErrSelfConversation {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "cannot create conversation with yourself"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create conversation"})
	}

	return c.JSON(http.StatusOK, conv)
}

// GetMessages returns paginated messages for a conversation
func (h *ChatHandler) GetMessages(c echo.Context) error {
	userID := c.Get("user_id").(string)
	conversationID := c.Param("id")

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if limit <= 0 {
		limit = 50
	}

	messages, err := h.chatService.GetMessages(c.Request().Context(), conversationID, userID, limit, offset)
	if err != nil {
		if err == service.ErrNotParticipant {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not a participant in this conversation"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get messages"})
	}

	return c.JSON(http.StatusOK, messages)
}

// SendMessage sends a message via REST (fallback for when WebSocket is unavailable)
func (h *ChatHandler) SendMessage(c echo.Context) error {
	userID := c.Get("user_id").(string)
	conversationID := c.Param("id")

	var req SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	msg, err := h.chatService.SendMessage(c.Request().Context(), conversationID, userID, req.Content, req.MessageType)
	if err != nil {
		if err == service.ErrNotParticipant {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not a participant in this conversation"})
		}
		if err == service.ErrEmptyMessage {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "message content cannot be empty"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to send message"})
	}

	// Broadcast to WebSocket clients
	h.hub.BroadcastToConversation(conversationID, userID, msg)

	return c.JSON(http.StatusCreated, msg)
}

// MarkAsRead marks all messages in a conversation as read
func (h *ChatHandler) MarkAsRead(c echo.Context) error {
	userID := c.Get("user_id").(string)
	conversationID := c.Param("id")

	if err := h.chatService.MarkAsRead(c.Request().Context(), conversationID, userID); err != nil {
		if err == service.ErrNotParticipant {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not a participant"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to mark as read"})
	}

	return c.JSON(http.StatusOK, SuccessResponse{Message: "marked as read"})
}

// GetUnreadCount returns the total unread message count
func (h *ChatHandler) GetUnreadCount(c echo.Context) error {
	userID := c.Get("user_id").(string)

	count, err := h.chatService.GetUnreadCount(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get unread count"})
	}

	return c.JSON(http.StatusOK, map[string]int64{"unread_count": count})
}

// ==================== WebSocket ====================

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// HandleWebSocket upgrades HTTP to WebSocket for real-time chat
func (h *ChatHandler) HandleWebSocket(c echo.Context) error {
	// Auth via query parameter since WebSocket can't use headers easily from mobile
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing token"})
	}

	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return err
	}

	client := &WebSocketClient{
		hub:    h.hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: claims.UserID,
	}

	h.hub.register <- client

	// Start read/write pumps
	go client.writePump()
	go client.readPump(h.chatService)

	return nil
}

// ==================== WebSocket Hub ====================

// WebSocketHub manages all active WebSocket connections
type WebSocketHub struct {
	clients    map[string]*WebSocketClient // userID -> client
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mu         sync.RWMutex
}

// NewWebSocketHub creates a new hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string]*WebSocketClient),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
	}
}

// Run processes hub events
func (hub *WebSocketHub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.mu.Lock()
			hub.clients[client.userID] = client
			hub.mu.Unlock()
			log.Printf("WebSocket client connected: %s", client.userID)

		case client := <-hub.unregister:
			hub.mu.Lock()
			if _, ok := hub.clients[client.userID]; ok {
				delete(hub.clients, client.userID)
				close(client.send)
			}
			hub.mu.Unlock()
			log.Printf("WebSocket client disconnected: %s", client.userID)
		}
	}
}

// BroadcastToConversation sends a message to conversation participants who are connected
func (hub *WebSocketHub) BroadcastToConversation(conversationID, senderID string, msg interface{}) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	payload := map[string]interface{}{
		"type":            "new_message",
		"conversation_id": conversationID,
		"message":         msg,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// Send only to the OTHER user in the conversation (not the sender)
	// Since conversations are between 2 people, just skip the sender
	for userID, client := range hub.clients {
		if userID != senderID {
			select {
			case client.send <- data:
			default:
				// Client buffer full, skip
			}
		}
	}
}

// SendToUser sends a message directly to a specific user
func (hub *WebSocketHub) SendToUser(userID string, data []byte) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	if client, ok := hub.clients[userID]; ok {
		select {
		case client.send <- data:
		default:
		}
	}
}

// ==================== WebSocket Client ====================

// WebSocketClient represents a single WebSocket connection
type WebSocketClient struct {
	hub    *WebSocketHub
	conn   *websocket.Conn
	send   chan []byte
	userID string
}

// WsMessage represents a WebSocket message from the client
type WsMessage struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
}

func (c *WebSocketClient) readPump(chatService service.ChatService) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WsMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case "message":
			msg, err := chatService.SendMessage(
				context.Background(),
				wsMsg.ConversationID,
				c.userID,
				wsMsg.Content,
				"text",
			)
			if err != nil {
				log.Printf("Failed to save WS message: %v", err)
				continue
			}

			// Broadcast to recipient
			c.hub.BroadcastToConversation(wsMsg.ConversationID, c.userID, msg)

			// Confirm to sender
			confirm := map[string]interface{}{
				"type":            "message_sent",
				"conversation_id": wsMsg.ConversationID,
				"message":         msg,
			}
			if data, err := json.Marshal(confirm); err == nil {
				select {
				case c.send <- data:
				default:
				}
			}

		case "typing":
			// Broadcast typing indicator
			typing := map[string]interface{}{
				"type":            "typing",
				"conversation_id": wsMsg.ConversationID,
				"user_id":         c.userID,
			}
			if data, err := json.Marshal(typing); err == nil {
				c.hub.mu.RLock()
				for userID, client := range c.hub.clients {
					if userID != c.userID {
						select {
						case client.send <- data:
						default:
						}
					}
				}
				c.hub.mu.RUnlock()
			}
		}
	}
}

func (c *WebSocketClient) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
