package controller

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

//@ types

// controller type that stores LikeDBModel which stores methods belongs to it related to like db queries
type LikeController struct {
 LikeDbModel *services.LikeDBModel
 PushNotificationService services.PushNotificationService
}


//  type of data struct client would send for liking post
type LikeRequest struct {
	// just need to know which Post is being liked,rest is not sent by the client
	PostID uint `json:"post_id"`
}

//  func that returns intance of type LikeController -> which stores handler methods for like related methods
func NewLikeController(likeDbModel *services.LikeDBModel,pushNotificationService *services.PushNotificationService) *LikeController {
	return &LikeController{
		LikeDbModel: likeDbModel,
		PushNotificationService: *pushNotificationService,
	}
}


// func that updates like on the post
func(l *LikeController) UpdateLike(c *gin.Context) {

	var likeReq LikeRequest
	
	err := c.ShouldBindJSON(&likeReq)
	if err!= nil {
		errMsg := "invalid request payload"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",err)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
		})
		return
	}

 	slog.Info("status","errors","payload accepted & passed for further processing...")
	
	//  fetching user id from the auth's set context from client token
	userID := c.GetUint("user_id")
	// todo - need to setup auth middleware
	//  fixed - added to use auth middleware for these routes
	// since its coming from auth, needs validation as if auth skipped by some nafarious way
	if userID == 0 {
		errMsg := "userID not found,failed to like post"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
		})
		return
	}
 	slog.Info("status","errors","successfully retrieved user_id from auth")

	// creating instanceo of like struct - that need to pass to LikeDbModel to insert like into the db
	likeToUpdate := models.Like{
		//  rest are done by the server
		PostID: likeReq.PostID,
		UserID: userID,
	}

 	slog.Info("status","errors","like instance created from like request for querying into the db...")

	// checking if user has already liked or not
	alreadyLiked,err := l.LikeDbModel.CheckIfAlreadyLiked(likeToUpdate)
	if err != nil {
		//  have to handle both err sent by repo method
		//  cause no implicit err return by query -> so on scan check what if there was no rows returned
		if err == sql.ErrNoRows {
			alreadyLiked = nil //* not already liked
		} else {
			// normal err
			//  just if something else gone wrong
			errMsg := "failed to check likes on the post"
			code := http.StatusServiceUnavailable
			slog.Error(errMsg,"error",err)
			c.AbortWithStatusJSON(code,utils.ErrResponse{
				Status: errMsg,
			})
			return
		}
	}

 	slog.Info("status","errors","post is not already liked,code proceeding to post like....")

	//  if it is already liked - update like count by +1
	if alreadyLiked != nil {
		updatedLikes,err := l.LikeDbModel.UpdateLikeByOneCount(likeToUpdate)
		if err != nil {
			errMsg := "failed to update like,post deleted or unavailable"
			code := http.StatusInternalServerError
			slog.Error(errMsg,"error",err)
			c.AbortWithStatusJSON(code,utils.ErrResponse{
				Status: errMsg,
			})
			return
		}

 	slog.Info("status","errors","like updated for already liked post")

	// testing like chan which redirecs like to the 
	// todo - need to send reciever'sID of the client whose post is being liked
	postDetails,err := l.LikeDbModel.GetUserDetailsByPostID(likeReq.PostID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound,utils.ErrResponse{
				Status: err.Error(),
			})
			return
		}
		slog.Error("error",err)
			c.AbortWithStatusJSON(500,utils.ErrResponse{
				Status: err.Error(),
			})
			return
	}
	
	postDeets := models.PostDetailedNotification{
		PostUserDetails: postDetails,
		LikeData: updatedLikes,
	}

	l.PushNotificationService.NotifiesLikePostedOnPost(&postDeets)

		// send resp to client - that post has been liked
		c.JSON(http.StatusOK,utils.LikeSuccessResponse{
			Ok: true,
			Status: "post liked👍",
			Code: http.StatusOK,
			Like: *updatedLikes,
		})

	}else {
		// * if not already liked, first like by the user
		firstLike,err :=l.LikeDbModel.PostFirstLike(likeToUpdate)
		if err != nil {
			errMsg := "failed to update likes"
			code := http.StatusInternalServerError
			slog.Error(errMsg,"error",err)
			c.AbortWithStatusJSON(code,utils.ErrResponse{
				Status: errMsg,
			})
			return
		}

	
	updatedPostDetails,err := l.LikeDbModel.GetUserDetailsByPostID(likeReq.PostID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound,utils.ErrResponse{
				Status: err.Error(),
			})
			return
		}
		slog.Error("error",err)
			c.AbortWithStatusJSON(500,utils.ErrResponse{
				Status: err.Error(),
			})
			return
	}

	updatedPostDeets := models.PostDetailedNotification{
		PostUserDetails:updatedPostDetails,
		LikeData: firstLike,
	}

	// testing like notification service
	l.PushNotificationService.NotifiesLikePostedOnPost(&updatedPostDeets)


 	slog.Info("status","errors","finally posted first like on the post")

		c.JSON(http.StatusOK,utils.LikeSuccessResponse{
			Ok: true,
			Status: "post liked👍",
			Code: http.StatusOK,
			Like: *firstLike,
		})
	}

	// @ flow
	// Get req hundi aa --> dekh os to je err a reya,par err eh ki res nhi aya query ta hogyi
	// ta jsnu check kr rahe usnu -> set nil <- kyo? kyoki koi row exists nhi krdi
	//  je oh nil result nhi aunda - mtlb struct nil nhi <- mtlb data hai -> update else insert new
	// resp for both



}