package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	cache "github.com/ishowsagar/go-blog-web-application/Cache"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// @ types
type AuthMiddlewareInventory struct {
	TokenModel *services.TokenDBModel
	RedisClient *cache.RedisCacheClient
}

//  func that returns instance of type AuthMiddlewareInventory which -> stores tokenModel which stores all methods on it
func NewAuthMiddlewareInventory(tokenModel *services.TokenDBModel,redisClient *cache.RedisCacheClient) *AuthMiddlewareInventory {
	return &AuthMiddlewareInventory{
		TokenModel: tokenModel,
		RedisClient: redisClient,
	}
}

// used as mw function --> need to return funciton of type gin.handlerFunc

// todo - need handler to fetch id from context, so it must have token in order to do something, else failed --> need to register
//  func that checks on every req -> if has token --> set on request for ready to fetch by handlers
func(a *AuthMiddlewareInventory) AuthMiddlewareFunction(jwtSecretSigningKey string) gin.HandlerFunc {
	return func(c *gin.Context) {

		//**      Apply to all requests      **//


		// scene 1 - directly sending plain token - no compare check
		
		//  scene 2 - fetched token from header - need to compare with stored hash and check for expiry

		//  retrieve tokenString from header - "Authorization" 
		tokenStr := c.GetHeader("Authorization")
		
		// todo - tStr validation
		//  fixed -> added validation and it is working✅✅
		if tokenStr == "" {
			slog.Warn("authorization header not found on the request!.")
			c.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{
				"error" :"token not found or header unavailable on the request",
			})
			return
		}
		// test - cause everytime it imports env vars and all that for every reqis heavy task, we will load once and pass as arg to all packages function which needs it

		
		// load env vars , we would load from passed Config for faster loads
		// jwtSecretSigningKey :=  utils.GetConfig().JwtSecret
		
		// loading env file
		// err := godotenv.Load()
		// if err != nil {
		// 	slog.Error("failed to load .env file","error",err)
		// 	return 
		// }
		
		// accessing env protected variables
		// jwtSecretSigningKey := os.Getenv("JWT_SECRET_KEY")
		
		//  parse to token with jwt's libraty parseToken meth with "secretKey" that was used to sign token
		token,err :=jwt.Parse(tokenStr,func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecretSigningKey),nil // passing key which used to sign that str
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Status: err.Error(),
			})
			return
		}
		if token == nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
				Status: "invalid token",
			})
			return
		}
		
		// fetch mapped data from token claims map
		tokenClaims,ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
				Status: "invalid token claims",
			})
			return
		}
		userID := tokenClaims["user_id"].(float64) // extracting user_id from claims <- set when created token
		


		// * befor calling next -> need to check if it is valid or not
		// todo - add expiry check by retrieved stored token from db and if expired --> won't process next request
		// fixed - added expiry check, return err if token time has passed more time than 24h



		

	

		//* checking if they match -> client would have only plain token <- checking against stored hash of it
		hashedIncomingToken,err := utils.HashToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{
				Status: "failed to hash token",
			})
			return
		}


		// ** CACHED REDIS DB FASTER PRIORITY CHECKS **//

		//todo - before checking from postgres db, check from redis cache first
		// this method returns unmarshalad data struct values which is being returned by the method 
		cachedHash,cachedExpiry,cachingErr := a.RedisClient.GetCachedToken(uint(userID)) // userID is fetched from active client's token's mapped claims
		if cachingErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,utils.CacheErrResponse{
				Ok: false,
				Status: "failed to fetch user data from cache in auth middleware block",
				Code: http.StatusInternalServerError,
			})
			return
		}

		// if successfully retrieved user data struct from cache ✅ -> first all expiry checks and hash checks done with cache, if miss then concrete postgres db only
		if cachedHash != string(hashedIncomingToken) {
			c.AbortWithStatusJSON(http.StatusUnauthorized,utils.CacheErrResponse{
				Ok: false,
				Status: "token mismatch,cache block", // for debugging purpose verbose err for now, change later once implemented fully
				Code: http.StatusUnauthorized,
			})
			return
		} 

		//  if time untill now 24hrs and time is elasped after this time -> declare it as 'expired'
		if time.Now().After(cachedExpiry) {
			c.AbortWithStatusJSON(http.StatusUnauthorized,utils.CacheErrResponse{
				Ok: false,
				Status: "token expired,cache block",
				Code: http.StatusUnauthorized,
			})
			return
		}

		// ...... end of cache check ...... //


		// ** CONCRETE POSTGRESS DB CHECKS **//

		// todo - since auth middleware checks for -> if there is entry existing in tokens db where token exists for this userID,
		// retrieving token stored in db -> using method that belongs to the type TokenModel
		hashedToken,err := a.TokenModel.GetTokenByUserID(uint(userID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{
				Status: "hashed token not found in db,failed to authenticate client's req",
			})
			return
		}
		if string(hashedIncomingToken) != hashedToken.Hash {
			c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
				Status: "token mismatch",
			})
			return
		}

		// if passed PlaintokenStr matches that stored in Db --> means user is valid

		// now checking expiry
		if time.Now().After(hashedToken.Expiry) {
			c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
				//  later would need less verbose error
				Status: "token expired",
			})
			return
		}

		//  otherwise if not expired, let client what was intended to ✅
		// set new context state on req with extracted id from mapped data
		// bug - as we retrieved value from map claim's user id -> in float64 type --> need to parse to uint to set uint which needed by req
		c.Set("user_id",uint(userID)) // fixed - now successfully working
		//  call the next method to satisfy the type of this function return 
		c.Next()

	}
}