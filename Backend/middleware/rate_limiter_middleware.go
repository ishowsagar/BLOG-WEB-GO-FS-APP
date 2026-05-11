package middleware

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/utils"
	"golang.org/x/time/rate"
)

// @types

//  type of client data struct that store client data & rate Limiter
type Client struct {
	ID uint 
	User models.User
	Limiter *rate.Limiter
}

// type that stores mutex and slice of *Client type elements {no limit untill not expanded}
type RateLimiter struct {
	mu sync.Mutex //* locker
	clients map[string]*Client
}

//  taking //* global instance of type RateLimiter
var ratelimit = RateLimiter {
	mu: sync.Mutex{}, // todo - check if not need literal
	clients: make(map[string]*Client),
}


//  rate limits malformed req's <- middleware function
func RateLimiterFunction() gin.HandlerFunc {

	
	return func(c *gin.Context) {
		
		//& fetch client userId, since auth is setting on the context req's as a key "user_id" on every key 
		retrievedUserID := c.GetUint("user_id")
		// if !exists {
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
		// 		Status : "token expired or invalid at limiter func",
		// 	})
		// 	return
		// }
		slog.Info("fetched user_id from the auth middleware","user_id",retrievedUserID)
		// fetching ip -> rate limit based off ip
		clientIP := c.ClientIP()

		// start lock once --> don't unlock it finished its work
		ratelimit.mu.Lock()

		// check if ip exists in clients map
		_,ipExists := ratelimit.clients[clientIP]
		
		if !ipExists {
			// ! if does not exists already
			// set clients map -> key being the ip ( it would be a string)- set rate Limiter value on it
			ratelimit.clients[clientIP] = &Client{
				// ! if there was no prior ip --> set it to only burst of 7 req per sec for 5 req
				Limiter: rate.NewLimiter(10,10), //(reqPerSecAllowed,burst)
			} 
		}

		// if already exists
		alreadyExistedIP := ratelimit.clients[clientIP] // fetch from the map
		ratelimit.mu.Unlock()


		//  checking if limit reached or not in ths already ip in the clients map
		if !alreadyExistedIP.Limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests,utils.ErrResponse{
				Status: "too many requests,try again later!.",
			})
			return
		}

		c.Next() // then only call next

	}
}

