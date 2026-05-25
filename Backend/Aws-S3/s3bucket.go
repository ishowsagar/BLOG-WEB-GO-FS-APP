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
	S3Client *s3.Client
	SecretKey string
	SecretAccessID string
	S3BucketName string
	// when created instance it -> will invoke method to use that client to check for operations or build if needed
}

// fix - we migrated to seperatly return instace so it could be used in operations which needs s3client

// func that returns instance of type BucketManager which -> stores methods that belongs to it
func NewBucketManager(key,accessID,bucket string) (*BucketManager,error) {
	// return type must satisfy s3.Client type -> initializes client for all s3 operations accross our 
	
	// credentials validation check first
	if Region == "" || accessID == "" || bucket == "" || key == "" {
		// todo - later if successfully worked, migrate this to env for protection
		return nil,errors.New("missing required aws connection variables")
	}

	// 1 - Get static Credentials 
	staticCredentials := credentials.NewStaticCredentialsProvider(accessID,key,"") // provider expects accessID first,accessKey

	// 2 - Get connection config
	cfg,err := config.LoadDefaultConfig(context.TODO(),config.WithRegion(Region),config.WithCredentialsProvider(staticCredentials))
		
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
	}
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

	// if headBucket which checks if that bucket exists or not returns no err -> means it exists already
	if err == nil {
		fmt.Printf("Bucket '%s' already exists and you owns it. Ready to hop!\n",bucketName)
		return nil
	}

	//* 2 - if it doesn't exists -> build new
	//! but if there was an err, failed to check/retrieve info that specified bucket exists or not
	var apiErr smithy.APIError //sdk err handeling
	if errors.As(err,&apiErr) {
		// if this err type is matching this target &err type

		if apiErr.ErrorCode() == "NotFound" {
			fmt.Printf("Bucket '%s' not found. Attempting to create new bucket...",bucketName)
			//* 3 - Building bucket

			// creating bucket with s3.CreateBucketInput method
			input := &s3.CreateBucketInput{
				Bucket: aws.String(bucketName),
			}
			slog.Info("successfully created bucket input")
			
			_, err := b.S3Client.CreateBucket(ctx,input,)
			if err != nil {
				return err
			}
			slog.Info("successfully created bucket from bucket-input")

			if region != "us-east-1" {
				// if region is passed diffrent other than "us-east-1"
				input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
					LocationConstraint: types.BucketLocationConstraint(region),
				}
			}

			fmt.Printf("Successfully created bucket '%s' in region '%s'!\n",bucketName,region)
			return nil
		}

		// if someone else owns it
		if apiErr.ErrorCode() == "AccessDenied" {
			return fmt.Errorf("the bucket name '%s' is already taken by someone else globally",bucketName )
		}
	}

	return fmt.Errorf("unexpected errror occurred during checking if buckets already exists or not :%w",err)
}