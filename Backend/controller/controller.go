package controller

// @types

//  stores all the handler type which -> stores controller method for corresponding functionality
type MasterController struct {
	// todo - add type which stores api models which stores methods related to db calls 
	UserController *UserController 
	PostController *PostController
	CommentController *CommentController  
	LikeController *LikeController
	FollowController *FollowController
	S3Controller *S3Controller
}

// func that creates instance of MasterController type w--> which stores all the corresponding controller methods types
func NewMasterController(userController *UserController,postController *PostController,commentController *CommentController, likeController *LikeController,followController *FollowController,s3Controller *S3Controller) *MasterController {
	return &MasterController{
		// todo - add later controller when they are done bareboning
		UserController: userController,
		CommentController: commentController,
		PostController: postController,
		LikeController : likeController,
		FollowController: followController,
		S3Controller: s3Controller,
	}
}