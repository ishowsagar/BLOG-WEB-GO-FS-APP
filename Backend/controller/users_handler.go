package controller

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	dumper "github.com/goforj/godump"

	"github.com/gin-gonic/gin"
	cache "github.com/ishowsagar/go-blog-web-application/Cache"
	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// @ types
type UserController struct {
	// todo - add interfaces implementation later
	UserDbModel *services.UserDBModel
	TokenDbModel *services.TokenDBModel
	RedisClient *cache.RedisCacheClient // just in - client holds all the operations for caching
}

//  func that returns instance of type UserController - which stores userDbModel -> stores all user related meths
func NewUserController(userDbModel *services.UserDBModel,tokenDbModel *services.TokenDBModel,redisClient *cache.RedisCacheClient) *UserController {
	return &UserController{
		UserDbModel:userDbModel,
		TokenDbModel: tokenDbModel,
		RedisClient: redisClient,
	}
}

// method that belongs to the UserController type - registers user
func(u *UserController) RegisterUser(c *gin.Context) {

	var registerReq RegisterRequest
	err := c.ShouldBindJSON(&registerReq)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest,gin.Error{
			Err: err,
		})
		return
	}
	
	
	//  todo - need to store hashed pass in User that we are registering
	// fixed - added hashed pass 
	
	// todo - add method check for checking if user already registered Using email
	// fixed - added check
	existingUser,err := u.UserDbModel.GetUserByEmail(registerReq.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			existingUser = nil // set it to nil as there was no user but query ran successfully
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError,gin.H{
					"error" : "lookup failed",
				})
				return
			}
		}
		if existingUser != nil {
			c.AbortWithStatusJSON(http.StatusConflict,gin.H{
				"error" : "user already exists",
			})
			return
		}
		
		// todo - need to store hashed password
		// & generating hash for storing hashed pass in the db, not just plain string
		hash := utils.SetHashedPassword(registerReq.Password)

		// creating instance of user that has to be created into the db from requested payload
		user := models.User{
			Name: registerReq.Name,
			Email: registerReq.Email,
			Password: string(hash), // bug --> might bug cause we storing as string
			// fixed --> no problems so far

			// bug - while creating user since nothing were taking care of *_count fields, so we did insert method to add default 0
			Bio: registerReq.Bio,
			Username: registerReq.Username,
			Nickname: registerReq.Nickname,
		}

	//  if req binds correct data ✅
	createdUser,err := u.UserDbModel.CreateUser(&user)
	if err!= nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{
			Status: "server error,failed to register user!.",
		})
		return
	}

	// struct that stores cahedUser data 
	cachedUser := models.User{
		Email: createdUser.Email,
		Password: registerReq.Password, //* storing plain password for now
	}
	// todo - cache user data into redis db for faster logins
	err = u.RedisClient.SetCachedUser(c.Request.Context(),cachedUser.Email,cachedUser.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
			Ok: false,
			Status: "failed to cache user data",
		})
		return
	}

	// if successfully cached it, return this response✅✅
	
	//  sending response to client
	dumper.DumpJSON("postgres entry :",createdUser)
	dumper.DumpJSON("cachedUser :",cachedUser)
	c.JSON(http.StatusOK,gin.H{
		"Ok" : true,
		"Code":http.StatusCreated,
		"Status": "Successfully resgistered User.",
	})
}


// login user
func(u *UserController) LoginUser(c *gin.Context) {

	// req should bind loginReq type payload
	var loginRequest LoginRequest

	err := c.ShouldBindJSON(&loginRequest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{Status: "Invalid request payload."})
		return 
	}
	// check if user exists through email method
	userFound,err := u.UserDbModel.GetUserByEmail(loginRequest.Email)
	if err != nil {
		// todo - less verbose err responses
		c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{Status: "lookup failed"})
		return 
	}
	// if no user data retrieved from query 
	if userFound == nil {
		c.AbortWithStatusJSON(http.StatusNotFound,utils.ErrResponse{Status: "user not found."})
		return
	}

	// // ** CACHE REDIS DB CHECKS **//
	// // todo - fetch stored cached user from redis db before postgres does its job
	// // if found - login

	// _,_, cachingUserErr := u.RedisClient.GetCachedUser(c.Request.Context(),userFound.Email)
	// if cachingUserErr != nil {
	// 	c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
	// 		Ok: false,
	// 		Status: "failed to fetch cached user data",
	// 	})
	// 	return
	// }

	// // if data is found -> login with token✅


	// // .... end .... //
	
	

	//  if yes found ✅, create token
	// todo - check for hashedPass not just plain pass - cause now we are storing hash,not just plain text
	authorized,err := utils.CheckHashedPass(loginRequest.Password,[]byte(userFound.Password))
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized,gin.Error{
			Err: err,
		})
		return
	}


	


	//  if passes dont match
	if !authorized  {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			//todo less verbose err response to the client
			Status: "wrong password", // inentionally setting for verbose debugging
		})
		return
	}

	newExpiry := time.Now().UTC().Add(24 * time.Hour)
	
	plainTokenString,err := utils.GenerateToken(userFound.ID) // passing found user id to generate stateless token
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{Status: err.Error()})
		return 
	}

	// todo - hash tokenString too,store hashed token
	// fix - hashed byte
	// todo - add seperate function for creating hashed token using seperate secret signing key 
	hashedTokenByte,err := utils.HashToken(*plainTokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{Status: err.Error()})
		return 
	}
	
	// slog.Info("token created successfully","userID",userFound.ID)
	
	token := models.Token{
		Hash: string(hashedTokenByte),//* storing hashed Token
		UserID: userFound.ID,
		Expiry: newExpiry, //* 24hour expiry time for the token
	}
	// first check if it is not already stored -> cause that would overide expiry time
	existingToken,err := u.TokenDbModel.GetTokenByUserID(userFound.ID)
	if err!= nil {
		//! don't do statusInternalErr -> it will stop furthur execution cause err could be no rows
		if err == sql.ErrNoRows {
			// if query successfull --> but returned no token
			existingToken = nil
			//  no return -> must continue to insert
			} else {
				slog.Error("failed to lookup","error",err)
				c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{
					Status : "lookup failed",
				})
				return
			}
		}
		
		// todo - after token is generated, cache token into redis db
		// todo - now after setting in cache, on every authMiddleware, fetch hashes for client from cache first
		// normally adding token into the cache db -. layer the flow logic just lil after it
		cachingTokenErr := u.RedisClient.SetCachedToken(token.UserID,token.Expiry,token.Hash)
		if cachingTokenErr!= nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,utils.CacheErrResponse{
				Ok: false,
				Code: http.StatusInternalServerError,
				Status: "failed to cache user login data",
			})
			return
		}

		// atp redis would have stored token data into cache -> yeah in form of var struct,not concrete vars val✅


	// if token already exists and not nil -> update it, otherwise insert a new one
	if existingToken != nil {
		// if already exists - update it
		err = u.TokenDbModel.UpdateTokenIfExists(newExpiry,userFound.ID,string(hashedTokenByte))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{Status: err.Error()})
			return
		}
	} else {
		// if not existed , create new and insert into the db
		err = u.TokenDbModel.InsertToken(token)
		if err!= nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{Status: err.Error()})
			return 
		}
	}

	//  send token response
	c.JSON(http.StatusOK,gin.H{
		"Ok" : true,
		"user" : userFound.Name,
		"token" : plainTokenString,
		"status" : "login successfull",
	})

}

// for blazing fast logins with cache only
func(u *UserController) SuperfastLogin(c *gin.Context) {

	// get payload
	var fastloginReq LoginRequest
		err := c.ShouldBindJSON(&fastloginReq)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{
				Ok: false,
				Status: "Invalid request payload",
			})
			return  
		}


	// check stored cache from redis db
		cachedEmail,cachedPassword,err := u.RedisClient.GetCachedUser(c.Request.Context(),fastloginReq.Email)
		if err!= nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
				Ok: false,
				Status: "Blazing fast login attempt failed!.",
			})
			return
		}

	// pass validation
		if cachedPassword != fastloginReq.Password {
			c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
					Ok: false,
					Status: "Invalid credentials",
				})
				return
		}
	
	// if successfully got them -> login user (non-token based) first,add token later
		c.JSON(http.StatusOK,gin.H{
			"Ok": true,
			"email" : cachedEmail,
			"Status": "blazing fast login is successfull⚡",
		})
	
}

// controller method that updates user password
func(u *UserController) UpdateUserPassword(c *gin.Context) {


	// fetch userId for
	var updateRequest struct {
		Email string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	err := c.ShouldBindJSON(&updateRequest) 
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{
			Status: "invalid payload request",
			Ok: false,
		})
		return 
	}

	

	//  if client correct format of request payload

	//  first check if user exists for that mail
	foundUser,err := u.UserDbModel.GetUserByEmail(updateRequest.Email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
				Status: "failed to check if user exists or not",
				Ok: false,
	})
	return	
	}

	if foundUser == nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
				Status: "user not found",
				Ok: false,
	})
	return	
	}

	//  if user is found and exists, not nil
	//  if userFound, update pass with new hash generated with password recieved
	// hash password and generate new one for it 
	updatedPasswordHash := utils.SetHashedPassword(updateRequest.Password)
	// and store in db to update pass
	updated,err := u.UserDbModel.ResetUserPassword(string(updatedPasswordHash),updateRequest.Email)
	if err != nil || !updated {
		c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{
		Status: "failed to reset password",
		Ok: false,
	})
	return
	}
	// if resetted delete old data stored in cache related to the token of that user
	err = u.RedisClient.DeleteCachedToken(foundUser.ID)
	if err != nil  {
		c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{
		Status: "failed to wipe cached token",
		Ok: false,
	})
return
	}
	// if successfully updated password
	c.JSON(http.StatusOK,utils.CommentSuccessResponse{
		Status: "successfully resetted the password",
		Ok: true,
	})
	}


func(u *UserController) FetchProfileData(c *gin.Context) {

	// fetching client userID from req -> as set by auth middleware on every req by proccessing header token
	userID := c.GetUint("user_id")
	if userID == 0 {
		errMsg := "userID not found,failed to get profile data"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	// client userID fetched successfully
	// bug - it was throwing empty strings on bio,user~nickname when route invoked,cause this was not returning user with those fields
	// fixed -now corresponding repo method has taken care of these fields and returning them in user struct it is returning now.
	foundUser,err := u.UserDbModel.GetUserByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
				Status: "user not found",
				Ok: false,
		})
		return
		}
		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
				Status: "failed to check if user exists or not",
				Ok: false,
		})
		return	
	}


	//  if userFound query was successfull and resulting struct also not nil
	c.JSON(http.StatusFound,utils.UserSuccessResponse{
		Ok: true,
		Status: "user data fetched successfully",
		Code: http.StatusFound,
		User: *foundUser,
	})
}



// for fetching data from Url param ID
func(u *UserController) FetchProfileDataByURlParamID(c *gin.Context) {

	// fetching client userID from req -> as set by auth middleware on every req by proccessing header token
	userID := c.GetUint("user_id")
	if userID == 0 {
		errMsg := "userID not found,failed to get profile data"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	// fetching userID from url param
	profileeIDStr := c.Param("userid") // client would request on url passing userid in url params
	profileeID,err := strconv.Atoi(profileeIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{
				Status: "invalid user id",
				Ok: false,
		})
		return	
	}

	// client userID fetched successfully
	foundUser,err := u.UserDbModel.GetUserByUserID(uint(profileeID))
	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusNotFound,utils.ErrResponse{
				Status: "user not found",
				Ok: false,
		})
		return
		}
		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
				Status: "failed to check if user exists or not",
				Ok: false,
		})
		return	
	}


	//  if userFound query was successfull and resulting struct also not nil
	c.JSON(http.StatusFound,utils.UserSuccessResponse{
		Ok: true,
		Status: "user data fetched successfully",
		Code: http.StatusFound,
		User: *foundUser,
	})
}

// method that belongs to the userController type -> which clears the cached token from redis db -> must be invoked with Auth header
func(u *UserController) WipeCachedToken(c *gin.Context) {

	
	//  fetch from active client's userID set by auth middleware on every req
	userID:= c.GetUint("user_id")
	
	// todo - add proper err return later
	if userID == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	
	// invoke method to clear db -> need userID
	err := u.RedisClient.DeleteCachedToken(userID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK,utils.CommentSuccessResponse{
		Ok: true,
		Status: "successfully deleted cached token",
	})

}


// since it will be a search -> 'name' would be in qparamy
func(u *UserController) FindUsersByNAME(c *gin.Context) {

	// active client validation with auth middleware which -> checks for token in req's header
	userID := c.GetUint("user_id")
	if userID == 0 {
		errMsg := "userID not found,failed to search users"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	// search query -> ?name=X
	name := c.Query("name")

	// fetch users
	usersFound,err := u.UserDbModel.FindUsersByName(name)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
			Ok: false,
			Status: "failed to search user",
		})
		return 
	}

	if usersFound == nil{
		c.AbortWithStatusJSON(http.StatusNotFound,utils.ErrResponse{
			// todo - had to check to send ok as true if not found but search was a success operation
			Ok: false,
			Status: "no user found",
		})
		return
	}

	// if successfully fetched all users
	c.JSON(http.StatusOK,utils.SuccessResponse{
		Ok: true,
		Status: "User found",
		Data: usersFound,
	})
}


// get full fledged profile data  of active client
func(u *UserController) FetchFullProfileData(c *gin.Context) {

	activeClientUserID := c.GetUint("user_id")
	if activeClientUserID  == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "Login expired",
		})
		return
	}

	profileData,err := u.UserDbModel.FetchFullProfileData(activeClientUserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
			Ok: false,
			Status: "failed to fetch user profileData",
		})
		return 
	}

	//  if query was successfull , but no profileData were retrieved
	if profileData == nil {
		c.AbortWithStatusJSON(http.StatusNotFound,utils.ErrResponse{
			Ok: false,
			Status: "no profile data has found",
		})
		return
	}

	c.JSON(http.StatusOK,utils.SuccessResponse{
		Ok: true,
		Status: "successfully retrieved profile data for the active client",
		Data: profileData,
	})

}

// controller method which -> retrieves all the profiles of what users would have followed
func(u *UserController) GetFollowedUserProfiles(c *gin.Context) {
	activeClientUserID := c.GetUint("user_id")
	if activeClientUserID  == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "Login expired or invalid token",
		})
		return
	}

	profilesData,err := u.UserDbModel.FetchAllFollowingUsers(activeClientUserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.ErrResponse{
			Ok: false,
			Status: "failed to fetch followed profiles data",
		})
		return 
	}

	//  if query was successfull , but no profileData were retrieved
	if profilesData == nil {
		c.AbortWithStatusJSON(http.StatusNotFound,utils.ErrResponse{
			Ok: false,
			Status: "no profiles data has found",
		})
		return
	}

	c.JSON(http.StatusOK,utils.SuccessResponse{
		Ok: true,
		Status: "successfully retrieved profile data for the active client",
		Data: profilesData,
	})

}

// @ next Implementation
// - add tx atomic method to fetch user to follow -> render everything related to him, => done with fetching data of that user from users from its id ✅
//  - when a comment is posted -> fetch user_id and from there user data  -> display commentor name => done by learning and implementing joining c with user as same id so from one big joined table -> fetch name ✅
//  - post like count on feed post, not each post => done with joint tables✅


// @ Advance routing
// - seperate sign up flow pages
// - search and display users in trays in search.jsx - need api corresponding models methods to fetch user profiles - done ✅
// - follow unfollow -> by adding a seperate nested injected page - when searched - clicked - load page of his info and follow unfollow there - Done ✅
// - for upper thing need single atomic transaction if needed  - done with multi staged events✅
// - fetch posts associated with active client ID and also for profile -> fetching for eachProfile -> redirected url where data is loaded -> fetch for that userID
// - fetch users profiles which active client has followed to be loaded in -> followings + messages tab