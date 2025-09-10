package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// ConnectionType represents the type of WebSocket connection
type ConnectionType string

const (
	ConnectionTypeVisitor ConnectionType = "visitor"
	ConnectionTypeAgent   ConnectionType = "agent"
	ConnectionTypeAiAgent ConnectionType = "ai-agent"
)

// Connection represents a WebSocket connection with metadata
type Connection struct {
	ID           string          `json:"id"`
	SessionID    uuid.UUID       `json:"session_id"`
	Type         ConnectionType  `json:"type"`
	AgentID      *uuid.UUID      `json:"user_id,omitempty"` // For agents
	ServerID     string          `json:"server_id"`         // Which server instance holds this connection
	ConnectedAt  time.Time       `json:"connected_at"`
	LastPingAt   time.Time       `json:"last_ping_at"`
	ProjectIDs   []uuid.UUID     `json:"project_ids"`
	WsConnection *websocket.Conn `json:"-"`
	writeMutex   sync.Mutex      `json:"-"` // Mutex to synchronize WebSocket writes
}

type DeliveryType string

const (
	Direct DeliveryType = "direct"
	Self   DeliveryType = "self"
)

// Message represents a chat message to be sent via WebSocket
type Message struct {
	Type         string          `json:"type"`
	SessionID    uuid.UUID       `json:"session_id"`
	Data         json.RawMessage `json:"data,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
	FromType     ConnectionType  `json:"from_type"`
	TenantID     *uuid.UUID      `json:"tenant_id,omitempty"`
	ProjectID    *uuid.UUID      `json:"project_id,omitempty"`
	AgentID      *uuid.UUID      `json:"agent_id,omitempty"`
	DeliveryType DeliveryType    `json:"delivery_type"`
}

// MessageHandler is a callback function for handling incoming messages
type MessageHandler func(connID string, message *Message)

// SessionMessageHandler is a callback function for handling session-wide messages
type SessionMessageHandler func(sessionID string, message *Message)

// ConnectionManager manages WebSocket connections using Redis only for enterprise scaling
type ConnectionManager struct {
	redis    redis.UniversalClient
	serverID string
	pubsub   *redis.PubSub
	ctx      context.Context
	cancel   context.CancelFunc

	// Message handlers
	messageHandler        MessageHandler
	sessionMessageHandler SessionMessageHandler

	// Local WebSocket connection registry for this server instance
	localConnections map[string]*Connection
	connMutex        sync.RWMutex

	// Configuration
	connectionTTL     time.Duration
	heartbeatInterval time.Duration
}

// NewConnectionManager creates a new Redis-only connection manager
func NewConnectionManager(redisClient redis.UniversalClient) *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())
	serverID := uuid.New().String() // Unique server instance ID

	cm := &ConnectionManager{
		redis:             redisClient,
		serverID:          serverID,
		ctx:               ctx,
		cancel:            cancel,
		localConnections:  make(map[string]*Connection),
		connectionTTL:     5 * time.Minute,
		heartbeatInterval: 30 * time.Second,
	}

	// Start background services
	go cm.startPubSubListener()

	log.Info().Str("server_id", serverID).Msg("Connection manager started")
	return cm
}

// AddConnection registers a new WebSocket connection in Redis only
func (cm *ConnectionManager) AddConnection(connType ConnectionType, sessionID uuid.UUID, projectIDs []uuid.UUID, agentID *uuid.UUID, conn *websocket.Conn) (string, error) {
	connID := uuid.New().String()
	// Store WebSocket connection locally for this server instance

	connection := &Connection{
		ID:           connID,
		SessionID:    sessionID,
		Type:         connType,
		AgentID:      agentID,
		WsConnection: conn,
		ProjectIDs:   projectIDs,
		ServerID:     cm.serverID,
		ConnectedAt:  time.Now(),
		LastPingAt:   time.Now(),
	}

	if conn != nil {
		cm.connMutex.Lock()
		cm.localConnections[connID] = connection
		cm.connMutex.Unlock()
	}

	if connType == ConnectionTypeVisitor {
		// Add to session-based lookup
		sessionKey := fmt.Sprintf("livechat:session:%s", sessionID)
		if err := cm.redis.SAdd(cm.ctx, sessionKey, connID).Err(); err != nil {
			log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to add connection to session set")
		}
		cm.redis.Expire(cm.ctx, sessionKey, cm.connectionTTL)
	} else if connType == ConnectionTypeAgent {
		// Add to agent-based lookup

		agentKey := fmt.Sprintf("livechat:agent:%s", agentID.String())
		fmt.Printf("Adding agent connection %s to agent %s\n", connID, agentKey)
		if err := cm.redis.SAdd(cm.ctx, agentKey, connID).Err(); err != nil {
			log.Error().Err(err).Str("agent_id", agentID.String()).Msg("Failed to add connection to agent set")
		}
		cm.redis.Expire(cm.ctx, agentKey, cm.connectionTTL)

		for _, projectID := range projectIDs {
			projectKey := fmt.Sprintf("livechat:project:%s", projectID.String())
			fmt.Printf("Adding project connection %s to project %s\n", connID, projectKey)
			if err := cm.redis.SAdd(cm.ctx, projectKey, connID).Err(); err != nil {
				log.Error().Err(err).Str("project_id", projectID.String()).Msg("Failed to add connection to project set")
			}
			cm.redis.Expire(cm.ctx, projectKey, cm.connectionTTL)
		}
	}

	// log.Info().
	// 	Str("connection_id", connID).
	// 	Str("session_id", sessionID.String()).
	// 	Str("type", string(connType)).
	// 	Msg("WebSocket connection added")

	return connID, nil
}

// RemoveConnection removes a WebSocket connection
func (cm *ConnectionManager) RemoveConnection(connID string) {
	// Remove from local connections first
	var connection *Connection
	cm.connMutex.Lock()
	if conn, exists := cm.localConnections[connID]; exists {
		conn.WsConnection.Close()
		connection = conn
		delete(cm.localConnections, connID)
	}
	cm.connMutex.Unlock()

	if connection == nil {
		log.Warn().Str("connection_id", connID).Msg("Connection not found for removal")
		return
	}

	if connection.Type == ConnectionTypeVisitor {
		// Add to session-based lookup
		sessionKey := fmt.Sprintf("livechat:session:%s", connection.SessionID)
		if err := cm.redis.SRem(cm.ctx, sessionKey, connID).Err(); err != nil {
			log.Error().Err(err).Str("session_id", connection.SessionID.String()).Msg("Failed to remove connection from session set")
		}
	} else if connection.Type == ConnectionTypeAgent {
		// Add to agent-based lookup
		agentKey := fmt.Sprintf("livechat:agent:%s", connection.AgentID.String())
		if err := cm.redis.SRem(cm.ctx, agentKey, connID).Err(); err != nil {
			log.Error().Err(err).Str("agent_id", connection.AgentID.String()).Msg("Failed to remove connection from agent set")
		}

		for _, projectID := range connection.ProjectIDs {
			projectKey := fmt.Sprintf("livechat:project:%s", projectID.String())
			if err := cm.redis.SRem(cm.ctx, projectKey, connID).Err(); err != nil {
				log.Error().Err(err).Str("project_id", projectID.String()).Msg("Failed to remove connection from project set")
			}
		}
	}

	// log.Info().
	// 	Str("connection_id", connID).
	// 	Str("session_id", connection.SessionID.String()).
	// 	Msg("WebSocket connection removed")
}

// SetMessageHandler sets the callback for handling direct connection messages
func (cm *ConnectionManager) SetMessageHandler(handler MessageHandler) {
	cm.messageHandler = handler
}

// SetSessionMessageHandler sets the callback for handling session-wide messages
func (cm *ConnectionManager) SetSessionMessageHandler(handler SessionMessageHandler) {
	cm.sessionMessageHandler = handler
}

// DeliverWebSocketMessage sends a message to all connections in a chat session
func (cm *ConnectionManager) DeliverWebSocketMessage(sessionID uuid.UUID, message *Message) error {
	message.Timestamp = time.Now()
	// Serialize message
	msgBytes, _ := json.Marshal(message)
	// Publish to Redis channel for cross-server delivery
	channelKey := "pubsub:livechat"
	if err := cm.redis.Publish(cm.ctx, channelKey, msgBytes).Err(); err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to publish message to Redis")
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// SendToConnection sends a message to a specific connection via Redis
func (cm *ConnectionManager) SendToConnection(connID string, message *Message) {
	// Since we don't store local connections, always publish to Redis for delivery
	message.Timestamp = time.Now()

	cm.connMutex.RLock()
	conn, exists := cm.localConnections[connID]
	cm.connMutex.RUnlock()

	if exists {
		message.SessionID = conn.SessionID

		// Use mutex to synchronize WebSocket writes and prevent concurrent write panic
		conn.writeMutex.Lock()
		err := conn.WsConnection.WriteJSON(message)
		conn.writeMutex.Unlock()

		if err != nil {
			log.Error().Err(err).Str("connection_id", connID).Msg("Failed to deliver session message to local connection")
			// Remove failed connection in background
			go cm.RemoveConnection(connID)
		} else {
			// log.Debug().Str("connection_id", connID).Str("session_id", message.SessionID.String()).Msg("Successfully delivered session message to local connection")
		}
	}
}

// SendToProjectAgents sends a message to all agents connected to a specific project
func (cm *ConnectionManager) SendToProjectAgents(projectID uuid.UUID, message *Message) error {
	// Set the message project context
	message.ProjectID = &projectID
	message.FromType = ConnectionTypeAgent
	message.Timestamp = time.Now()

	// Publish to the standard livechat channel - the handleRedisMessage will route it properly
	channelKey := "pubsub:livechat"
	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Str("project_id", projectID.String()).Msg("Failed to marshal message for project")
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := cm.redis.Publish(cm.ctx, channelKey, msgBytes).Err(); err != nil {
		log.Error().Err(err).Str("project_id", projectID.String()).Msg("Failed to publish message to project")
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Debug().Str("project_id", projectID.String()).Msg("Message published to project agents")
	return nil
}

// GetSessionConnections returns all connections for a session (across all servers)
func (cm *ConnectionManager) GetSessionConnections(sessionID string) ([]*Connection, error) {
	sessionKey := fmt.Sprintf("session:%s:connections", sessionID)
	connIDs, err := cm.redis.SMembers(cm.ctx, sessionKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session connections: %w", err)
	}

	var connections []*Connection
	for _, connID := range connIDs {
		connKey := fmt.Sprintf("connection:%s", connID)
		connData, err := cm.redis.Get(cm.ctx, connKey).Result()
		if err != nil {
			if err != redis.Nil {
				log.Error().Err(err).Str("connection_id", connID).Msg("Failed to get connection data")
			}
			continue
		}

		var connection Connection
		if err := json.Unmarshal([]byte(connData), &connection); err != nil {
			log.Error().Err(err).Str("connection_id", connID).Msg("Failed to unmarshal connection data")
			continue
		}

		connections = append(connections, &connection)
	}

	return connections, nil
}

// GetRedisClient returns the Redis client for external use
func (cm *ConnectionManager) GetRedisClient() redis.UniversalClient {
	return cm.redis
}

// GetServerID returns the unique server ID for this instance
func (cm *ConnectionManager) GetServerID() string {
	return cm.serverID
}

// Shutdown gracefully shuts down the connection manager
func (cm *ConnectionManager) Shutdown() {
	log.Info().Msg("Shutting down connection manager")

	// Cancel context to stop background services
	cm.cancel()

	// Close all local WebSocket connections
	cm.connMutex.Lock()
	for connID, conn := range cm.localConnections {
		conn.WsConnection.Close()
		delete(cm.localConnections, connID)
	}
	cm.connMutex.Unlock()

	// Close pub/sub
	if cm.pubsub != nil {
		cm.pubsub.Close()
	}

	// Remove all connections for this server from Redis
}

func (cm *ConnectionManager) startPubSubListener() {
	// Subscribe to project channels and connection-specific channels
	patterns := []string{
		"pubsub:livechat",
	}

	cm.pubsub = cm.redis.PSubscribe(cm.ctx, patterns...)
	defer cm.pubsub.Close()

	ch := cm.pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			cm.handleRedisMessage(msg)
		case <-cm.ctx.Done():
			return
		}
	}
}

func (cm *ConnectionManager) handleRedisMessage(msg *redis.Message) {
	var message Message
	if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal Redis message")
		return
	}

	cm.deliverSessionMessage(&message)
}

// deliverSessionMessage sends a message to all local connections in a session
func (cm *ConnectionManager) deliverSessionMessage(message *Message) {
	// Get all connections for this session from Redis
	projectID := message.ProjectID
	msgComingFrom := message.FromType
	sessionID := message.SessionID

	var connIDs []string
	var sessionKey string

	fmt.Println("Received message. from tenant:", message.TenantID)
	fmt.Println("Received message. from project:", message.ProjectID)
	fmt.Println("Received message. from session:", message.SessionID)
	fmt.Println("Received message. for agent:", message.AgentID)

	// Handle alarm messages specifically - they should go to all agents in the project
	if message.Type == "alarm_triggered" || message.Type == "alarm_acknowledged" || message.Type == "alarm_escalated" || message.Type == "agent_handoff_request" {
		if projectID != nil {
			sessionKey = fmt.Sprintf("livechat:project:%s", projectID.String())
		}
	} else if msgComingFrom == ConnectionTypeVisitor {
		// Handle visitor-specific logic
		if projectID != nil {
			sessionKey = fmt.Sprintf("livechat:project:%s", projectID.String())
		} else {
			// fallback to session if project is not available
			sessionKey = fmt.Sprintf("livechat:session:%s", sessionID.String())
		}

		// If an agent ID is present and not uuid.Nil, route to that specific agent
		if message.AgentID != nil && *message.AgentID != uuid.Nil {
			sessionKey = fmt.Sprintf("livechat:agent:%s", message.AgentID.String())
		}
	} else if msgComingFrom == ConnectionTypeAgent {
		// Handle agent-specific logic
		sessionKey = fmt.Sprintf("livechat:session:%s", sessionID.String())
	}

	fmt.Printf("Pubsub is working -> sessionKey: %s (message type: %s)\n", sessionKey, message.Type)

	connIDs, err := cm.redis.SMembers(cm.ctx, sessionKey).Result()
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to get session connections for delivery")
		return
	}

	fmt.Printf("Pubsub is working, for sessionKey: %s-> connIDs: %v\n for messageType: %s", sessionKey, connIDs, message.Type)

	// Send to local connections only (this server instance)
	cm.connMutex.RLock()
	defer cm.connMutex.RUnlock()

	deliveredCount := 0
	for _, connID := range connIDs {
		if conn, exists := cm.localConnections[connID]; exists {
			// For alarm messages, only send to agent connections
			if message.Type == "alarm_triggered" || message.Type == "alarm_acknowledged" || message.Type == "alarm_escalated" || message.Type == "agent_handoff_request" {
				if conn.Type != ConnectionTypeAgent {
					continue // Skip non-agent connections for alarm messages
				}
			}

			// Use mutex to synchronize WebSocket writes and prevent concurrent write panic
			conn.writeMutex.Lock()
			err := conn.WsConnection.WriteJSON(message)
			conn.writeMutex.Unlock()

			if err != nil {
				log.Error().Err(err).Str("connection_id", connID).Msg("Failed to deliver session message to local connection")
				// Remove failed connection in background
				go cm.RemoveConnection(connID)
			} else {
				deliveredCount++
				// log.Debug().Str("connection_id", connID).Str("session_id", sessionID.String()).Msg("Successfully delivered session message to local connection")
			}
		}
	}

	if message.Type == "alarm_triggered" || message.Type == "alarm_acknowledged" || message.Type == "alarm_escalated" {
		log.Debug().Str("message_type", message.Type).Int("delivered_count", deliveredCount).Msg("Alarm message delivered to agents")
	}
}

// UpdateConnectionPing updates the last ping time for a connection in Redis
func (cm *ConnectionManager) UpdateConnectionPing(connID string) {
	// Get connection from Redis and update ping time
	connKey := fmt.Sprintf("livechat:connection:%s", connID)
	connData, err := cm.redis.Get(cm.ctx, connKey).Result()
	if err != nil {
		if err != redis.Nil {
			log.Error().Err(err).Str("connection_id", connID).Msg("Failed to get connection for ping update")
		}
		return
	}

	var connection Connection
	if err := json.Unmarshal([]byte(connData), &connection); err != nil {
		log.Error().Err(err).Str("connection_id", connID).Msg("Failed to unmarshal connection for ping update")
		return
	}

	// Update ping time and store back to Redis
	connection.LastPingAt = time.Now()
	go cm.storeConnectionInRedis(&connection)
}

func (cm *ConnectionManager) storeConnectionInRedis(connection *Connection) error {
	// Don't store the WebSocket connection in Redis

	connData, err := json.Marshal(connection)
	if err != nil {
		return fmt.Errorf("failed to marshal connection: %w", err)
	}

	connKey := fmt.Sprintf("livechat:connection:%s", connection.ID)
	return cm.redis.Set(cm.ctx, connKey, connData, cm.connectionTTL).Err()
}

// GetConnection retrieves a connection by ID from Redis
func (cm *ConnectionManager) GetConnection(ctx context.Context, connID string) (*Connection, error) {
	// Get from Redis only
	connKey := fmt.Sprintf("connection:%s", connID)
	connData, err := cm.redis.Get(ctx, connKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("connection not found: %s", connID)
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	var connection Connection
	if err := json.Unmarshal([]byte(connData), &connection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal connection: %w", err)
	}

	return &connection, nil
}

// RedisInfo represents Redis server information for metrics
type RedisInfo struct {
	CommandsPerSecond      float64
	TotalCommandsProcessed int64
	UsedMemory             int64
	ConnectedClients       int64
}

// GetRedisInfo retrieves Redis server information for performance metrics
func (cm *ConnectionManager) GetRedisInfo(ctx context.Context) (*RedisInfo, error) {
	info, err := cm.redis.Info(ctx, "stats", "memory", "clients").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	redisInfo := &RedisInfo{}

	// Parse Redis INFO output - simplified version
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "instantaneous_ops_per_sec":
				if val, err := strconv.ParseFloat(value, 64); err == nil {
					redisInfo.CommandsPerSecond = val
				}
			case "total_commands_processed":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					redisInfo.TotalCommandsProcessed = val
				}
			case "used_memory":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					redisInfo.UsedMemory = val
				}
			case "connected_clients":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					redisInfo.ConnectedClients = val
				}
			}
		}
	}

	return redisInfo, nil
}
