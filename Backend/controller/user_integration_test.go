package controller

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// ! Integration testing **//

//test IDEA - mocking a full http request with a preset payload
// 3. trace full request
// create httptest req with a preset payload
// make request and intercept response
// 4. do check explicit err check

// test out integrated req
func TestUpdatePassword_Integration(t *testing.T) {
	
	
	// *flow -
	
	// 1. setup gin server test mode
	gin.SetMode(gin.TestMode)
	
	// 2. setup temp router
	tempRouter := gin.New()
	
	// register handler with same userhandler method { since its integration means integrating it with preloaded data to trace request independently and testing in isolated environment}
	tempRouter.POST("/test/update-password",)




}