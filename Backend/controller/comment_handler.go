package controller

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// @types

//  type that stores which belongs to the -> CommentDbModel
type CommentController struct {
	CommentDbModel *services.CommentDBModel
	PushNotificationService *services.PushNotificationService
}

// bug - comment req type payload - mistakenly used fullComment struct - which is for db, not what client would send

type CommentRequest struct {
	// * now we will fetch id from url directly
	// PostID uint `json:"post_id" binding:"required"`
	Content string `json:"content" binding:"required"` 
}

type CommentDeleteRequest struct {
	PostID uint `json:"post_id" binding:"required"`
}


//  func that returns the instance of type CommentController which -> stores handlers method for comments
func NewCommentController(commentDbModel *services.CommentDBModel,pushNotificationService *services.PushNotificationService) *CommentController {
	return &CommentController{
		CommentDbModel: commentDbModel,
		PushNotificationService: pushNotificationService,
	}
}

// note - every req need userId for which user is posting comment - fetch from auth middleware set Context
// App's controller pattern --> use cC type case, modelReq for binded req payloads

//  controller method that belongs to CommentController -> which stores all the comment related handle methods
func(cC *CommentController) PostComment(c *gin.Context) {
// pointer context cause we need actual pointer value,not just the copy 

// * just need comment content, will fetch postID from url path to query post comment on that post
	var commentReq CommentRequest //fixed - this type of data type struct json client would bind the request with 
	err := c.ShouldBindJSON(&commentReq) // req should bind this type of struct data in json
	
	// err pattern
	if err != nil {
		errMsg := "invalid request payload"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",err)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
		})
		return
	}
	//* but we need userID from req, not sent by client, fetch from auth's middleware passed req's state context
	userID := c.GetUint("user_id") // fetched from req context, must pass token -> stores userID in its claims map <- map[keyString]:userId(uint)
	
	//  validating userID, as if auth middleware don't work or removed, so panic before crash
	if userID == 0 {
		errMsg := "userID not found,failed to post comment"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
		})
		return
	}

	// fetching postID -> to where post comment , form <- post/comment/:commentid
	postIDStr := c.Param("postid")
	postID,err := strconv.Atoi(postIDStr)
	if err != nil {
		errMsg := "Invalid postID in url"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Ok: false,
			Status: errMsg,
		})
		return
	}

	// instance of comment for creating comment
	commentToPost := models.Comment{
		UserID: userID,// fetched from auth token context
		//  binded req payload
		Content: commentReq.Content, 
		PostID: uint(postID),
	}

	
	
	//  if req correctly bind that it expects ✅
	//   access method from controller.DbModel which stores method to query comment into the CommentDbModel
	postedComment,err := cC.CommentDbModel.PostComment(commentToPost) // fixed - inserting comment instance
	// err handle
	if err != nil {
		errMsg := "failed to post comment"
		code := http.StatusServiceUnavailable
		slog.Error(errMsg,"error",err)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
		})
		return
	}	

	// testing - invoking method attached on *pns -> to redirect corres to corres chan for reader to read
	// todo - fetch full comment payload and send to pns method who redirects payload to the pns
	commentNotificationPayload,err := cC.CommentDbModel.GetUserDetailByPostsCommentID(postedComment.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("no comment found","error",err)
			return
		}
		slog.Error("failed to comment info for notification payload","error",err)
		return
	}
	cC.PushNotificationService.NotifiesCommentPostedOnPost(commentNotificationPayload) // redirect payload to the pns method which redirects to tthe pns
	// cC.PushNotificationService.NotifiesCommentPostedOnPost(*postedComment)
	


	//  if posted comment successfully 🚀
	// response that "comment" is posted
	c.JSON(http.StatusOK,utils.CommentSuccessResponse{
		Ok: true,
		Status: "comment posted successfully",
		Code: http.StatusCreated,
		Comment: postedComment.Content,
		Data: postedComment,
	})

}


//  load all comments associated with users & posts from the db
func(cC *CommentController) LoadAllCommentsAssociatedWithPostAndUsers(c *gin.Context) {


	userID := c.GetUint("user_id") 
	if userID == 0 {
		errMsg := "userID not found,failed to load all comments"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
		})
		return
	}

	comments,err := cC.CommentDbModel.LoadAllComments()
	if err != nil {
		errMsg := "failed to load comments!."
		code := http.StatusServiceUnavailable
		slog.Error(errMsg,"error",err.Error())
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
		})
		return
	}

	// response that "post" is created - if query was successfull
	c.JSON(http.StatusOK,gin.H{
		"Status": "comments loaded",
		"Code": http.StatusOK,
		"comments": comments,
	})

}

// delete comment - by providing postID(from client), userID(from its token on its req)
func(cC *CommentController) DeleteCommentByUser(c *gin.Context) {


	// todo - we could delete comment by fetching postID directly from url Param instead of passed json

	// client would send which post comment to delete
	var commentDeleteReq CommentDeleteRequest

	err := c.ShouldBindJSON(&commentDeleteReq)
	if err != nil {
		errMsg := "invalid request"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			// * sending bool for client side validation of server sent response with ease
			Ok: false,
			Status: errMsg,
		})
		return
	}
	//  fetch userID from req as set by auth middleware from header's token~str
	userID := c.GetUint("user_id")
	if userID == 0 {
		errMsg := "userID not found,failed to delete comment"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Ok: false,
			Status: errMsg,
		})
		return
	}

	//  call CommentDbModel's delet method to query call to db to delete the comment
	err = cC.CommentDbModel.DeleteCommentByUserID(userID,commentDeleteReq.PostID)
	if err != nil {
		if err == sql.ErrNoRows {
		errMsg := "comment not found on the post"
		code := http.StatusNotFound
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Ok: false,
			Status: errMsg,
		})
		return
		}
		errMsg := "failed to delete comment"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Ok: false,
			Status: errMsg,
		})
		return
	}

	// if found n deleted
	c.JSON(http.StatusOK,utils.SuccessResponse{
		Ok: true,
		Status: "comment deleted successfully",
		Code: http.StatusOK,
	})
}


//  load all comnments related to a post
func(cC *CommentController) LoadPostComments(c *gin.Context) {

	// userID - 'token' validation
	userID := c.GetUint("user_id")
	if userID == 0 {
		errMsg := "userID not found,failed to load post comments"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Ok: false,
			Status: errMsg,
		})
		return
	}

	// fetch postID from url path -> fetch all comments of that post
	postIDStr := c.Param("postid")
	postID,err := strconv.Atoi(postIDStr)
	if err != nil {
		errMsg := "Invalid postID in url"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Ok: false,
			Status: errMsg,
		})
		return
	}

	comments,err := cC.CommentDbModel.LoadAllCommentsOfPost(uint(postID))
	if err!= nil {
		errMsg := "server error,failed to load all comments of this post"
		code := http.StatusServiceUnavailable
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Ok: false,
			Status: errMsg,
		})
		return
	}
	
	// fix - changed to send name too not just user.id of person who is commenting
	// might need err handeling for nil res - if query was a success operation but nothing returned so only nil data struct returned from db query call
	if comments == nil {
		comments = []*services.CommentsData{} //empty on that post, not err
	}

	c.JSON(http.StatusOK,utils.CommentSuccessResponse{
		Ok: true,
		Status: "successfully loaded all comments of this post",
		Data: comments,
	})


}