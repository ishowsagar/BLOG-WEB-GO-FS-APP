package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ishowsagar/go-blog-web-application/controller"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// func that belongs to the AuthMiddlewareInventory which -> already has redisClient
func(a *AuthMiddlewareInventory) SuperFastLoginMiddleware() gin.HandlerFunc {

	
	//& idea💡 -> only caLL next if cached failed

	// this func satisfy return type of gin.HandlerFunc
	return func(c *gin.Context) {

		var loginReqMiddleware controller.LoginRequest
		err := c.ShouldBindJSON(&loginReqMiddleware)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{
				Ok: false,
				Status: "Invalid request payload",
			})
			return  
		}

		_,_,err = a.RedisClient.GetCachedUser(c.Request.Context(),loginReqMiddleware.Email)
		if err!= nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
				Ok: false,
				Status: "lazing fast login attempt failed!.",
			})
			return
		}

		// if fetched -> login is successfull
		c.Next()

	}

}


// call this method to get cached user data -> login superfast 
// func(u *UserController) FastestLogin(c *gin.Context,email string) {

	
// 	cachedEmail,cachedPassword,err := u.RedisClient.GetCachedUser(c.Request.Context(),email)
// 	if err!= nil {
// 		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
// 			Ok: false,
// 			Status: "failed to fetched cached User for blazing fast login attemp!.",
// 		})
// 		return
// 	}
	
// }


