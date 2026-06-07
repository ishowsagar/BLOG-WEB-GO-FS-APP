package s3bucket

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// creates invalidation request & fires it
func InvokeCdnInvalidation(distID string) (error) {


	// utils
	ctx := context.Background()

	//* 1- config - loading default used automatically
	awsConfig,err :=config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Error("failed to load aws config","error",err)
		return err
	}

	//* 2- creating cdn client from sdk <- need to feed in aws config which =>handles all the cdn operations
	cdnClient := cloudfront.NewFromConfig(awsConfig)

	// unique id for invalidation record
	uniqueInvalidationReqID := fmt.Sprintf("go-sdk-aws-cdn-%d",time.Now().Unix())


	// * 3- structuring invalidation request payload
	invalidationReqPayload := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(distID),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: &uniqueInvalidationReqID,
			Paths: &types.Paths{
				Quantity: aws.Int32(1), // batch request in num
				Items: []string{
					"/*", // for all routes
				},
			},
		},
	}
	
	// * 4- triggering invalidation request from *client
	output,err := cdnClient.CreateInvalidation(ctx,invalidationReqPayload)
	if err != nil {
		slog.Error("failed to fire invalidation request","error",err)
		return err
	}

	slog.Info("successfully ran invalidation🎉","invalidationID",output.Invalidation.Id)
	return nil
}