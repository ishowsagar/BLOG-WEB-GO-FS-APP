package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	cache "github.com/ishowsagar/go-blog-web-application/Cache"
	"github.com/ishowsagar/go-blog-web-application/controller"
	"github.com/ishowsagar/go-blog-web-application/database"
	"github.com/ishowsagar/go-blog-web-application/migrations"
	routes "github.com/ishowsagar/go-blog-web-application/router"
	"github.com/ishowsagar/go-blog-web-application/services"
	_ "github.com/ishowsagar/go-blog-web-application/store"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// types


func main() {

	// slog logger for entire application
	logger := slog.New(slog.NewTextHandler(os.Stdout,&slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
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
	err = migrations.AutoMigrate(baseDbModel.DB)
	if err != nil {
		slog.Warn("failed to migrate models","error",err)
		return	
	}

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
	followController := controller.NewFollowController(followDbModel)

	// master controller -> stores all corresponding controllers
	masterController := controller.NewMasterController(userController,postController,commentController,likeController,followController)

	router := gin.Default()
	routes.ServeRoutes(router,masterController,config) // serving controller to router to route methods on them routes
	router.Run(fmt.Sprintf(":%s",config.ServerPort))
}