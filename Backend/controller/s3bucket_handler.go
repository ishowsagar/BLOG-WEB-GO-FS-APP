package controller

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/ishowsagar/go-blog-web-application/services"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// @ types

// type that stores s3BucketModel which -> stores methods which called by client to do the s3 bucket operations like api service
type S3Controller struct {
	S3BucketModel *services.S3BucketModel
	PostDbModel *services.PostDBModel
}

// func that returns the instace of type S3Controlle which > stores Controller method for serving uploads n whatnot
func NewS3Controller(s3BucketModel *services.S3BucketModel,postDbModel *services.PostDBModel) *S3Controller {
	return &S3Controller{
		S3BucketModel: s3BucketModel,
		PostDbModel: postDbModel,
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

	// fetch requested profile id if provided, otherwise fall back to the authenticated user
	requestedUserID := c.Query("userid")
	userID := c.GetUint("user_id")
	if requestedUserID != "" {
		var parsedID uint
		if _, err := fmt.Sscanf(requestedUserID, "%d", &parsedID); err != nil || parsedID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, utils.S3UploadErr{
				Ok: false,
				Error: "invalid userid",
			})
			return
		}
		// if query sending id of profile userID, set userID to be of that, instead activeClientID
		userID = parsedID
	}
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


// ** POSTS **//

// handler method that uploads fileImage to the bucket via bucketInsertor method call
func(s *S3Controller) HandlePostsImageStream(c *gin.Context) {


	// client needs a i/o - reader for uploading multi-parted file to the bucket
	// io.Reader gives a small container which loads data into it and send to bucket and get back tells how much it poured into to the bucket...keep going unitll it does not hit EOF
	// once data emptied -> closes stream
	// this is how large data files are uploaded by sending chunks of data in containers in stream by the reader and keeps doing untill done

	
	// fetching userId from the request's token via auth middleware
	clientID := c.GetUint("user_id")
	if clientID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,utils.ErrResponse{
			Ok: false,
			Status: "Access denied - login expired or invalid token",
		})
		return
	}


	// * fetching postID from the query -> so it frontend will call on passing postID in the url
	postIDStr := c.Query("postid")
	if postIDStr == "" {
		// tell client to must include postID in the url else it will not able to store url mapped to the post
		c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{
			Ok: false,
			Status: "post id not found in the query,must sent with key being 'postid'",
		})
		return

	}

	postIDInt,err := strconv.Atoi(postIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{
			Ok: false,
			Status: "wrong post id,please pass correct 'postid' in the query",
		})
		return
	}
	postID := uint(postIDInt)
	// str formatter to convert it into str for obj key
	clientIDStr:=fmt.Sprintf("%v",clientID) // must pass val via %v val placeholder

	
	// before upload,must check if active clients own the post
	createdPost,err :=s.PostDbModel.GetPostbyID(postIDInt)
	if err!= nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{
			Ok: false,
			Status: "failed to get post",
		})
		return
	}
	if createdPost == nil {
		c.AbortWithStatusJSON(http.StatusNotFound,utils.ErrResponse{
			Ok: false,
			Status: "post not found;rolled back uploaded image and removed from the bucket🚨",
		})
		return
	}

	// & if post exists but checking is active is the owner
	if clientID != createdPost.UserID {
		c.AbortWithStatusJSON(http.StatusForbidden, utils.ErrResponse{Ok:false, Status:"client is not post owner🚨🚨"})
   		 return
	}

	oldImageURL := createdPost.ImageSource // old could either be "default.png or ''"
	
	//stream of type io.Reader -> bears a small container ->for sending data in chunks
	unhandledChunkStream := c.Request.Body // later we will develop own reader stream
	// -f files data are recieved in parts/chunks, not in single atomic load
	defer unhandledChunkStream.Close()
	dataLength := c.Request.ContentLength
	
	//! incoming data validation - using multireader to check if incoming data is only images
	initialBuffer := make([]byte,512) // initial buffer of size 512bytes
	
	const maxSizeAllowed = 2*1024*1024 //2mb
	// assigning max byte to be read
	unhandledChunkStream = http.MaxBytesReader(c.Writer,unhandledChunkStream,maxSizeAllowed)

	// if data length is more than maxallowed size,send client err status
	if dataLength > maxSizeAllowed {
		slog.Error("failed to read incoming data","error","file size too large")
		c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{
			Ok: false,
			Status: "file too large, it should not exceed => size > 2mb",
		})
		return	
	}

	// if content length validated and passed size limit ✅ send to client

	// read atleast data it could read but under iB size
	chunkLength,err :=io.ReadAtLeast(unhandledChunkStream,initialBuffer,1) // read from body but upto first buffer limit, whatever read -> store in the buffer
	// it returns n~cL as till what byte number data is successfully read
	if err!= nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		// handeling read err of the buffer,//! but it was eof or like normal err but could not read this file
		slog.Error("failed to read incoming data stream","error",err)
		c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{
			Ok: false,
			Status: "failed to read incoming data stream",
		})
		return	
	} 

	// checking if initial buffer read empty bytes as client uploaded nothing
	if chunkLength == 0 {
		slog.Error("client sent empty file to upload","error","no data recieved in the stream")
		c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{
			Ok: false,
			Status: "file cannot be empty",
		})
		return
	}

	// checking upto what length of buffer data which was filled has what "content-type"
	incomingDataType := http.DetectContentType(initialBuffer[:chunkLength]) //* checking what type of data is coming till this buffer limit
	slog.Info("successfully recieved file initial buffer's content type","contentType:",incomingDataType)
	allowedDataTYPES := map[string]bool {
		// ! bug - if file is sent via multlipart it gives err, but if sent via file in body, it will success as then it attached file content tyoe automaticaly which then here is what we are validating
		"image/jpeg" :true,
		"image/png" :true,
		"image/webp" :true,
	}

	// if incoming data type could not be validated 
	if !allowedDataTYPES[incomingDataType] {
		// if val of this passed data type is false or does exists as false, then not allowed
		c.AbortWithStatusJSON(http.StatusNotAcceptable,utils.ErrResponse{
			Ok: false,
			Status: "Invalid file type",
		})
		return
	}


	// attach buffer to the newly created stream if passed validation check
	handledMutatedStream := io.MultiReader(bytes.NewReader(initialBuffer[:chunkLength]),unhandledChunkStream)


	// bug - if by chance file is not uploaded the correct way, it would prints out "-1" data length from content length which could entirely crash our sdk
	// fix - Read at once from the mutated reader total length it has sent in containers as its what it all about-> sending chunk in smal container giving the metrics

	var verdictedDataLength int64 = c.Request.ContentLength
	if verdictedDataLength <= 0 {
		// if content length less than 0 
		slog.Info("Content-Length is unknown (-1) or 0. Computing actual size safely in memory...")
		actualBytesOFData,readErr :=io.ReadAll(handledMutatedStream)
		if readErr != nil {
			slog.Error("count not able to read total byte size of the uploaded content","error",readErr)
			c.AbortWithStatusJSON(http.StatusBadRequest,utils.ErrResponse{
				Ok: false,
				Status: "failed to get stream data length",
			})
			return 
		}

		// since bytes data are in [], we can retrieve length from it
		n := len(actualBytesOFData)
		// setting it to be the final content length
		verdictedDataLength = int64(n)

		// redefining stream reader with this byte of data
		handledMutatedStream = bytes.NewReader(actualBytesOFData)

	}

	// validating content length -> not allowing it to be execeeded than allowed
	if verdictedDataLength > maxSizeAllowed {
		slog.Error("failed to read incoming data","error","file size too large")
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, utils.ErrResponse{
			Ok: false,
			Status:"file size is too large,only allowed file size of max 2mb",
		})
		return
	} 

	// calling method to put data into the bucket - body as io.Reader -> bears a small container to load and send chunks of file data into bucket untill not fully sent the file
 	resolvedPostImageURL,bucketErr := s.S3BucketModel.UploadPostImageStream(
		c.Request.Context(),
		handledMutatedStream,
		clientIDStr,
		verdictedDataLength,
	)
	if bucketErr != nil {
		slog.Error("failed to upload post's image into the bucket❌","error",bucketErr)
		c.AbortWithStatusJSON(http.StatusBadRequest,utils.S3UploadErr{
			Ok: false,
			Error: "failed to upload post's image into the bucket",
		})
		return
	}

	slog.Info("Post image is uploaded successfully to the bucket✅","url",resolvedPostImageURL)

	// context for cleanup func
	cleanupCtx,timeout := context.WithTimeout(context.Background(),time.Second * 10 ) // max 10 sec alloted for req cancellation
	defer timeout()
	// todo - need a way to store this file url into the "posts" table -> need Image field
	err = s.PostDbModel.UpdatePostToStoreImageUrlByPostID(resolvedPostImageURL,postID)
	if err != nil {
		slog.Error("failed to store post image url in the db","error",err)
		slog.Info("attempting to delete uploaded image from the bucket...")
		// ! if err occured, we need to delete stored image from the bucket
		urlParts := strings.Split(resolvedPostImageURL,".amazonaws.com/") // splits the desired string from the given pattern
		if len(urlParts) < 2 {
        slog.Error("failed to parse S3 URL for cleanup", "url", resolvedPostImageURL)
        c.AbortWithStatusJSON(http.StatusInternalServerError, utils.ErrResponse{
            Ok:     false,
            Status: "database failure, and failed to parse cloud storage key for cleanup",
        })
        return
    	}

		// as we recieves two splits,posts...,http,taking first
		objKey := urlParts[1]
		slog.Info("extracted key...trying to delete object from the bucket","key",objKey)
		if _,err = s.S3BucketModel.BucketManager.S3Client.DeleteObject(cleanupCtx,&s3.DeleteObjectInput{
			Key: aws.String(objKey),
			Bucket: aws.String(s.S3BucketModel.BucketManager.S3BucketName),
		}); err != nil {
			slog.Error("failed to delete post image from the bucket","error",err)
			c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{
				Ok: false,
				Status: "failed to delete post image from the bucket",
			})
			return
		}
		
		slog.Info("successfully deleted uploaded image of the post from the bucket.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,utils.ErrResponse{
			Ok: false,
			Status: "failed to store image of the post in db; successfully deleted from the bucket for cleanup",
		})
		return
	}

	var defaultImageSrc string = "default.png"
	// if update is a success, cleaning up old stored image
	if oldImageURL != nil && *oldImageURL != defaultImageSrc && *oldImageURL != "" {
		go func(oldImageURL string) {
			// Run this in a background goroutine so the client doesn't have to wait for the old deletion!
			oldImgCleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			splitedURl := strings.Split(oldImageURL,".amazonaws.com/")
			// gives us two splitted parts
			if len(splitedURl) >=2 {
				_,err := s.S3BucketModel.BucketManager.S3Client.DeleteObject(oldImgCleanupCtx,&s3.DeleteObjectInput{
					Bucket: aws.String(s.S3BucketModel.BucketManager.S3BucketName),
					Key: aws.String(splitedURl[1]), // as first splitted part is what splitted after from the operation
				})
				if err != nil {
					slog.Error("failed to remove stored old post image from the bucket","error",err)
					return
				}

			} 
		}(*oldImageURL)
	}

	// if successfully uploaded file and stored the resolved url in the db by updating post url, send to the client
	c.JSON(http.StatusOK,utils.S3UploadSuccessResponse{
		Ok: true,
		Status: "successfully uploded post's image📤",
		ImageURL: resolvedPostImageURL,
	})


}


// flow
// 1. We need io reader stream pipeline which bears container to load and send data in chunks
// 2. body satisfies that but validation and handeling is utterly poorly managed
// 3. We need a multi-readers attached reader, where first reader for validation and rest prev body
// 4. We create a intialBuffer of size 512bytes which checks incoming data, and checks its content data type and validates it
// 5. If validated, we took initial buffer and attach to the body which has the remaining buffer
// 6. This creates a strealined multi reader where readers are consecutively attached and serves the data in the chunks as before