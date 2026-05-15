package utils

import (
	"time"

	"github.com/ishowsagar/go-blog-web-application/models"
)

const DbTimeoutDuration = 5 * time.Second


//  type struct for General success responses
type SuccessResponse struct {
	// json tag with omit empty makes it optional
	Code uint `json:"code,omitempty"`
	Status string
	Data interface{}
	Ok bool
}

//  type struct for success responses
type ErrResponse struct {
	Status string
	Ok bool
}


//  type struct for dedicated comment success response
type CommentSuccessResponse struct {
	Code uint
	Ok bool
	Status string
	Comment string
	Data interface{}
}

//  type struct for dedicated post success response
type PostSuccessResponse struct {
	Ok bool
	Code uint
	Status string
	Post models.Post
}

//  type struct for dedicated user success response
type UserSuccessResponse struct {
	Ok bool
	Code uint
	Status string
	User models.User
}

//  type struct fpr sending like success response
type LikeSuccessResponse struct {
	Ok bool
	Code uint
	Status string
	Like models.Like
}

//  type struct for only success cached response
type CacheErrResponse struct {
	Ok bool
	Code uint
	Status string
}

// type struct for batch related response
type BatchResponse struct {
	Ok bool
	Status string
	HasMore bool
	NextCursor string
	Batch interface{}
}
