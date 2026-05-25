package services

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3bucket "github.com/ishowsagar/go-blog-web-application/Aws-S3"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// @ types declaration

//type that stores bucketManager which -> underlying stores s3Client which holds all bucket related operations
type S3BucketModel struct {
	BucketManager *s3bucket.BucketManager
	// now this also stores db for queries
	DB *sql.DB
}


// func that returns the instance of type S3BucketModel which -> stores services method that belongs to it
func NewS3BucketModel(bucketManager *s3bucket.BucketManager,db *sql.DB) *S3BucketModel {
	return &S3BucketModel{
		BucketManager:bucketManager,
		DB: db,
	}
}


// mehtod that belongs to the type S3BucketModel which -> uploads image to the s3 bucket
func(s *S3BucketModel) UploadImageStream(ctx context.Context,subFolder string,uniqueID string,originalFileName string,fileBody io.Reader,fileSize int64) (string,error) {

	// 1 - check what type of image it was
	ext := filepath.Ext(originalFileName) // returns file extension it had to set on object key

	// 2 - generating s3 object key string <- client not guess to override things manually
	objectKey := fmt.Sprintf("%s/user-%s-%d%s",subFolder,uniqueID,time.Now().Unix(),ext) // Example: profiles/user-12345-1716300000.jpg
	
	// 3 - pipeline this input to upload to s3
	_,err := s.BucketManager.S3Client.PutObject(ctx,&s3.PutObjectInput{
		Bucket: aws.String(s.BucketManager.S3BucketName), // putting where
		Key: aws.String(objectKey), // object key
		Body: fileBody, // what is putting eg.body -> of type io.reader cuz files are recieved and sent in chunks.parts
		ContentLength: aws.Int64(fileSize),
	})

	// if failed to upload to s3
	if err != nil {
		// tip - instead of method err, return explicit and dynamic errors with fmt.Errorf()
		return "",fmt.Errorf("failed to upload image stream to s3 :%w",err)
	}

	// 5 - if successfully uploaded✅✅, return url where it is stored by -> setting it the url as this is standard url just need to put those val in url is same created by s3 
	fileURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s",s.BucketManager.S3BucketName,objectKey)
	return fileURL,nil
}


func(s *S3BucketModel) InsertImage(userID uint,uploadedPictureCloudURL string) (error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	// its like if client tried to upload again and this fired again -> but insert might failed as it cannot insert as there is already one
	// so we are saying -> if there's conflict on (fieldBeing=>user_id) do this -> Do Update Set thisFieldVal=be this, execluded means to set newwly coming val and as excluded(backup) upsert url
	query := `
		Insert into
			profile_picture_storages(user_id,profile_picture_url)
		values
		 	($1,$2)
		On
		 Conflict
			(user_id)
		Do 
			Update
			Set
				profile_picture_url=Excluded.profile_picture_url
	`

	resRow,err := s.DB.ExecContext(ctx,query,userID,uploadedPictureCloudURL)
	if err != nil {
		return err	
	}

	rowsAffected,err := resRow.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	// if successfully inserted data into the db✅✅
	return nil
}