package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gopi.com/api/http/dto"
	"gopi.com/internal/app/chat"
	"gopi.com/internal/app/user"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should check the origin
		return true
	},
}

// WebSocket connection manager
type WebSocketManager struct {
	clients     map[*websocket.Conn]string   // conn -> userID
	groups      map[string][]*websocket.Conn // groupSlug -> connections
	broadcast   chan []byte
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
	chatService *chat.ChatService
	userService *user.UserService
}

func NewWebSocketManager(chatService *chat.ChatService, userService *user.UserService) *WebSocketManager {
	return &WebSocketManager{
		clients:     make(map[*websocket.Conn]string),
		groups:      make(map[string][]*websocket.Conn),
		broadcast:   make(chan []byte),
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
		chatService: chatService,
		userService: userService,
	}
}

func (manager *WebSocketManager) Run() {
	for {
		select {
		case <-manager.register:
			// Connection registered, but user not authenticated yet
			log.Println("WebSocket connection registered")

		case conn := <-manager.unregister:
			// Remove connection
			if userID, exists := manager.clients[conn]; exists {
				delete(manager.clients, conn)
				// Remove from all groups
				for groupSlug, connections := range manager.groups {
					for i, c := range connections {
						if c == conn {
							manager.groups[groupSlug] = append(connections[:i], connections[i+1:]...)
							break
						}
					}
				}
				log.Printf("User %s disconnected", userID)
			}
			conn.Close()

		case message := <-manager.broadcast:
			// Broadcast message to all connected clients
			for conn := range manager.clients {
				select {
				case <-time.After(time.Second * 1):
					conn.Close()
					delete(manager.clients, conn)
				default:
					if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
						conn.Close()
						delete(manager.clients, conn)
					}
				}
			}
		}
	}
}

func (manager *WebSocketManager) AuthenticateConnection(conn *websocket.Conn, userID string) {
	manager.clients[conn] = userID
	log.Printf("User %s authenticated", userID)
}

func (manager *WebSocketManager) JoinGroup(conn *websocket.Conn, groupSlug string) {
	manager.groups[groupSlug] = append(manager.groups[groupSlug], conn)
	log.Printf("Connection joined group %s", groupSlug)
}

func (manager *WebSocketManager) LeaveGroup(conn *websocket.Conn, groupSlug string) {
	if connections, exists := manager.groups[groupSlug]; exists {
		for i, c := range connections {
			if c == conn {
				manager.groups[groupSlug] = append(connections[:i], connections[i+1:]...)
				break
			}
		}
	}
	log.Printf("Connection left group %s", groupSlug)
}

func (manager *WebSocketManager) BroadcastToGroup(groupSlug string, message []byte) {
	if connections, exists := manager.groups[groupSlug]; exists {
		for _, conn := range connections {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				conn.Close()
				delete(manager.clients, conn)
			}
		}
	}
}

// WebSocket handler
type ChatWebSocketHandler struct {
	manager *WebSocketManager
}

func NewChatWebSocketHandler(manager *WebSocketManager) *ChatWebSocketHandler {
	return &ChatWebSocketHandler{
		manager: manager,
	}
}

// HandleWebSocket handles WebSocket connections for chat
// HandleWebSocket godoc
// @Summary WebSocket chat connection
// @Description Upgrade the HTTP connection to a WebSocket for realtime chat within a group. Requires JWT auth. Alias route: /ws/chat/group/{groupSlug}.
// @Tags chat
// @Security BearerAuth
// @Param groupSlug path string true "Group slug"
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Group not found"
// @Router /ws/chat/groups/{groupSlug} [get]
func (h *ChatWebSocketHandler) HandleWebSocket(c *gin.Context) {
	groupSlug := c.Param("groupSlug")
	userID := c.GetString("user_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if user is a member of the group
	group, err := h.manager.chatService.GetGroupBySlug(groupSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	isMember := false
	for _, memberID := range group.MemberIDs {
		if memberID == userID {
			isMember = true
			break
		}
	}

	if !isMember && group.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You must be a member of this group"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Register connection
	h.manager.register <- conn
	h.manager.AuthenticateConnection(conn, userID)
	h.manager.JoinGroup(conn, groupSlug)

	// Send connected message
	conn.WriteJSON(dto.WebSocketMessage{
		Type: "connected",
	})

	// Handle messages
	for {
		var wsMessage dto.WebSocketMessage
		err := conn.ReadJSON(&wsMessage)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			h.manager.unregister <- conn
			break
		}

		// Handle different message types
		switch wsMessage.Type {
		case "chat_message":
			h.handleChatMessage(conn, groupSlug, userID, wsMessage)
		case "typing":
			h.handleTypingMessage(groupSlug, userID, wsMessage)
		case "stop_typing":
			h.handleStopTypingMessage(groupSlug, userID, wsMessage)
		}
	}
}

func (h *ChatWebSocketHandler) handleChatMessage(conn *websocket.Conn, groupSlug, userID string, wsMessage dto.WebSocketMessage) {
	// Get user info
	user, err := h.manager.userService.GetUserByID(userID)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		return
	}

	// Resolve group to get its ID
	group, err := h.manager.chatService.GetGroupBySlug(groupSlug)
	if err != nil {
		log.Printf("Failed to get group by slug: %v", err)
		conn.WriteJSON(dto.WebSocketMessage{
			Type:    "error",
			Message: "Group not found",
		})
		return
	}

	// Save message to database
	_, err = h.manager.chatService.SendMessage(userID, group.ID, wsMessage.Message)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
		conn.WriteJSON(dto.WebSocketMessage{
			Type:    "error",
			Message: "Failed to send message",
		})
		return
	}

	// Prepare broadcast message
	broadcastMessage := dto.WebSocketMessage{
		Type:            "chat_message",
		Message:         wsMessage.Message,
		Username:        user.Username,
		UserID:          userID,
		GroupSlug:       groupSlug,
		UserImage:       user.ProfileImageURL,
		IsAuthenticated: true,
	}

	// Broadcast to group
	messageBytes, _ := json.Marshal(broadcastMessage)
	h.manager.BroadcastToGroup(groupSlug, messageBytes)
}

func (h *ChatWebSocketHandler) handleTypingMessage(groupSlug, userID string, wsMessage dto.WebSocketMessage) {
	// Get user info
	user, _ := h.manager.userService.GetUserByID(userID)

	// Prepare typing message
	typingMessage := dto.WebSocketMessage{
		Type:      "typing",
		Username:  user.Username,
		UserID:    userID,
		GroupSlug: groupSlug,
		UserImage: user.ProfileImageURL,
	}

	// Broadcast to group
	messageBytes, _ := json.Marshal(typingMessage)
	h.manager.BroadcastToGroup(groupSlug, messageBytes)
}

func (h *ChatWebSocketHandler) handleStopTypingMessage(groupSlug, userID string, wsMessage dto.WebSocketMessage) {
	// Get user info
	user, _ := h.manager.userService.GetUserByID(userID)

	// Prepare stop typing message
	stopTypingMessage := dto.WebSocketMessage{
		Type:      "stop_typing",
		Username:  user.Username,
		UserID:    userID,
		GroupSlug: groupSlug,
		UserImage: user.ProfileImageURL,
	}

	// Broadcast to group
	messageBytes, _ := json.Marshal(stopTypingMessage)
	h.manager.BroadcastToGroup(groupSlug, messageBytes)
}
