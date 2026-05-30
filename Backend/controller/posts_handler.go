package controller

import (
	"database/sql"
	"fmt"
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

//  func that create  instance of type PostController which -> stores handler method belong to post related 
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
		Post: *createdPost, //* must send id for updating url
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


// handles batch get request
func(p *PostController) FeedBatchRequest(c *gin.Context) {

	// must validate req's token first with fetched userID by auth middleware
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

	// fetch qpramas like limit,nextCursor from the request
	limitStr := c.Query("limit")
	nextCursorStr := c.Query("nextCursor")

	limit,err := strconv.Atoi(limitStr)
	if err!= nil {
		errMsg := "failed to parse limit,failed to fetch batch of Posts"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}
	if limit == 0 {
		errMsg := "limit not found or invalid,failed to fetch batch of Posts"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	// invoke method that belongs to the PostDBModel to fetch a batch request providing these
	batch,err := p.PostDbModel.GetFeedByBatches(limit,nextCursorStr)
	if err!= nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,utils.BatchResponse{
			Ok: false,
			Status: "Internal server error",
			HasMore: false,
			Batch: nil,
			NextCursor: "",
		})
		return 
	}
	
	// for nextCursor --> need to extract last postID which was fetched
	var nextCursor string
	var hasMore bool

	// posts are more than 0 , batch exists
	if len(batch) > 0  {
		//* extracting from batch[EL], El => lastEL <- -1 for index match,'s ID
		nextCursor = fmt.Sprint(batch[len(batch)-1].ID)
		hasMore = len(batch) == limit // when no of current batch elements equals to 
	}

	

	c.JSON(http.StatusOK,utils.BatchResponse{
		Ok: true,
		Status:"successfully loaded batch",
		HasMore: hasMore,
		Batch:batch,
		NextCursor: nextCursor,
	})



}



// postID fetched from url throgh the client - must be called on url - /api/feed/post/comments/:postid
func(cC*CommentController) GetCommentsCountByPostID(c *gin.Context) {

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

	
	// fetching postID from url for which ->post we are looking into to fetch comment count of that post
	idstr := c.Param("postid")
	postID,err:= strconv.Atoi(idstr)
	if err!= nil{
		// verbose err handeling needed here
		return
	}

	// single func for testing
	var commentCount struct {
		Count uint
	}
	res := cC.CommentDbModel.DB.QueryRow(`select count(*) as comments_count from comments c
    where post_id=$1  
    group by post_id`,uint(postID))
	
	err = res.Scan(
		&commentCount.Count,
	)

	if err!= nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK,gin.H{
				"CommentCount" : 0,
				"Ok" : true,
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError,gin.H{
			"CommentCount" : nil,
			"Ok" : false,
		})
		return 
	}

	c.JSON(http.StatusFound,gin.H{
		"CommentCount" : commentCount.Count,
		"Ok" : true,
	})
}

// get all posts of active client
func(p *PostController) GetAllPostsOfClient(c *gin.Context) {

	activeClientUserID := c.GetUint("user_id")
	if activeClientUserID  == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "Login expired",
		})
		return
	}

	posts,err := p.PostDbModel.GetPostsOfAnyUserByUserID(activeClientUserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
			Ok: false,
			Status: "failed to fetch user posts",
		})
		return 
	}

	//  if query was successfull , but no posts were retrieved
	if posts == nil {
		c.AbortWithStatusJSON(http.StatusNotFound,utils.ErrResponse{
			Ok: false,
			Status: "no post found",
		})
		return
	}

	c.JSON(http.StatusOK,utils.SuccessResponse{
		Ok: true,
		Status: "post fetched successfully for the active client",
		Data: posts,
	})

}



