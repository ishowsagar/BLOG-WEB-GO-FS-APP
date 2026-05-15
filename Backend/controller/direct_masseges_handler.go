package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// handler method for ws connection which -> handles dm's
func (ws *WSController) HandleDMs(c *gin.Context) {
	// WebSocket upgrades cannot reliably use the Authorization header.
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrResponse{
			Ok:     false,
			Status: "token not found in query parameter",
		})
		return
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(ws.jwtSecret), nil
	})
	if err != nil || token == nil || !token.Valid {
		slog.Error("failed to parse dm token", "error", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrResponse{
			Ok:     false,
			Status: "invalid token",
		})
		return
	}

	tokenClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrResponse{
			Ok:     false,
			Status: "invalid token claims",
		})
		return
	}

	userIDFloat, ok := tokenClaims["user_id"].(float64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrResponse{
			Ok:     false,
			Status: "user_id not found in token",
		})
		return
	}

	userID := uint(userIDFloat)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrResponse{
			Ok:     false,
			Status: "invalid user id",
		})
		return
	}

	wsConnection, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("failed to open dm websocket", "error", err)
		c.AbortWithStatusJSON(http.StatusBadGateway, utils.ErrResponse{
			Ok:     false,
			Status: "failed to open webSocket connection for Dm's handler",
		})
		return
	}

	activeClient := &services.Client{
		ID:                  userID,
		Hub:                 ws.hub,
		Send:                make(chan *services.ClientNotifyPayload),
		WebsocketConnection: wsConnection,
	}

	ws.hub.ActiveClients <- activeClient
	go activeClient.MessageReader(ws.db)
	go activeClient.MessageWriter()
}
