package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	s3bucket "github.com/ishowsagar/go-blog-web-application/Aws-S3"
	cache "github.com/ishowsagar/go-blog-web-application/Cache"
	"github.com/ishowsagar/go-blog-web-application/controller"
	"github.com/ishowsagar/go-blog-web-application/database"
	"github.com/ishowsagar/go-blog-web-application/events"
	"github.com/ishowsagar/go-blog-web-application/initializers"
	routes "github.com/ishowsagar/go-blog-web-application/router"
	"github.com/ishowsagar/go-blog-web-application/services"
	_ "github.com/ishowsagar/go-blog-web-application/store"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// types


func main() {

	// slog logger for entire application
	logger := slog.New(slog.NewJSONHandler(os.Stdout,nil))
	slog.SetDefault(logger) //* default logger set for whole app

	// custom logger
	f,err := os.Create("server_logs")
	if err != nil {
		slog.Error("failed to create log file for server logs","error",err)
		return
	}
	gin.ForceConsoleColor()
	gin.DefaultWriter = io.MultiWriter(f,os.Stdout)


	// // load env
	// err = godotenv.Load()
	// if err != nil {
	// 	return
	// }

	// // access env protected variables
	// dbUSER := os.Getenv("DB_USER")
	// dbPASS := os.Getenv("DB_PASSWORD")
	// dbHOST := os.Getenv("DB_HOST")
	// dbPortStr := os.Getenv("DB_PORT") // * need to parse for cred
	// serverPortStr := os.Getenv("SERVER_PORT") // * need to parse for cred
	// dbPort,err := strconv.Atoi(dbPortStr)
	// if err != nil {
	// 	slog.Error("failed to get port","error",err)
	// 	return
	// }
	// // port,err := strconv.Atoi(serverPortStr)
	// // if err != nil {
	// // 	slog.Error("failed to get port","error",err)
	// // 	return
	// // }

	// dbName := os.Getenv("DB_DBASE_NAME")

	// test - for faster loads - i've put everything in config type struct, load once serve to all who needs
	config,err := utils.LoadConfig()
	if err != nil {
		slog.Error("failed to load envConfig type which had env protected variables","error",err)
		return 
	}

	// log DB host/name for debugging which DB instance we connect to (no password logged)
	slog.Info("DB config","host",config.DbHost,"dbname",config.DbName)

	connectionStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable",config.DbHost,config.DbUser,config.DbPass,config.DbName,config.DbPort)
	baseDbModel,err := database.ConnectToDatabase(connectionStr)
	if err != nil {
		slog.Error("failed to open db connection","error",err)
		return
	}

	// todo - serve this to API
	sqlDB,err := baseDbModel.DB.DB() //* underlying sqlDb 
	if err != nil {
		slog.Error("failed to load underlying sql db","error",err)
		os.Exit(1) // will stop cause our whole app needed this db
	}

	// migrations
	// err = migrations.AutoMigrate(baseDbModel.DB)
	// if err != nil {
	// 	slog.Warn("failed to migrate models","error",err)
	// 	return	
	// }

	//  for constraints
	// err = migrations.EnsureCascadeConstraints(baseDbModel.DB)
	// if err != nil {
	// 	slog.Warn("failed to replace constraints","error",err)
	// 	return	
	// }
	//  demigrations - only invoke when hit bug and need,otherwise don't - it will throw err since that col already been removed from migrations
	// err = migrations.Demigrate(baseDbModel.DB)
	// if err != nil {
	// 	slog.Warn("failed to Demigrate model","error",err)
	// 	return 
	// }
	defer sqlDB.Close() // deferring it to close db conn after done invocation
	slog.Info("successfully connected to the database✅")


	//* initializing needful instances to invoke route handlers
	
	// models
	userDbModel := services.NewUserDbModel(sqlDB)
	tokenDbModel := services.NewTokenDbModel(sqlDB)
	postDbModel := services.NewPostDbModel(sqlDB)
	commentDbModel := services.NewCommentDbModel(sqlDB)
	likeDbModel := services.NewLikeDbModel(sqlDB)
	followDbModel := services.NewFollowDbModel(sqlDB)

	// todo - add interface for all services, need to change handlers for them -> cause handlers stores concrete type for calling services
	// stores
	// masterStore := store.NewMasterStore(likeDbModel,userDbModel,postDbModel,commentDbModel) 
	// todo -> addstores - types that store instances 
	
	//& cached db redis client 
	// pulling var from config loaded
	redisClient := cache.NewRedisCacheClient(config.RedisDBPassword,config.RedisDB,config.RedisHost)

	// go chan services
	pushNotificationService := services.NewPNService() //* starts worker which continusoly reads noti which is redirected to the pns's notif channel
	// todo - just need to redirect stuff to the chan only+
	// func based mini-controllers
	userController := controller.NewUserController(userDbModel,tokenDbModel,redisClient)
	commentController := controller.NewCommentController(commentDbModel,pushNotificationService)
	postController := controller.NewPostController(postDbModel,pushNotificationService)
	likeController := controller.NewLikeController(likeDbModel,pushNotificationService)
	followController := controller.NewFollowController(followDbModel,pushNotificationService)

	// **  Stack verification  health check firstly before running main app  **//

	initializers.VerifyInfraStack(baseDbModel.DB,redisClient.Client,"amqp://guest:guest@instagram_rabbitmq_container:5672/")

	// ** ....END... **//


	// ! bug -> fix s3 related err tmrw it maybbe due to sending wrong keys and ids
	
	// start hub service to handle broadcasts and client management
	hub := services.IntializeNewHubInstance()
	go hub.RunService()
	
	// start pubsub
	broker := events.NewPubSubBroker(hub) // intializes an instance with hub
	// start connection
	// testing - since composed services depends on each other so mapping to service from where it will pull via internal networking
	err = broker.Connect("amqp://guest:guest@instagram_rabbitmq_container:5672/") //& initializes rabbit connection and exchange and queues and stores in this instance on which <- this being called on
	if err != nil {
		slog.Info("failed to open rabbit connection to declare exchanges and all","error",err)
		return
	}


	// ! KEEPS RUNNING IN BACKGROUND -> checks for incoming deliveries related to "user.*" -> might need diff exchange for seperate but for now only this
	err = broker.StartConsumingDeliveries() //& checks for incoming stamped deliveries events in "notification" exchange router
	if err != nil {
		slog.Error("failed to start subs service which consumes incoming data to subscribers","error",err.Error())
		return
	}
	defer broker.Close() // closing connection once it is done
	
	// inject broker into hub so clients can publish delivery messages
	hub.SetBroker(broker) //sets broker in the hub instance

	// ws controller
	wsController := controller.NewWsController(baseDbModel.DB,hub,config.JwtSecret,broker)


	// connect notification service to hub for broadcasting
	pushNotificationService.SetHub(hub)

	
	//& AWS-S3-BUCKET SETUP
	
	// bucketManeger type's instance which -> stores s3Client which holds all the bucket operations
	bucketManager,err := s3bucket.NewBucketManager(config.S3SecretKey,config.S3AccessKeyID,config.S3BucketName,config.S3RegionName) //& returns s3client in instance
	if err !=nil {
		slog.Error("failed to initialze bucketManager","error",err)
		return
	}
	slog.Info("bucket manager is successfully created","accessKeyID",bucketManager.SecretAccessID,"secretKey",bucketManager.SecretKey)
	// fires up ensureBucketExists method to check or build bucketD
	err = bucketManager.ConnectToS3Bucket() 
	if err != nil {
		slog.Error("failed to setup s3 bucket in our go application.","error",err)
		return
	}

	s3bucketModel := services.NewS3BucketModel(bucketManager,sqlDB)
	s3Controller := controller.NewS3Controller(s3bucketModel)
	
	// master controller -> stores all corresponding controllers
	masterController := controller.NewMasterController(userController,postController,commentController,likeController,followController,s3Controller)



	router := gin.Default()
	routes.ServeRoutes(router,masterController,config,wsController) // serving controller to router to route methods on them routes
	router.Run(fmt.Sprintf(":%s",config.ServerPort))
}