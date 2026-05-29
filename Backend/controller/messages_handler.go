package controller

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
)



type DirectMessageRecord struct {
	ID uint `json:"id"`
	SenderID uint `json:"sender_id"`
	RecieverID uint `json:"reciever_id"`
	SenderName string `json:"sender_name"`
	RecieverName string `json:"reciever_name"`
	Content string `json:"content"`
	CreatedAt string `json:"created_at"`
}

func (ws *WSController) LoadMessages(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrResponse{
			Ok: false,
			Status: "unauthorized user",
		})
		return
	}

	peerIDStr := c.Query("peer_id")
	query := `
		SELECT
			m.id,
			m.sender_id,
			m.reciever_id,
			COALESCE(s.name, '') AS sender_name,
			COALESCE(r.name, '') AS reciever_name,
			m.content,
			m.created_at
		FROM messages m
		LEFT JOIN users s ON s.id = m.sender_id
		LEFT JOIN users r ON r.id = m.reciever_id
		WHERE (m.sender_id = ? OR m.reciever_id = ?)
		ORDER BY m.created_at ASC
	`
	args := []interface{}{userID, userID}

	if peerIDStr != "" {
		peerID, err := strconv.ParseUint(peerIDStr, 10, 64)
		if err != nil || peerID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, utils.ErrResponse{
				Ok: false,
				Status: "invalid peer_id",
			})
			return
		}

		query = `
			SELECT
				m.id,
				m.sender_id,
				m.reciever_id,
				COALESCE(s.name, '') AS sender_name,
				COALESCE(r.name, '') AS reciever_name,
				m.content,
				m.created_at
			FROM messages m
			LEFT JOIN users s ON s.id = m.sender_id
			LEFT JOIN users r ON r.id = m.reciever_id
			WHERE (m.sender_id = ? AND m.reciever_id = ?) OR (m.sender_id = ? AND m.reciever_id = ?)
			ORDER BY m.created_at ASC
		`
		args = []interface{}{userID, peerID, peerID, userID}
	}

	var messages []DirectMessageRecord
	if err := ws.db.Raw(query, args...).Scan(&messages).Error; err != nil {
		slog.Error("failed to load direct messages", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.ErrResponse{
			Ok: false,
			Status: "failed to load messages",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Ok: true,
		Status: "messages loaded",
		Data: messages,
	})
}



//@ types

// controller type that stores -> messageDbModel which stores methods on it
type MessagesController struct {
	MessagesDbModel *services.MessagesDBModel
}

// func that returns instance of type -> MessagesController
func NewMessagesController(messagesDbModel *services.MessagesDBModel) *MessagesController {
	return &MessagesController{
		MessagesDbModel: messagesDbModel,
	}
}

func StoreDmMessages(c *gin.Context) {
	
	// validate req has userID attached to it by <= Auth middleware
	activeClientID := c.GetUint("user_id")
	if activeClientID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "login expired or invalid token",
		})
		return
	}

	// store dm
}
