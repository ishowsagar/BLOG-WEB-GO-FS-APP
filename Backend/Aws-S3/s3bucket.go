package s3bucket

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// @types

// type that stores client of type s3.Client which -> holds all those operations related to s3
type BucketManager struct {
	// this stores s3, so whatever uses this will belongs to this parent type struct
	S3Client *s3.Client //! this the client for s3 
	SecretKey string
	SecretAccessID string
	S3BucketName string
	S3Region string
	// when created instance it -> will invoke method to use that client to check for operations or build if needed
}

// fix - we migrated to seperatly return instace so it could be used in operations which needs s3client

// func that returns instance of type BucketManager which -> stores methods that belongs to it
func NewBucketManager(key,accessID,bucket,region string) (*BucketManager,error) {
	// return type must satisfy s3.Client type -> initializes client for all s3 operations accross our 
	
	// credentials validation check first
	if region == "" || accessID == "" || bucket == "" || key == "" {
		// todo - later if successfully worked, migrate this to env for protection
		return nil,errors.New("missing required aws connection variables")
	}

	// 1 - Get static Credentials 
	staticCredentials := credentials.NewStaticCredentialsProvider(accessID,key,"") // provider expects accessID first,accessKey

	// 2 - Get connection config
	cfg,err := config.LoadDefaultConfig(context.TODO(),config.WithRegion(region),config.WithCredentialsProvider(staticCredentials))
		
	if err != nil {
	return nil,fmt.Errorf("failed to initialze config :%v",err)
	}

	// 3 - Load s3 Client from passed config which has all the credentials passed for connection
	s3Client := s3.NewFromConfig(cfg)

	// 4 - initializing instance which stores s3 client
	bucketManager := &BucketManager{
		S3Client: s3Client,
		SecretKey: key,
		SecretAccessID: accessID,
		S3BucketName: bucket,
		S3Region: region,
	}
	slog.Info("successfully created s3's client","bucketName",bucketManager.S3BucketName)
	return bucketManager,nil

}

// this func makes sure bucket always exists before operations, if not -> it builds the bucket
func(b *BucketManager) EnsureBucketExists(ctx context.Context,bucketName string,region string) (error) {

	// * 1 - check if bucket exists or not firstly
	// since client is one which -> holds all the operations so using it to check them
	_,err := b.S3Client.HeadBucket(ctx,&s3.HeadBucketInput{
		// headBucket checks if bucket exists or not
		Bucket: aws.String(bucketName),//* use aws formatter for strings
	})

	// if headBucket returns no err -> the bucket is reachable and ready
	if err == nil {
		fmt.Printf("Bucket '%s' already exists and you own it. Ready to hop!\n",bucketName)
		return nil
	}

	//* 2 - if headBucket fails, try to create the bucket unless we can prove the request itself is invalid
	//! many valid S3 setups return AccessDenied/Forbidden here when the caller lacks HeadBucket permission
	var apiErr smithy.APIError //sdk err handeling
	if errors.As(err,&apiErr) {
		// if this err type is matching this target &err type

		switch apiErr.ErrorCode() {
		case "NotFound", "AccessDenied", "Forbidden":
			fmt.Printf("Bucket '%s' is not confirmed as available. Attempting to create it...",bucketName)
			input := &s3.CreateBucketInput{
				Bucket: aws.String(bucketName),
			}

			if region != "us-east-1" {
				input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
					LocationConstraint: types.BucketLocationConstraint(region),
				}
			}

			_, createErr := b.S3Client.CreateBucket(ctx,input)
			if createErr != nil {
				var createAPIErr smithy.APIError
				if errors.As(createErr, &createAPIErr) {
					if createAPIErr.ErrorCode() == "BucketAlreadyOwnedByYou" || createAPIErr.ErrorCode() == "BucketAlreadyExists" {
						fmt.Printf("Bucket '%s' already exists and is usable.\n", bucketName)
						return nil
					}
				}
				return fmt.Errorf("failed to create bucket '%s': %w", bucketName, createErr)
			}

			fmt.Printf("Successfully created bucket '%s' in region '%s'!\n",bucketName,region)
			return nil
		default:
			return fmt.Errorf("unexpected error checking bucket '%s': %w", bucketName, err)
		}
	}

	return fmt.Errorf("unexpected error occurred during checking if buckets already exists or not: %w",err)
}