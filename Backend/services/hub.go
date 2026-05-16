package services

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ishowsagar/go-blog-web-application/models"
	"gorm.io/gorm"
)

// client payload for post notificaton type struct
type ClientNotifyPayload struct {
	SenderID uint `json:"sender_id"`
	RecieverID uint `json:"reciever_id"`
	Type string `json:"type"`
	Content string `json:"content"`
	PostID uint `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Client struct {
	ID uint 
	Send chan *ClientNotifyPayload // Send chan which stores val of type ClientNotifyPayload
	Hub *Hub
	WebsocketConnection *websocket.Conn //* web socket connection
}

// centre point for all operations - all clients connection/disconnection,msg are redirected here and then handled
type Hub struct {
	Broadcast chan *ClientNotifyPayload // broadcasted msg struct that needs to be sent
	Clients map[*Client]bool // all clients stored in here
	ActiveClients chan *Client // chan whose have client val but only active clients will be redirected here
	DisconnectedClients chan *Client // chan for storing disconnected clients
	// DirectMessagesHub chan *models.DirectMessage
}


// initializes pointer instance of type Hub
func IntializeNewHubInstance() *Hub {
	return &Hub{
		Broadcast: make(chan *ClientNotifyPayload),
		Clients: make(map[*Client]bool),
		ActiveClients: make(chan *Client),
		DisconnectedClients: make(chan *Client),
		// DirectMessagesHub: make(chan *models.DirectMessage),
	}
}


// func that belongs to the type Hub
func(h *Hub) RunService() {
	
	// always running & keep checking which invokes and does its job in bg
	for {
		// infinite loop
		select {
			// select invoke case selected statements on chan's
		case incomingclient := <- h.ActiveClients :
			// if active client chan is reading val -> store in all clients
			h.Clients[incomingclient] = true // making it true -> saved it for ws conn
		
		case readdisconnectedclient := <- h.DisconnectedClients :
			// first validate if it exists or not with _,ok idiomatic way
			_,ok := h.Clients[readdisconnectedclient]
			if ok {
					// if disconnected remove from clients map [key is client itself]
					delete(h.Clients,readdisconnectedclient)
					close(readdisconnectedclient.Send) // stop furthur ingres on this chan's send chan which stores actual content

				}
			
		// if payload is redirected to chan <- which has client data
			case msg := <- h.Broadcast : // if something is redirected to this chan
				// then we check who send and where to and got -> redirect there to its send chan

				// by looping over each ranged client to check which matches -> redirects to its send chan which accepts same type of data struct 
				for client := range h.Clients {
					// sender sends and reader sends to broadcast chan 
					//& checking -> if there exists a client in the active clients whose id matches to whom msg is redirected for
					if client.ID == msg.RecieverID {
						// * redirecting msg to his send chan which recieves the msg in writer functio & writes to it by writeJson response
						client.Send<- msg // since both clients have both readers/writer activated, they keep checking and if something expected comes,write response
					
						// todo - now next step would be to diplay user name, not just id
					}
				}
			
		
		}
		
	}
}



// methods

// method that belongs to the type Client to read incoming data struct of type InboundMsg
func(c *Client) MessageReader(db *gorm.DB) {
//  then ,query to store msg into db, redirect msg output as Outbound to broadcast chan in hub

	// * defer this function invocation at the end
	defer func() {
		// unregistering the channel on which this meth is defined on - after read
		c.Hub.DisconnectedClients <- c // since hub is centre of all things, redirecting client that would be disconnected to the disconnectedCLient chan -> rest is done by hub method RunService
		c.WebsocketConnection.Close() //! close close as soon as *c is unregistered

		// @ chan methodology 
		// * var <- chan (reading from chan n storing onto var or reading onto var)
		// ! chan <- vat (redirecting output to chan)
	}()

	// loop --> infinite loop that keeps checking & recieving json payloads
	for {
		var msg ClientNotifyPayload // for every incoming msg
		
		// reading from the ws connection only, sent from there only

		// reading msg/eventData recieved on ws set path and if it recieves this type of msg -> process it for store and broadcasting

		// connection supports only one operation at the time either -> concurrent reader or writer 	

		// Writer and Reader method must not be executed concurrently, other than this writeControl and close can be called concurrent
		err := c.WebsocketConnection.ReadJSON(&msg) //* reading every recieved incoming msg --> if unmarshaled into struct n populated --> success
		if err != nil {
			slog.Error("failed to read request","error",err)
			break /// break loop from furthur calls
		}

		// * if recieving msg is correct of type inBMsg - making query to store in db
		err = db.Create(&models.Message{
			SenderID:msg.SenderID,
			RecieverID: msg.RecieverID,
			Content: msg.Content,
			CreatedAt: time.Now(),
		}).Error

		if err != nil {
			slog.Error("failed to load message","error",err)
			break
		}

		// * if successfully stored in db --> need to broadcast this message to reciever 👇
		
		c.Hub.Broadcast <- &ClientNotifyPayload {
			// which is recieved bny broadcast and send to that client 
			Type: msg.Type,
			PostID: msg.PostID,
			CreatedAt: msg.CreatedAt,
			SenderID: msg.SenderID,
			RecieverID: msg.RecieverID,
			Content : msg.Content,
		}

	}
}


// method that belongs to the type *Client which w-> checks for chan data coming in client's send chan -> write it to the ws writer
func(c *Client) MessageWriter() {

	defer c.WebsocketConnection.Close() // defer to close chan 

	// loop - looping over every Send chan data <- loop cause it continously keep checking if c.Send has outbound type data
	for msg := range c.Send {

		// writes current msg to the json
		err := c.WebsocketConnection.WriteJSON(msg) //* writes data to ws connection,where it is recieved as event.Data on the defined handler url path
		if err != nil {
			break
		}	
	} 
}



