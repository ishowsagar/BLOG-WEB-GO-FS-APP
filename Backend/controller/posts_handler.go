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

// types

type PostController struct {
	PostDbModel *services.PostDBModel
	// todo - type for noti service
	PushNotificationService *services.PushNotificationService
}

//  func that creates intance of type PostController which -> stores handler method belong to post related 
func NewPostController(postDbModel *services.PostDBModel,pushNotificationService *services.PushNotificationService) *PostController {
	return &PostController{
		PostDbModel: postDbModel,
		PushNotificationService: pushNotificationService,
	}
}

//  type of data struct request -> client would make to create post
type PostRequest struct {
	Title string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"` 
}

// method that belongs to the PostController type --> which creates post
func(pC *PostController) CreatePost(c *gin.Context) {

	var postReq PostRequest
	err := c.ShouldBindJSON(&postReq)
	if err != nil {
		errMsg := "invalid request payload"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	userID := c.GetUint("user_id") 
	if userID == 0 {
		errMsg := "userID not found,failed to create Post"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	// instance of type post -> this type of post dbQuery method accepts to CreatePost 
	postToCreate := models.Post{
		UserID:userID,
		Content: postReq.Content,
		Title: postReq.Title,
	}
	
	createdPost,err := pC.PostDbModel.CreatePost(postToCreate)
	if err != nil {
		errMsg := "failed to create post"
		code := http.StatusServiceUnavailable
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	// * if post is created - redirect post to send post to chan
	// todo - add more methods for redirection of created input to channalize them and add readers for them in service reader service which has dedicated case blocks for them
	pC.PushNotificationService.NotifiesPostCreation(*createdPost)

	// response that "post" is created - if query was successfull
	c.JSON(http.StatusOK,utils.PostSuccessResponse{
		Ok: true,
		Status: "post created successfully ✅",
		Code: http.StatusOK,
		Post: *createdPost,
	})

	
}



//  delete post req -> /posts/:id
func(p *PostController) DeletePost(c *gin.Context) {

	postID := c.Param("id")
	id,err := strconv.Atoi(postID)
	if err!= nil {
		errMsg := "Please pass correct post ID, either wrong format or inncorrect ID"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: err.Error(),
			Ok: false,
		})
		return
	}

	err = p.PostDbModel.DeletePostById(uint(id))
	if err!= nil {
		errMsg := "failed to delete post"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: err.Error(),
			Ok: false,
		})
		return
	}

	c.JSON(http.StatusNoContent,utils.SuccessResponse{
		Status: "successfully deleted post",
		Code: http.StatusNoContent,
		Ok: true,
	})
}


//  load all posts from the db
func(pC *PostController) LoadFeed(c *gin.Context) {


	// uncomment for auth
	userID := c.GetUint("user_id") 
	if userID == 0 {
		errMsg := "userID not found,failed to load feed"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	posts,err := pC.PostDbModel.LoadFeed()
	if err != nil {
		errMsg := "failed to load feed!."
		code := http.StatusServiceUnavailable
		slog.Error(errMsg,"error",err.Error())
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	// response that "post" is created - if query was successfull
	c.JSON(http.StatusOK,gin.H{
		"Ok": true,
		"Status": "feed loaded",
		"Code": http.StatusOK,
		"Post": posts,
	})

}


//  load post from the db by its id - /api/feed/post/:id
func(pC *PostController) GetPostByID(c *gin.Context) {
	// uncomment for auth
	userID := c.GetUint("user_id") 
	if userID == 0 {
		errMsg := "userID not found,failed to load post"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}


//  frontend will call this handler method on the url, fethc id from the url to make db query call and send response
	postIDStr := c.Param("id")
	id,err := strconv.Atoi(postIDStr)
	if err != nil {
		slog.String("error",err.Error())
		return
	}

	post,err := pC.PostDbModel.GetPostbyID(id)
	if err != nil {
		errMsg := "failed to load post!."
		code := http.StatusServiceUnavailable
		slog.Error(errMsg,"error",err.Error())
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	// response that "post" is created - if query was successfull
	c.JSON(http.StatusOK,gin.H{
		"Ok": true,
		"Status": "feed loaded",
		"Code": http.StatusOK,
		"Post": post,
	})

}


//  func that get post count 
func(p *PostController) GetPostCountByUserID(c *gin.Context) {
	// todo - could do with json, but testing with id url param first
	//  since we need to know about posts by UserID, we are alr fetchig from req token using auth middleware
	

	userID := c.GetUint("user_id") 
	if userID == 0 {
		errMsg := "userID not found,failed to load posts count"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}
	slog.String("info :","userID fetched")
	postCount,err := p.PostDbModel.GetPostCountForAUser(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.String("info :","sql no err block")
			c.AbortWithStatusJSON(http.StatusNotFound,utils.ErrResponse{
				Status: "no post found",
				
			})
			return
			} else {	
				slog.String("info :","nil err block")
				errMsg := "failed to fetch post count"
				code := http.StatusServiceUnavailable
				slog.Error(errMsg,"error",errMsg)
				c.AbortWithStatusJSON(code,utils.ErrResponse{
					Status: errMsg,
					Ok: false,
				})
			}
			return
		}
		
		slog.String("info :","req successfull")
	c.JSON(http.StatusOK,utils.CommentSuccessResponse{
		Ok: true,
		Status: "successfully post count",
		Data: postCount,
		
	})


}