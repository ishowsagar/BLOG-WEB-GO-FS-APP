package controller

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// @ types

// type that stores s3BucketModel which -> stores methods which called by client to do the s3 bucket operations like api service
type S3Controller struct {
	S3BucketModel *services.S3BucketModel
}

// func that returns the instace of type S3Controlle which > stores Controller method for serving uploads n whatnot
func NewS3Controller(s3BucketModel *services.S3BucketModel) *S3Controller {
	return &S3Controller{
		S3BucketModel: s3BucketModel,
	}
}


// handler method that when invoked -> uploads the image to the s3
func(s *S3Controller) HandleUploadImageStream(c *gin.Context) {

	// 1 - fetch active client ID
	userID := c.GetUint("user_id")
	// fetching active clientID which -> set by AuthMiddleware from client's token
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.S3UploadErr{
			Ok: false,
			Error: "user not found",
		})
		return
	}

	//* converting id to string type using Sprintf
	userIDStr := fmt.Sprintf("%v",userID)

	// 2 - return header of file matching this key "avatar" set on it\
	// /avatar=@C:\Users\asus\Pictures\sample.png"  <- this header is checked if it is there it can't fetch the file

	// ! when req is made on this url --> check if req has that -f header means the path where image is coming from and it has "avatar" in header, get its metadata 
	// fileHeader,err := c.FormFile("avatar")  // checks if its coming with "avatar" in header
	// // what we get is -> metdatam {filename,size,type}
	//  //*as file comes in multi-part file type
	// if err != nil {
	// 	slog.Error("FormFile failed", "raw_error", err)
	// 	c.AbortWithStatusJSON(400,utils.S3UploadErr{
	// 		Ok: false,
	// 		Error: "No image found in form-data payload",
	// 	})
	// 	return
	// }


	// 3 - opening recieved file in network pipe (io.Reader),not locally on disk -> so it opens a direct stream to s3 when service called
	//* since Files are recieved in parts -> needs a streamline pipe to recieve data in parts and send in parts too 
   fileStream := c.Request.Body
	// if err != nil {
	// 	c.AbortWithStatusJSON(400,utils.S3UploadErr{
	// 		Ok: false,
	// 		Error: "Unable to open file reader pipeline",
	// 	})
	// 	slog.Info("c.Abort err does not stop method from furthur executionz")
	// 	return
	// }

	defer fileStream.Close()
	fallbackFilename := "avatar.png"
	
	// bug - since we are opening stream using c.R.Body it does not tell when to end stream, so it crashed out we need to provide length to fix it
	// fix - add content-length to let it know when to finish stream if you are not using c.FormFile() which fetches provided file from and knows when to end stream once done
	fileSizeINBytes := c.Request.ContentLength

	// 4 - pass everything to s3BucketModel service which pipes the upload to s3
	retrievedUploadedImageURL,err := s.S3BucketModel.UploadImageStream(
		c.Request.Context(),
		"profiles", //& virtual dir where this would be stored
		userIDStr, //& userID
		fallbackFilename, //& filename from header
		fileStream, //&reader
		fileSizeINBytes,
	)

	
	if err != nil {
		slog.Error("s3 upload err","error",err)
		c.AbortWithStatusJSON(400,utils.S3UploadErr{
			Ok: false,
			Error: "failed to upload image to the aws s3 storage",
		})
		return
	}
	
	// inserting imageURL into the db mapped to the user 
	err = s.S3BucketModel.InsertImage(userID,retrievedUploadedImageURL)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("status","error",err)
			c.AbortWithStatusJSON(http.StatusInternalServerError,utils.S3UploadErr{
				Ok: false,
				Error: "no rows were inserted",
			})
			return
		}
		slog.Error("failed to insert profile picture into the db","error",err)
		c.AbortWithStatusJSON(http.StatusInternalServerError,utils.S3UploadErr{
			Ok: false,
			Error: "failed to upload image url in the db",
		})
		return
	}

	slog.Info("successfully stored retrieved s3 image uploaded url in db","userID :",userID)

	
	// if successfully uploaded and retrieved url ✅
	// todo - store image url mapped to user in DB 
	c.JSON(http.StatusOK,utils.S3UploadSuccessResponse{
		Ok: true,
		Status: "profile picture uploaded successfully🚀",
		ImageURL:retrievedUploadedImageURL,
	})

}

// handler method to fetch pfp exclusively
func(s *S3Controller) GetProfilePictureBucketURl(c *gin.Context) {

	// fetch active clientID -> whose pfp has to be fetched
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.S3UploadErr{
			Ok: false,
			Error: "user not found",
		})
		return
	}

	//  fetch url stored in db for that userDI in db through db call
	pfpURL,err := s.S3BucketModel.GetStoredPFPImageURL(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable,utils.S3UploadErr{
			Ok: false,
			Error: "failed to get pfp",
		})
		return
	}

	if pfpURL == nil {
		c.AbortWithStatusJSON(http.StatusNotFound,utils.S3UploadErr{
			Ok: false,
			Error: "pfp not found",
		})
		return
	}

	// if successfully retrieved stored s3url for this user, send to the client
	resolvedURL := *pfpURL
	c.JSON(http.StatusOK,gin.H{
		"Ok" : true,
		"ImageURL" : resolvedURL,
	})
}