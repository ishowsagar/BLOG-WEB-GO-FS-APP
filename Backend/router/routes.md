// //! special token tester route -> pass token to check if auth working which checks for token
	// testAuth := router.Group("/test")
	// // todo - add rate limiter middleware function to test what happens if too many req bursted by client
	// testAuth.Use(middleware.RateLimiterFunction()) // rate limiter
	// testAuth.Use(authMiddleware.AuthMiddlewareFunction()) // auth mw that allow only if token exists on the req
	// testAuth.Use(middleware.LatencyCheckerMiddlewareFunction())
	// // todo - add mw to use this function to test as intended
	// // fix - testing with middleware function now
	// // first accessing route without token - //* tested and it worked good
	// // with token - //* tested and it was successfull
	// {
	// 	testAuth.GET("/auth/middleware/token", func(c *gin.Context) {
	// 		c.File("./static/avenger-meme-37.png")
	// 	})
	// }
	// router.Static("/static","./static")


	// // grouping Auth routes
	// auth := router.Group("/auth")
	// {
	// 	//  define routes in these braces with parent route path being "/auth"
	// 	auth.POST("/register",masterController.UserController.RegisterUser) // just in case - need binded req payload of type registerReq
	// 	auth.POST("/login",masterController.UserController.LoginUser)
	// 	// todo - need method to store tokens too and table migrations - add later
	// }




	// // ! for comments
	// //  must add seperately middlewares for groups, as they behave isolated routes
	
	// cmt := router.Group("/comment")
	// cmt.Use(authMiddleware.AuthMiddlewareFunction())
	// {
	// 	cmt.GET("/feed",masterController.CommentController.LoadAllCommentsAssociatedWithPostAndUsers)
	// 	cmt.POST("/",masterController.CommentController.PostComment)
	// }


	// // ! for posts

	// post := router.Group("/post")
	// // post.Use(authMiddleware.AuthMiddlewareFunction())
	// {
	// 	post.GET("/feed",masterController.PostController.LoadFeed)
	// 	post.POST("/",masterController.PostController.CreatePost)
	// 	post.DELETE("/:id",masterController.PostController.DeletePost)
	// }


	// // ! for likes - must be proctected by auth middleware
	// like := router.Group("/like")
	// // bug - got error for not finding id, cause we forgot to add auth middleware which -> checks for header token and parse it to fetch id from the token claims
	// // like.Use(authMiddleware.AuthMiddlewareFunction())
	// {
	// 	// todo - add Likecontroller in masterController
	// 	like.POST("/",masterController.LikeController.UpdateLike)	
	// }



<!--# testing follow route -->
<!-- todo - must add user check first before doing this atomic transaction -->
## followee - client who would be logged in and follow someone (should exist)
>{
"name" : "followee",
"email" : "followee@gmail.com",
"password": "changedPassword"
}

id =38 
following_count = 0

> {
    "Ok": true,
    "status": "login successfull",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcnkiOiIyMDI2LTA1LTA5VDE3OjQ2OjM0LjQ5Mjc3ODkrMDU6MzAiLCJ1c2VyX2lkIjozOH0.T5uM4s5bWdSZuFccNsXRC3UQIi_zSbXvcCoDglwMVy8",
    "user": "followee"
}


## user to follow
userID - 17, email - sagar1@gmail.com
followers_count = 0
