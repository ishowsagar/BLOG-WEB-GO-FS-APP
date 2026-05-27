package controller

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// @ types
type FollowController struct {
	FollowDbModel *services.FollowDBModel
	Pns *services.PushNotificationService
}

func NewFollowController(followDbModel *services.FollowDBModel,pns *services.PushNotificationService) *FollowController {
	return &FollowController{
		FollowDbModel: followDbModel,
		Pns: pns,
	}
}

// controller method for following user -> does followers count, following count
func(f *FollowController) FollowUser(c *gin.Context) {


	//  need both ids,followee is fetched from auth middleware token's string
	// fetch from url - api/users/follow/:followeeID
	// todo - fetch userID from passed param in url - of whom client would follow
	id := c.Param("followeeID")
	userToFollowID,err := strconv.Atoi(id)
	if err != nil {
		errMsg := "following id error,failed to follow user"
		code := http.StatusBadRequest
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	clientID := c.GetUint("user_id")
	if clientID == 0 {
		errMsg := "userID not found,failed to follow user"
		code := http.StatusUnauthorized
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	//  call method to follow user sequentially <- belongs to FollowDbModel

	followEntryID,err := f.FollowDbModel.FollowUserTransaction(clientID,uint(userToFollowID))
	if err != nil {
		if err == services.ErrFollowAlreadyExists {
			errMsg := "user already followed"
			// bug - was getting server eror but need to send this error to don't let user follow someone
			// fix - added statusConflict to fix that err when client follows same user repeatedly
			code := http.StatusConflict
			slog.Error(errMsg,"error",errMsg)
			c.AbortWithStatusJSON(code,utils.ErrResponse{
				Status: errMsg,
				Ok: false,
			})
			return
		}
		errMsg := "server error,failed to follow user"
		code := http.StatusInternalServerError
		slog.Error(errMsg,"error",errMsg)
		c.AbortWithStatusJSON(code,utils.ErrResponse{
			Status: errMsg,
			Ok: false,
		})
		return
	}

	// * sending follow payload to the methods of pns which -> redirects payload to the pns for publisher
	// test - it might show many rows as follower as sender would have meny entries associated with that id
	// fix - we to get exact entry using newly record created for the follow, need to grab entry's enntryID
	// fix2 -> easy, instead of getting entryID, grab all follows order by desc order limit res to1 -> fetch the latest entry
	followNotificationPayload,followNotiErr := f.FollowDbModel.GetFollowDetailsByFollowerUserID(followEntryID)
	if followNotiErr != nil {
		if followNotiErr == sql.ErrNoRows {
			slog.Error("user not found that was requested to follow","error",err)
			return
		}
		slog.Error("failed to get follow payload","error",err)
		return
	}
	f.Pns.RetryFollowWithTimeout(f.Pns,100 * time.Millisecond,followNotificationPayload)

	//  if succesffully done its work
	c.JSON(http.StatusOK,utils.SuccessResponse{
		Ok: true,
		Status: "followed User Successfully",
	})

}
