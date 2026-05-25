package services

import (
	"context"
	"log/slog"
	"time"

	// "github.com/goforj/godump"
	"github.com/ishowsagar/go-blog-web-application/models"
)

// @types declaration

//  central hub which -> hae notifi chan for post related data,and rest
type PushNotificationService struct {
	//! central hub for all type of notification chan we needed throughout the application
	PostNotification chan models.Post //  chan for recieving and sending post related data struct notifications
	LikeNotification chan *models.PostDetailedNotification
	CommentNotification chan *models.CommentPayload
	Hub *Hub // WebSocket hub for broadcasting to connected clients

}

// PushBroker defines the methods the notification service uses to publish events.


// ! workflow - add reader select's case for reading redirected corresponding ouput by method attached on *pns => which redirects corres ouput to corres chan 

// returns instance of type pns -> which stores chans
func NewPNService() *PushNotificationService {

	pns := &PushNotificationService{
		PostNotification: make(chan models.Post,5), // can buffer upto 5 posts redirections
		// chan whose -> value is rdirected Like
		LikeNotification: make(chan *models.PostDetailedNotification, 10), // can max buffer upto 5 likes
		CommentNotification: make(chan *models.CommentPayload,10), // must add buffer size for chan to make them robust

	}

	// starting go routine to keep reading notification
	slog.Info("notification service has started🚀","waiting for any notification to come⏳","...")
	
	// instead of just loggin to the term and reading,we can store it somewhere or cache it, so we can fetch in frontend to show them
	// todo - Store notification to render them to the client 

	// idea #1 - store
	//#1 -> store noti in db/or cache
	//#2 -> call the handler to retrieve them correspondingly
	

	// idea #2 - websockets✅ trying this
	//#1 -> use Websockects or SSE
	//#2 -> keep connection alive, keep hitting throgh w.s to push json to the browser in realtime


	// noti logger
	go pns.StartService() //* whenever post is redirected it reads to the logger

	return pns
}

// SetHub connects the notification service to the WebSocket hub for broadcasting
func(pns *PushNotificationService) SetHub(hub *Hub) {
	pns.Hub = hub
}


//  logs the redirect notifcations posts data to pns notifation Chan
func(pns *PushNotificationService) StartService() {

	
	//  since its a pointer val -> ranging over each notification data and logging out these things out of it
	// todo - need a reader for other services like -> Like, mioght add more later
	// bug - due to all chan, need a infinite loop with select group to fire only needy event
	// fixed - 'return' was missing from the select's case 
	// testing it with infinite loop & select group
	for {
	// testing - instead of loggin id's actually log what happened
		select {
			// * if reading from these chan are successfull
		case post := <- pns.PostNotification :
			slog.Info("proccessing notification for post","postID :",post.ID,"userID",post.UserID)
			slog.Info("notification recieved ✅","post created-by",post.UserID)
			// Broadcast to connected WebSocket clients
			payload := &ClientNotifyPayload{
				SenderID: post.UserID,
				RecieverID: 0, // Broadcast to all
				Type: "post_created",
				Content: post.Content,
				PostID: post.ID,
				CreatedAt: time.Now(),
			}

			
			if pns.Hub != nil {
				pns.Hub.Broadcast <- payload
			}
		// todo - add a redirection method to redirect Like output so this chan can read
		// fixed - added method which invokes this method through the corresponding handler
		case likedata := <- pns.LikeNotification :
			slog.Info("someone liked your post","postID",likedata.LikeData.PostID,"userID",likedata.LikeData.UserID)
			slog.Info("notification recieved ✅","liked By UserID",likedata.LikeData.UserID, "On PostID",likedata.LikeData.PostID,"Post like-Count",likedata.LikeData.LikeCount)
			// Broadcast to connected WebSocket clients
			payload := &ClientNotifyPayload{
				SenderID: likedata.LikeData.UserID,
				RecieverID: likedata.LikeData.UserID, // Broadcast to all
				Type: "like_posted",
				Content: "Someone liked your post",
				PostID: likedata.LikeData.PostID,
				CreatedAt: time.Now(),
			}
			// * When pns method are invoked, they redirect output to pns chan <- which upon recieving send to hub
			// test - implement broker to route them to hub for targetted noitification
			if pns.Hub != nil {
				// if broker is passed correctly
				// todo - fetch reciever from repo by -> getting which user post is being interacted
				if payload.RecieverID != 0 {
					err := pns.Hub.BrokerInterface.PublishEvents(payload.RecieverID,payload)
					slog.Info("successfully published event delivery to the exchange","like_recieverID :",payload.RecieverID)
					if err != nil {
						// * also need consumer which checks for this event and sends to the targetted user
						slog.Error("failed to publish and stamp like event delivery","error",err)
						// no return cause if it fails we won't shutdown
					}
				}
			}
			// uncommented below 3 lines of code cause now we want p2p notification,not general sent to all
			// if pns.Hub != nil {
			// 	pns.Hub.Broadcast <- payload
			// }
		// test - adding a reader select's case to read incoming comment chan val
		case comment := <- pns.CommentNotification :
			slog.Info("someone posted comment on your post","postID",comment.PostID,"userID",comment.CommentorID)
			slog.Info("notification recieved ✅","comment",comment.CommentContent)
			// Broadcast to connected WebSocket clients
			payload := &ClientNotifyPayload{
				SenderID: comment.CommentorID,
				// todo - recieving reciever id of the client whose post would be on where commentor would have commented
				RecieverID: comment.RecieverID, // Broadcasting to that client whose post will be 
				Type: "comment_posted",
				Content: comment.CommentContent,
				PostID: comment.PostID,
				CreatedAt: time.Now(),
			}
			//& publish comment delivery on the exchange only if there is recieverID available
			// todo - ch.Publish  to stamp the delivery of this comment in the exchange to commentorID
			// if pns.Hub != nil {
			// 	pns.Hub.Broadcast <- payload
			// }
			pns.Hub.BrokerInterface.PublishEvents(comment.RecieverID,payload) //* stamping the delivery of this payload on the reciever which is checked by the ch.Consume if there is any delivry coming on this ex
			slog.Info("successfully published the delivery of this commentPayload & sent to the broker consumer","recieverID :",comment.RecieverID)
		}
	}


	// for post := range pns.PostNotification {
	// 	//  access to each noti post
	// 	slog.Info("proccessing notification for post","postID :",post.ID,"userID",post.UserID)
	// 	// godump.Dump(post)
	// }
	
	// for like := range pns.LikeNotification {
	// 	slog.Info()
	// }
}


//  which redirects created post to notification chan
func(pns *PushNotificationService) NotifiesPostCreation(post models.Post) {

	// calling retry call to manage post creation notification with timed request
	// * for any chan based func/meth to be run, must run with 'go' keyword to run in goroutine
	go RetryPostQueryWithTimeout(pns,500 * time.Millisecond,post) 

		// //  selects -> runs the only matching block or queue full it
		// select {
		// 	// redirecting post to noti chan, max 5
		// case pns.Notification <- post :
		// 	slog.Info(" post is created","postID",post.ID,"userID",post.UserID)
		// 	default :
		// 	slog.Warn("notification queue is full") 
		// }

}

// method that belongs to the *pns type which -> redirects like to the chan for reader 
func(pns *PushNotificationService) NotifiesLikePostedOnPost(payload *models.PostDetailedNotification) {
	//  test by invoking in like handler which has like to redirect
	slog.Info("like posted","likeID",payload.LikeData.ID)
	go RetryLikeQueryWithTimeout(pns,700 *time.Millisecond,payload)
}

// todo - it will recieved payload of type CommentPayload when invoked in the handler -> redirects data to the psn from where data is flowed to the broker -> hub-> reciever
func(pns *PushNotificationService) NotifiesCommentPostedOnPost(comment *models.CommentPayload) {
	slog.Info("comment posted","commentID",comment.CommentorID)
	// must run go routine else won't fire in background
	go RetryCommentQueryWithTimeout(pns,1700 * time.Millisecond,comment)
}

// runs the redirection post service to -> log passed post with timeout
func RetryPostQueryWithTimeout(pns *PushNotificationService,timeout time.Duration,post models.Post) {
	
	ctx, cancelCall := context.WithTimeout(context.Background(),timeout)
	defer cancelCall()
	
	newTicker := time.NewTicker(70 *time.Millisecond)
	defer newTicker.Stop() // stop ticker after time elasped
	// todo - notifies that like is made, need a reader which reads it that liked chan fired an event to redirects it
	// fixed - done added a reader select's case block which reads if there is something incoming on that chan
	for {
		select {
		case pns.PostNotification <- post :
			slog.Info("post is created","postID",post.ID,"userID",post.UserID)
			return
		case <- ctx.Done() :
				slog.Warn("queue is timed out","dropping post",post.ID,"userID",post.UserID)
				return
		case  <- newTicker.C :
		// retry new ticker afer this time interval
		}
	}
}

// runs the redirection post service to -> log passed post with timeout
func RetryLikeQueryWithTimeout(pns *PushNotificationService,timeout time.Duration,data *models.PostDetailedNotification) {
	
	ctx, cancelCall := context.WithTimeout(context.Background(),timeout)
	defer cancelCall()
	
	newTicker := time.NewTicker(70 *time.Millisecond)
	defer newTicker.Stop() // stop ticker after time elasped
	// todo - notifies that like is made, need a reader which reads it that liked chan fired an event to redirects it
	// fixed - handler is arguementing like to this method,which is redirected to chan for reader to read in select grouy
	slog.Info("like notification proccessing","waiting for any like to be posted","...")

	for {
		select {
			// fix - redirecting like chan output to reader's block which reads if there is value coming in corresponding chan
		case  pns.LikeNotification <- data :
			slog.Info("liked is posted","redirected like to reader chan",data.LikeData.LikeCount)
			return
		case <- ctx.Done() :
				slog.Warn("queue is timed out","dropping like",data.LikeData.ID,"userID",data.LikeData.UserID)
				return
		case <- newTicker.C :
		// retry new ticker afer this time interval
		}
	}
}

// todo - add a method which invokes it in handler and add reader for this whoich redirected correspoinding output
// runs the redirection comment service to -> log passed comment with timeout
func RetryCommentQueryWithTimeout(pns *PushNotificationService,timeout time.Duration,comment *models.CommentPayload) {
	
	ctx, cancelCall := context.WithTimeout(context.Background(),timeout)
	defer cancelCall()
	
	newTicker := time.NewTicker(70 *time.Millisecond)
	defer newTicker.Stop() // stop ticker after time elasped
	slog.Info("comment notification proccessing","waiting for any comment to be posted","...")

	for {
		select {
		case  pns.CommentNotification <- comment :
			slog.Info("comment is posted","redirected comment to reader chan",comment.CommentorID)
			return
		case <- ctx.Done() :
				slog.Warn("queue is timed out","dropping comment",comment.CommentID,"userID",comment.CommentorID)
				return
		case <- newTicker.C :
			// retry new ticker afer this time interval
		}
	}
}