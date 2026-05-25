package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
	"github.com/ishowsagar/go-blog-web-application/events"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
	"gorm.io/gorm"
)

type WSController struct {
	db *gorm.DB
	hub *services.Hub
	jwtSecret string
	rabbitBroker *events.PubSubBroker
}

// returns instance of type WsController which -> stores method for handeling ws handlers
func NewWsController(db *gorm.DB,hub *services.Hub,jwtSecret string,rabbitBroker *events.PubSubBroker) *WSController {
	return &WSController{
		db: db,
		hub : hub,
		jwtSecret: jwtSecret,
		rabbitBroker: rabbitBroker,
	}
}

// upgrade connection from http to web
type connectionUpgrader struct {
	webSocketUpgrader websocket.Upgrader
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request)bool {
		return true
	},
}

// returns connection upgrader type which has ws upgrader
func NewUpgrader() *connectionUpgrader {
	upgrader := websocket.Upgrader{
		// set CheckOrigin field -> return true if upgraded
		CheckOrigin: func(r *http.Request) bool {
			return true // if it returns true then -> the connection is successfully migrated to webSocket conn
		},
	}
	return &connectionUpgrader{
		webSocketUpgrader: upgrader,
	}
}



// registering routes for realtime notifications
func(ws *WSController) ServeRealtimeNotification(c *gin.Context) {

	// Get token from query parameter (WebSocket doesn't support Authorization header in upgrade request)
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "token not found in query parameter",
		})
		return
	}

	// Parse and validate token
	token,err := jwt.Parse(tokenStr,func(t *jwt.Token) (interface{}, error) {
		return []byte(ws.jwtSecret),nil
	})
	if err != nil {
		slog.Error("failed to parse token","error",err)
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "invalid token",
		})
		return
	}
	
	if token == nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "invalid or expired token",
		})
		return
	}

	// Extract user ID from token claims
	tokenClaims,ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "invalid token claims",
		})
		return
	}

	userIDFloat, ok := tokenClaims["user_id"].(float64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "user_id not found in token",
		})
		return
	}

	userID := uint(userIDFloat)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "invalid user id",
		})
		return
	}

	//*http request to ws

	// upgrade connection - upgrader connection needs a handler method to migrate it, need to pass handler's req and res writer, if upgrader's check origin returns true -> upgraded to the webSocket conn
	wsConn,err := upgrader.Upgrade(c.Writer,c.Request,nil)
	if err != nil{
		slog.Error("failed to upgrade connection","error",err)
		c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{	
			Ok: false,
			Status: "failed to upgrade connection",
		})
		return
	}

	slog.Info("websockets connection started for the active client✅")

	// * active client is created

	// active client is initialized and stored in  hub's clients []
	client := &services.Client{
		ID: userID,
		WebsocketConnection: wsConn,
		Hub: ws.hub,
		Send: make(chan *services.ClientNotifyPayload, 100),
		BroadcastStatus: make(chan *services.StatusPayload, 50), // buffered so status updates queue up even if writer is temporarily busy
	}

	// set optional disconnect callback so controller can unbind user from broker
	client.OnDisconnect = func(id uint) {
		if ws.rabbitBroker != nil {
			if err := ws.rabbitBroker.UnbindUserFromExchange(id); err != nil {
				slog.Error("failed to unbind user in rabbit broker on disconnect","error",err)
			}
		}
	}

	slog.Info("Client stored in active clients successfully⚡","clientID -",client.ID)
	
	//* adding rabbitMQ broker after creating client to -> to bind user in the "exchange-noti..."
	err = ws.rabbitBroker.BindUserToTheExchange(client.ID) // & bind to broker exchange when client is created and also unnbind when client gets disconnected
	if err != nil {
		slog.Error("failed to bib=nd user in the rabbitMQ broker's declared exchange","error",err)
	}

	// register it as active client -> redirect client there
	ws.hub.ActiveClients <- client
	ws.hub.Online <- client

	// * reading and sending responses

	// fire both writer & reader go routines
	// reader uses decoder and writer uses encoder in underlying technology

	// frontend sends data -> read by reader -> sends to broadcast chan -> which checks target from active clients -> sends to that client and done by writing to it
	go client.MessageReader(ws.db) // sends the recieved payload of type notifyPostNoti to broadcast chan -> braodcast chan shares to client send where it is wriiten to ws conn client
	go client.MessageWriter() // sends response to that reciever active client -> sends response to his frontend ws connection
	// return
	
}