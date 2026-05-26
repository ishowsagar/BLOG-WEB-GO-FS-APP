package s3bucket

import (
	"context"
	"fmt"
	// "os"
)

// @Aws Lifecycle
// 1 - Context heavily used for requests
// 2 - awsConfig for accessing client credentials and talk on his behalf
// 3 - s3.Client -> This what created from configs -> handles all those upload get operations

// @ variables for connecting to the aws s3

	// &Note -> aws console is very glitchy and prone to err -> use aws cli <- smooth,easy and best
	// create user -> aws --create-user --user-name X
	// attach policies on the created user - aws iam attach-user-policy --user-name X --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess
	// generate keys - aws iam create-access-key --user-name X
	//! get credentials from aws console so it verifies and connect in connect function
	// Region string = "us-east-1"
	// s3Endpoint string = ""
	// bug- exposed keys in the codespace
	// fix - fetch through env only, make sure naming is same as that would be checked by containe in its own environement space

	// fixed - injected into bucketManager type
	// bug - failed to create bucket 'aws-s3-insta-bucket-' seems invalid cause of -?
	// fixed - yes that '-' causes error as aws strictly does not let u create those names with uncivilized names





// & Commands for listing resources we have just created
// list s3 - aws s3 ls
//  list specfic iam user - aws iam get-user X 

// ! retrieved info from creating iam user,policies and secretAccessKeys

//* connect to s3 -> creates client for connecting to s3 for -> doing all operations accross the app
func(b *BucketManager) ConnectToS3Bucket() error {
// os.Getenv()
	//  invoke method which belongs to type BuckerManager -> uses underlying s3Client which -> checks for bucket existence if not -> build it
	err := b.EnsureBucketExists(context.Background(),b.S3BucketName,b.S3Region)
	if err != nil {
		// bug - was using log.Fatalf() -> which logs but immeditiatly exits(1) out so next return never executes
		// fix - use fmt.Errorf() -> which returns the err directly and also could intake placeholders for dynamic err handeling operations
		return fmt.Errorf("error occurred managing bucket check/build :%v",err)
	}

	fmt.Printf("s3 setup is complete and fully integrated into Go application🚀")
	return nil
}



