package routes

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ishowsagar/go-blog-web-application/controller"
	"github.com/ishowsagar/go-blog-web-application/middleware"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

func ServeRoutes(router *gin.Engine,masterController *controller.MasterController,config *utils.ENVConfig)  {

	// auth instances
	authMiddleware := middleware.NewAuthMiddlewareInventory(masterController.UserController.TokenDbModel,masterController.UserController.RedisClient)

	// * health check function to check health of API
	health := router.Group("/health")
	// health.Use(middleware.SlogLoggerMiddlewareFunction())
		{
			health.GET("/",func(c *gin.Context) {
				c.JSON(http.StatusOK,gin.H{
					"status":"OK",
					"message":"API IS RUNNING FINE⚡...",
				})
			})
			
		} 
	
	// * router Configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"POST","GET","PUT","DELETE"},
		AllowHeaders: []string{"Origin","Authorization","Content-type"},
		AllowBrowserExtensions:false, //! don't let mess with site headers or anything by installing scripts or like we did with auth header with mod header
		AllowCredentials: true,
		MaxAge: 1 * time.Hour,
	}))

	// * frontend Testing
	client := router.Group("/form")
	{
		client.POST("/register",masterController.UserController.RegisterUser) // just in case - need binded req payload of type registerReq
		client.POST("/login",masterController.UserController.LoginUser)
		client.POST("/password/reset",masterController.UserController.UpdateUserPassword)
	} 
	
	cached := router.Group("/cached") 
	cached.Use(middleware.LatencyCheckerMiddlewareFunction())
	{
		cached.POST("/register",masterController.UserController.RegisterUser)
		cached.POST("/login",masterController.UserController.SuperfastLogin)
	}
	
	api := router.Group("/api") 
	api.Use(authMiddleware.AuthMiddlewareFunction(config.JwtSecret))
	api.Use(middleware.RateLimiterFunction())
	// api.Use(middleware.SlogLoggerMiddlewareFunction())
	api.Use(middleware.LatencyCheckerMiddlewareFunction())
	{
		// user
		api.GET("/users/profile",masterController.UserController.FetchProfileData)
		
		// comment
		api.GET("/feed/comments/:postid",masterController.CommentController.LoadPostComments)
		api.GET("/feed/comment",masterController.CommentController.LoadAllCommentsAssociatedWithPostAndUsers)
		
		// * changing to post comment on post - by postID in url path, instead of sending json embedded postID
		api.POST("/post/comment/:postid",masterController.CommentController.PostComment)
		api.DELETE("/comment/delete",masterController.CommentController.DeleteCommentByUser)

		// feed
		api.GET("/feed",masterController.PostController.LoadFeed)
		api.GET("/post/count",masterController.PostController.GetPostCountByUserID)
		api.POST("/post/create",masterController.PostController.CreatePost)
		api.GET("/feed/post/:id",masterController.PostController.GetPostByID)
		api.DELETE("/post/:id",masterController.PostController.DeletePost)

		// like
		api.POST("/like",masterController.LikeController.UpdateLike)
		
		// follow
		api.POST("/users/follow/:followeeID",masterController.FollowController.FollowUser)
	} 
}