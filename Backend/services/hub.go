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
	// $ since we can set this as central payload for notifications, so we can as many fields it will be conditonal checked on frontend side to check what is coming and deal with it
	SenderID uint `json:"sender_id"`
	RecieverID uint `json:"reciever_id"`
	SenderName string `json:"sender_name"`
	RecieverName string `json:"reciever_name"`
	RoomID uint `json:"room_id"`
	RoomStatus bool `json:"room_status"` // clients sends payload with bool true -> joined, false -> left
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
	// testing - adding more chan to recieve and send more data
	BroadcastStatus chan *StatusPayload // for tracking status of client activity
	OnDisconnect func(userID uint)

}

//@ Interface that stores method -> which belongs to the hub which -> when called calls publisher method
type HubBroker interface {
	PublishEvents(userID uint, payload *ClientNotifyPayload) error
}

// centre point for all operations - all clients connection/disconnection,msg are redirected here and then handled
type Hub struct {
	Broadcast chan *ClientNotifyPayload // broadcasted msg struct that needs to be sent
	Clients map[*Client]bool // all clients stored in here
	// * keep active clients in same way clients but for key being clientID
	ClientStore map[uint] *Client // mapped data with key being clientID and stores that client
	ActiveClients chan *Client // chan whose have client val but only active clients will be redirected here
	DisconnectedClients chan *Client // chan for storing disconnected clients
	Online  chan *Client //chan for tracking if client is active or not
	Offline  chan *Client //chan for tracking if client is offine or not
	BrokerInterface HubBroker // optional broker for publishing delivery across instances
	TargettedBrokerMessages chan *ClientNotifyPayload // chan that stores only targetted User message
	ChatRoomClients  map[uint]map[*Client]bool // for rooms -> map where //&[keyBeingRoomID]  //*valueBeing slice[]*els are Client [roomID : [clien1,client2]]
	RoomClientsPayloads chan *ClientNotifyPayload 
	RegisterRoomClient chan *Client // recieves client for room
	TargettedClientNotificationTypeOnly chan *ClientNotifyPayload
	// DirectMessagesHub chan *models.DirectMessage
}

// for status broadcasting to all other clients
type StatusPayload struct {
	UserID uint `json:"user_id"`
	Status string `json:"status"`
}


// initializes pointer instance of type Hub
func IntializeNewHubInstance() *Hub {
	return &Hub{
		Broadcast: make(chan *ClientNotifyPayload, 20),
		Clients: make(map[*Client]bool),
		ActiveClients: make(chan *Client),
		DisconnectedClients: make(chan *Client),
		Online: make(chan *Client),
		Offline: make(chan *Client),
		TargettedBrokerMessages: make(chan *ClientNotifyPayload, 20),
		RoomClientsPayloads: make(chan *ClientNotifyPayload, 20),
		RegisterRoomClient: make(chan *Client),
		ChatRoomClients: make(map[uint]map[*Client]bool),
		ClientStore: make(map[uint]*Client),
		// DirectMessagesHub: make(chan *models.DirectMessage),
	}
}

// assigns the broker in the hub so -> that method that belongs to hub can use it
func (h *Hub) SetBroker(b HubBroker) {
	h.BrokerInterface = b
}

// method that belongs to hub which -> calls hub.BrokerInterface's publisher method to do the task
func (h *Hub) Publish(userID uint, payload *ClientNotifyPayload) {
	if h.BrokerInterface != nil {
		if err := h.BrokerInterface.PublishEvents(userID, payload); err != nil {
			slog.Error("failed to publish message via broker", "error", err, "user_id", userID)
		}
	}
}


// func that belongs to the type Hub
func(h *Hub) RunService() {
	slog.Info("HUB RunService started")
	slog.Info("HUB channels initialized", "broadcast_cap", cap(h.Broadcast), "targeted_cap", cap(h.TargettedBrokerMessages))
	
	// always running & keep checking which invokes and does its job in bg
	for {
		// infinite loop
		select {
			// select invoke case selected statements on chan's
		case incomingclient := <- h.ActiveClients :
			slog.Debug("HUB: ActiveClients case triggered")
			// if active client chan is reading val -> store in all clients
			h.Clients[incomingclient] = true // making it true -> saved it for ws conn
			// * also setting that client in the ALLClients type store -> which tracks incoming client but with userID
			h.ClientStore[incomingclient.ID] = incomingclient  //map key being client's UserID and val being client itself

		case readdisconnectedclient := <- h.DisconnectedClients :
			// first validate if it exists or not with _,ok idiomatic way
			_,ok := h.Clients[readdisconnectedclient]
			if ok {
					// if disconnected remove from clients map [key is client itself]
					delete(h.Clients,readdisconnectedclient)
				}
				
			// * deleting from client store too
			_,exists :=h.ClientStore[readdisconnectedclient.ID]
			if exists {
				// if exists delete and close connection
				delete(h.ClientStore,readdisconnectedclient.ID) // passing ID as id is the key for that key-var to be removed
				slog.Info("client gracefully shutdowned from system","userID",readdisconnectedclient.ID)
			}
				
			//&deleting client from the room too
			 //need to provide room id to delete from them only
			// since we need room id we do deleting there only 
			// delete him from every room map, then remove empty rooms
			for roomID, roomClients := range h.ChatRoomClients {
				delete(roomClients, readdisconnectedclient)
				if len(roomClients) == 0 {
					delete(h.ChatRoomClients, roomID)
				}
			}
			slog.Info("client removed from all rooms on disconnect", "client_id", readdisconnectedclient.ID)


		
			 // if payload is redirected to chan <- which has client data
			case msg := <- h.Broadcast : // if something is redirected to this chan
				// then we check who send and where to and got -> redirect there to its send chan

				// if no id provided broadcast to all clients 
				if  msg.RecieverID  == 0 {
					for eachClient := range h.Clients {
						// send in a goroutine with timeout to avoid blocking the hub
						go func(cl *Client, payload *ClientNotifyPayload) {
							select {
							case cl.Send <- payload:
								// sent
							case <-time.After(time.Second):
								slog.Warn("HUB broadcast timeout sending to client", "client_id", cl.ID)
							}
						}(eachClient, msg)
					}
				} else {
					// else only send to the targetted user
					// by looping over each ranged client to check which matches -> redirects to its send chan which accepts same type of data struct 
					for client := range h.Clients {
						if client.ID == msg.RecieverID {
							go func(cl *Client, payload *ClientNotifyPayload) {
								select {
								case cl.Send <- payload:
								case <-time.After(time.Second):
									slog.Warn("HUB targeted broadcast timeout", "client_id", cl.ID)
								}
							}(client, msg)
						}
					}
				}
			
			// & broker case -> when payload comes from consumer
			case targttedMsg := <- h.TargettedBrokerMessages :
			slog.Info("HUB received TargettedBrokerMessages", "receiver", targttedMsg.RecieverID, "sender", targttedMsg.SenderID, "clients_count", len(h.Clients))
			// check for target client and redirect to its ws writer for response
			found := false
			sentCount := 0
			seen := make(map[uint]bool) // dedupe by user ID in case multiple *Client entries exist for same user
			slog.Debug("HUB looping through clients to find target", "target_id", targttedMsg.RecieverID)
			for currentActiveClient := range h.Clients {
				// skip duplicate client IDs
				if seen[currentActiveClient.ID] {
					continue
				}
				// do not echo the message back to the sender
				if currentActiveClient.ID == targttedMsg.SenderID {
					continue
				}
				if currentActiveClient.ID == targttedMsg.RecieverID {
					seen[currentActiveClient.ID] = true
					found = true
					slog.Info("HUB found target client, sending to Send channel", "client_id", currentActiveClient.ID)
					go func(cl *Client, payload *ClientNotifyPayload) {
						select {
						case cl.Send <- payload:
							slog.Info("HUB message sent to client.Send")
							// incrementing sentCount in goroutine not safe for this loop; it's okay to log only
						case <-time.After(time.Second):
							slog.Warn("HUB failed to route targeted message (timeout)", "client_id", cl.ID)
						}
					}(currentActiveClient, targttedMsg)
				}
			}
			if !found {
				slog.Error("HUB target client not found", "target_id", targttedMsg.RecieverID, "active_clients", len(h.Clients))
			} else {
				slog.Debug("HUB targeted delivery summary", "sent_count", sentCount)
			}
			
			
			//@ Recieves room client and set them in desired room
			case client := <- h.RegisterRoomClient :
				// set in the room
			slog.Info("client is ready to hop into the room...","ClientID :",client.ID)
				
			
			// & Rooms - if being able to read payload from hub's roomClienttPayload chan
			case roomMsgPayload := <-h.RoomClientsPayloads:
				slog.Info("roomMsg payload received in hub RoomClientsPayloads", "room_id", roomMsgPayload.RoomID, "sender_id", roomMsgPayload.SenderID, "room_status", roomMsgPayload.RoomStatus)


				// bug - payload is successfully recieved, only thing it is doing wrong is sending only to other and itself is excluded
				// for ex => if payload has room.1 and two clients are active 1,41, if payload coming from 41, will goes to 1 but not himself,like if 1 sends room msg, it will sent to room but exluding him

				// ensure inner map exists when joining
				roomClients, roomExists := h.ChatRoomClients[roomMsgPayload.RoomID]
				if roomMsgPayload.RoomStatus {
					// if first roomStatus in payload is true then add it to the activeRoom clients mapped to that roomID's room only
					if !roomExists {
						roomClients = make(map[*Client]bool)
						h.ChatRoomClients[roomMsgPayload.RoomID] = roomClients
						slog.Info("HUB created room map", "room_id", roomMsgPayload.RoomID)
					}

					// add sender if active
					var senderClient *Client
					for ac := range h.Clients {
						if ac.ID == roomMsgPayload.SenderID {
							senderClient = ac
							break
						}
					}
					if senderClient == nil {
						slog.Warn("HUB room join: sender not active", "room_id", roomMsgPayload.RoomID, "sender_id", roomMsgPayload.SenderID)
						continue
					}
					roomClients[senderClient] = true //* stores sender
					slog.Info("HUB: added sender to room", "room_id", roomMsgPayload.RoomID, "client_id", senderClient.ID)

					// optionally add receiver when provided
					if roomMsgPayload.RecieverID != 0 && roomMsgPayload.RecieverID != roomMsgPayload.SenderID {
						for ac := range h.Clients {
							if ac.ID == roomMsgPayload.RecieverID {
								roomClients[ac] = true //* stores reciever //! btw why we need reciever we can flow all in room
								slog.Info("HUB: added receiver to room", "room_id", roomMsgPayload.RoomID, "client_id", ac.ID)
								break
							}
						}
					}
					continue
				}

				// handle leave (room_status false + empty content) -> remove sender from room
				if roomMsgPayload.Content == "" {
					if !roomExists {
						slog.Warn("HUB room leave for non-existent room", "room_id", roomMsgPayload.RoomID)
						continue
					}
					for rc := range roomClients {
						if rc.ID == roomMsgPayload.SenderID {
							delete(roomClients, rc)
							slog.Info("HUB removed client from room", "room_id", roomMsgPayload.RoomID, "client_id", rc.ID)
							break
						}
					}
					if len(roomClients) == 0 {
						delete(h.ChatRoomClients, roomMsgPayload.RoomID)
						slog.Info("HUB: room deleted (empty)", "room_id", roomMsgPayload.RoomID)
					}
					continue
				}

				// regular room message -> ensure room exists locally
				if !roomExists {
					slog.Warn("HUB room message received but room does not exist locally", "room_id", roomMsgPayload.RoomID)
					continue
				}

				// fan-out to all room members non-blocking
				for roomClient := range h.ChatRoomClients[roomMsgPayload.RoomID] {
					// dedupe by user id (in case same user has multiple connections)
					select {
					case roomClient.Send <- roomMsgPayload:
						slog.Info("HUB: message sent to room client", "room_id", roomMsgPayload.RoomID, "client_id", roomClient.ID)
					default:
						slog.Warn("HUB failed to send room message, client.Send full", "client_id", roomClient.ID, "room_id", roomMsgPayload.RoomID)
					}
				}
				slog.Debug("HUB room fanout summary", "room_id", roomMsgPayload.RoomID,)

				// for activeClient := range h.Clients {
				// 	if activeClient.ID == roomPayload.RecieverID || activeClient.ID == roomPayload.SenderID {
				// 		// if there is any client in clients [] exists whose id matches any of these -> if they are active and then are binded to exchange
				// 		// storing them in a room

				// 		h.ChatRoomClients[roomPayload.RoomID][activeClient] = true // storing those clients in the room
				// 		// key[innerMapKey] = val //* if this set to true means innerMap key is stored whose, we use "key" only <- same logic in setting active clients
				// 		// * if there is any activeClient coming -> read and if it is -> set "key" being client to be true so -> now it stored as pair, we check on everything from key
				// 		slog.Info("Room is created and clients are stored in chatRoomClientsNestedMap","roomID",roomPayload.RoomID,"clients :",h.ChatRoomClients)
					
				// 		// but need to send message too to all room clients
				// 		for eachRoomClient := range h.ChatRoomClients[roomPayload.RoomID] {
				// 			eachRoomClient.Send <- roomPayload // redirect payload to eachClient send chan which would be proactively checked by the active client writer to write res what recieved in send chan
				// 		}
				// 	}
				// }


			
			// tracking online-offline activity 
			case currentClient := <- h.Online :
				// set it online with status being true
				
			
					// // first if it exists there in clients
					// _,ok := h.Clients[currentClient]
					// if ok {
					// 	// test - if it shows these indicators
					// 	// simply loggin for now to test it
					// 	for client := range h.Clients {
					// 		// checking which client we are targetting - we are here talking about each active client, not sender or reciever for now
					// 		if client.ID == currentClient.ID {
					// 			// send status to that client's OnlineStatus chan 
					// 			client.OnlineStatus <- true // when recieved it will check if it is true then send -> response as online 🟢 
					// 			slog.Info("user is online🟢","userID:",currentClient.ID)
					// 		}else {
					// 			// else label it as offline
					// 			client.OnlineStatus <- false  
					// 		} 
					// 	}

					// checking what is being recieved here and send to all clients which ->
					statusPayload := StatusPayload{
						Status: "online🟢",
						UserID: currentClient.ID, // passing id of current client which is active
					}
					slog.Info("user activity status", "status", statusPayload.Status, "userID", currentClient.ID)

					// broadcast to all other connected clients
					for eachClient := range h.Clients {
						if eachClient.ID == currentClient.ID {
							continue
						}
						
						select {
						case eachClient.BroadcastStatus <- &statusPayload:
						default:
							slog.Warn("HUB failed to broadcast online status, BroadcastStatus channel full", "client_id", eachClient.ID)
						}
					}
					
					
					//  redirects to all client's statusChan where writers writes through status Chan seperately handels writing fo rit
			
			

			// & for targetted client notification of all types {likes,comment...} 
			// this chan is dedicated to type related notifications only which check for type exclusively
			case notifyPayload := <- h.TargettedClientNotificationTypeOnly :
				slog.Info("notification payload is successfully recieved in hub's TargettedClientNotificationTypeOnly chan")
				
				var targettedActiveClient *Client 
				// ! Optimization issues -> had to check all client from active clients []
				// fix - just check if that specific client is in the [] of active clients
				targettedActiveClient,exists := h.ClientStore[notifyPayload.RecieverID] //* checking if targtted reciever client is active or not
				if !exists || targettedActiveClient == nil {
					slog.Warn("could not send notification to the reciever","error","client is currently offline❌")
					// ! important - don't put return as that would crash the application
					continue // just break out of this nested loop and keep running hub
				}

				//test - since only type makes the diffrence in the payload , strict type check is must so only send that type of notification
				switch notifyPayload.Type {
					// case handeling for type check
					// testing with switch statement cause select checks for the channel
					case  "like_posted" :
						slog.Info("successfully recieved payload of type Like")
						// ticker sends when done ticked
						targettedActiveClient.Send <- notifyPayload
					case "comment_posted" :
						slog.Info("successfully recieved payload of type Comment")
						// ticker sends when done ticked
						targettedActiveClient.Send <- notifyPayload
					// must have default for fallback if nothing meets the conditions declared above
					default : 
						slog.Info("unknown notificaiton payload","notification_type",notifyPayload.Type)
				}
				// if found that client send it to client chan where writer recieves and send as a response
				slog.Info("successfully sent notification to the client✅","recieverID",notifyPayload.RecieverID)


			case currentDisconnectedClient := <- h.Offline :
					// for disconnect it would be simulanously removed from clients, so checking if its not there and successfully disconnected
					// time.Sleep(time.Millisecond * 200) // adding a lil delay to let other go routine finish its work
					// deferred disconnected offlien client is redirected to this chan and checking if it recieved here and we can read from the chan
					statusPayload := StatusPayload{
						UserID: currentDisconnectedClient.ID,
						Status: "offline🔴",
					}
					slog.Info("user activity status", "status", statusPayload.Status, "userID", currentDisconnectedClient.ID)

					// sending this status payload to all other active clients
					for eachClient := range h.Clients {
						if eachClient.ID == currentDisconnectedClient.ID {
							// & continue logic -> if condition passes -> jumps to the next ieration
							continue // skip the current client
						}
						
						select {
						case eachClient.BroadcastStatus <- &statusPayload:
						default:
							slog.Warn("HUB failed to broadcast offline status, BroadcastStatus channel full", "client_id", eachClient.ID)
						}
					}

					// _,disconnectedClientExists := h.Clients[currentClient]
					// if !disconnectedClientExists {
					// 	slog.Info("user is offline🔴","userID:",currentClient.ID)
					// }


		}
		
	}
}



// methods related to Client

var (
	pongWaitDuration = 10 * time.Second  // time frame for how long server would wait for ping reponse to check if client is active or not 
	pingInterval = (pongWaitDuration * 9) / 10 // time frame for sedning ping message that client is active
)

// method that belongs to the type Client to read incoming data struct of type InboundMsg
func(c *Client) MessageReader(db *gorm.DB) {
//  then ,query to store msg into db, redirect msg output as Outbound to broadcast chan in hub

	// * defer this function invocation at the end
	defer func() {

		// this is where clients gets disconnected is tracked

		// unregistering the channel on which this meth is defined on - after read
		// since hub is centre of all things, redirecting client that would be disconnected to the disconnectedCLient chan -> rest is done by hub method RunService
		c.Hub.DisconnectedClients <- c
		c.Hub.Offline <- c // redirecting disconnected client to offline chan which stores val of type client-> so client is stored which gets disconnected
		// call optional OnDisconnect callback (set by controller) to allow unbind or cleanup
		if c.OnDisconnect != nil {
			c.OnDisconnect(c.ID)
		}
		close(c.Send)

		
		c.WebsocketConnection.Close() //! close close as soon as *c is unregistered

		// @ chan methodology 
		// * var <- chan (reading from chan n storing onto var or reading onto var)
		// ! chan <- vat (redirecting output to chan)
	}()


	// *heartbeat - for pong response check after this set interval
	// set timeout to check to checl for incoming req
	err := c.WebsocketConnection.SetReadDeadline(time.Now().Add(pongWaitDuration)) // firstly set -> for how long we would wait for pong to recieved
	if err != nil {
		slog.Error("failed to set read deadlien timeout","error",err)
		return
	} 
	// this keep ping-pong cycle running, untill pong is not delivered by the client -> would be disconnected
	c.WebsocketConnection.SetPongHandler(func(pongMsg string) error {
		slog.Info("pong")
		// if incoming req came -> set timer again, if not, it won't set cuz client would be disconnected already
		return c.WebsocketConnection.SetReadDeadline(time.Now().Add(pongWaitDuration)) //$ set readDeadline again for checking msg is incoming in this timeframe from now
	})

	// when request is recieved, this is checked by readDeadline if it is coming inside the time frame,... and if request pong handler checks 
	// if pong handler recieves pingTypeMessage, if yes, automatically responds with pong which is checked by the pong handler -> if ponged by the clients browser then resets the readDeadline
	// which again checks for -> if ping type msg is incoming by the pong handler then reset back -> this way loop is keep running always, if not -> client would be axtivated
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

		// & if payload has roomID -> publish in the exchange marked with room.ID -> consumer redirects to hub which sends response to the room clients
		if msg.RoomID != 0 {
			var senderUser models.User
			if senderResErr := db.Select("id", "username", "name").First(&senderUser, msg.SenderID).Error; senderResErr != nil {
				if senderResErr == gorm.ErrRecordNotFound {
					slog.Error("sender not found", "sender_id", msg.SenderID)
					return
				}
				slog.Error("failed to get sender details", "error", senderResErr)
				return
			}

			payload := &ClientNotifyPayload{
				Type:       msg.Type,
				PostID:     msg.PostID,
				CreatedAt:  time.Now(),
				SenderID:   msg.SenderID,
				RecieverID: msg.RecieverID,
				SenderName: senderUser.Name,
				RoomID:     msg.RoomID,
				RoomStatus: msg.RoomStatus,
				Content:    msg.Content,
			}

			if msg.RoomStatus || msg.Content == "" {
				c.Hub.RoomClientsPayloads <- payload
				slog.Info("room membership update routed to hub", "room_id", msg.RoomID, "sender_id", msg.SenderID, "room_status", msg.RoomStatus)
				continue
			}

			//& publish via hub wrapper so it safely checks broker presence
			c.Hub.Publish(msg.RoomID, payload)
			slog.Info("room message published via broker", "room_id", msg.RoomID, "sender_id", msg.SenderID)
			continue
		} // ...room end...
		
		// fetch sender and receiver details
		var senderUser models.User
		var recieverUser models.User

		if senderResErr := db.Select("id", "username", "name").First(&senderUser, msg.SenderID).Error; senderResErr != nil {
			if senderResErr == gorm.ErrRecordNotFound {
				slog.Error("sender not found","sender_id",msg.SenderID)
				return
			}else {
				slog.Error("failed to get sender details","error",senderResErr)
				return
			}
		}
		

		if err := db.Select("id", "username", "name").First(&recieverUser, msg.RecieverID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				slog.Error("receiver not found", "reciever_id", msg.RecieverID)
			} else {
				slog.Error("failed to fetch receiver", "error", err)
			}
		}

		senderName := senderUser.Name
		recieverName := recieverUser.Name

		// at this point, reader would have got sender's/reciever's name from his id recieved in payload sent from the frontend✅✅
		// -> since same func is acting for both we can't assume by default active client is the sender in this case
		

		// * if recieving msg is correct of type inBMsg - making query to store in db
		err = db.Create(&models.Message{
			// we not storing name just we will send result back with name 
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
		
		payload := &ClientNotifyPayload {
			// which is recieved bny broadcast and send to that client 
			Type: msg.Type,
			PostID: msg.PostID,
			CreatedAt: msg.CreatedAt,
			SenderID: msg.SenderID,
			RecieverID: msg.RecieverID,
			Content : msg.Content,
			SenderName: senderName,
			RecieverName: recieverName,
		}

		// send ack back to the sender so frontend can update UI (delivery/persisted)
		ack := &ClientNotifyPayload{
			Type:      "ack",
			SenderID:  msg.SenderID,
			RecieverID: msg.RecieverID,
			Content:   msg.Content,
			CreatedAt: time.Now(),
		}
		// non-blocking ack send but attempt with small timeout to avoid blocking reader
		go func() {
			select {
			case c.Send <- ack:
			case <-time.After(time.Second):
				slog.Warn("failed to send ack to sender, sender.Send busy", "sender_id", c.ID)
			}
		}()

		// & mark delivery in "notifications" exchange for this userID when its not nil
		if msg.RecieverID != 0 && payload.RoomID == 0 {
			//* checked by consumer which -> redirects to targetted chan of hub -> which sends to targetted client only for reciever send chan
			c.Hub.Publish(msg.RecieverID, payload)
		} else {
			//$ for broadcasting to all except the sender
			c.Hub.Broadcast <-payload // send payload to broadcast chan of hub which sends to all client's send chan to send response to all
		}

	}
}


// method that belongs to the type *Client which w-> checks for chan data coming in client's send chan -> write it to the ws writer
func(c *Client) MessageWriter() {

	defer c.WebsocketConnection.Close() // defer to close chan 

	// * constant check for active client check via heartbeat method <- ping client, waut how long it does to pong if not, client is declared inactive
	ticker := time.NewTicker(pingInterval) // when ticker timer is finished(ticked)<- recieved via chan ticker.C

	writerLoop :
	for {
		select {
			// loop - looping over every Send chan data <- loop cause it continously keep checking if c.Send has outbound type data
		case msg, ok := <- c.Send :
			if !ok {
				// The channel was closed, which means the client disconnected. Break the loop.
				slog.Info("WRITER Send channel closed, breaking writer loop", "client_id", c.ID)
				break writerLoop
			}
			slog.Info("WRITER about to WriteJSON to client", "client_id", c.ID, "sender", msg.SenderID, "receiver", msg.RecieverID, "type", msg.Type)
			// writes current msg to the json
			err := c.WebsocketConnection.WriteJSON(msg) //* writes data to ws connection,where it is recieved as event.Data on the defined handler url path
			if err != nil {
				slog.Error("WRITER failed to WriteJSON", "client_id", c.ID, "error", err)
				// bug - you can't break out of infinite loop like that, u would need "label Blocking"
				break writerLoop
			}
			slog.Info("WRITER successfully wrote message to WebSocket", "client_id", c.ID)
		// todo - must redirect status to client when got info if user was offline or online
		// checking if case is able to read from BroadcastStatus chan - payload of type statsuPayload consisting data who is active or not
		case statusPayload := <- c.BroadcastStatus :
			// each active client recieves it
			// respond by sending it to the reciever -


			// this is now recieved on all active clients request's writers chan as this is where actually it is recieving from the hub
			slog.Info("user activity status","status",statusPayload.Status,"userID",statusPayload.UserID)
			if err := c.WebsocketConnection.WriteJSON(statusPayload); err != nil {
				slog.Warn("status write failed, unregistering client","error",err)
				select {
				case c.Hub.DisconnectedClients <- c:
				default :
				}
				break writerLoop
			}
		
		// * whenever specified timer is ticker and recieved .C that it has been ticked, we send a ping type msg, if client would be active it would respond with pong by the browser itself
		case  <- ticker.C :
			// todo - need to set waited interval so it reads pong 
			// if that time of ticker is elasped and then this is invoked recieved the .C tick
			slog.Info("ticker is finished and successfully ticked","pinging client with UserID :",c.ID)

			// sending a ping to client <- automatically respond with pong inbuilt browser support
			err := c.WebsocketConnection.WriteMessage(websocket.PingMessage,[]byte(""))
			if err != nil {
				slog.Error("failed to write messagr","error",err)
				break writerLoop
			}
			}// .. select end ..
		}
			
	}
			
			
			
			
			
		
		
		