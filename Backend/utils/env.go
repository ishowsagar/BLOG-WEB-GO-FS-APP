package utils

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type ENVConfig struct {
	DbHost string
	DbUser string
	DbPass string
	DbPort int
	DbName string
	ServerPort string
	JwtSecret string
	RedisDB int
	RedisDBPassword string
	RedisHost string 
	RabbitMQURL string
	S3AccessKeyID string
	S3SecretKey string
	S3BucketName string
	S3RegionName string
}

var once sync.Once 
var config *ENVConfig // var that stores type of ENVConfig
// returns instance of type Config -> which stores all env variables
func LoadConfig() (*ENVConfig,error) {

	// load env
	err := godotenv.Load()
	if err != nil {
		return nil,err
	}

	//** Note - since our app is pulling var from local env file which is inaccessible to the docker containers, so we had to define matching variables in yml so it pulls from yml env **// 
	// access env protected variables
	dbUSER := os.Getenv("DB_USER")
	dbPASS := os.Getenv("DB_PASSWORD")
	dbHOST := os.Getenv("DB_HOST")
	dbPortStr := os.Getenv("DB_PORT")         // * need to parse for cred
	serverPortStr := os.Getenv("SERVER_PORT") // * need to parse for cred
	JwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	redisDbPass := os.Getenv("REDIS_DB_PASSWORD")
	redisDbStr := os.Getenv("REDIS_DB")
	redisDbHost := os.Getenv("REDIS_HOST_ADDR")
	rabbitmqURL := os.Getenv("RABBITMQ_URL")

	// * since we stored aws s3 important keys in env, container would look for them in its space <- must define there too
	s3AccessKeyID := os.Getenv("S3AccessKeyID")
	s3SecretKey := os.Getenv("S3SecretKey")
	s3BucketName := os.Getenv("S3BucketName")
	s3RegionName := os.Getenv("S3Region")

	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		slog.Error("failed to get port", "error", err)
		return nil,err
	}

	redisDb, err := strconv.Atoi(redisDbStr)
	if err != nil {
		slog.Error("failed to get redis db", "error", err)
		return nil,err
	}
	// port,err := strconv.Atoi(serverPortStr)
	// if err != nil {
	// 	slog.Error("failed to get port","error",err)
	// 	return
	// }

	dbName := os.Getenv("DB_DBASE_NAME")


	// env check
	fmt.Println("--- DOCKER ENV CHECK ---")
	fmt.Printf("Access ID:  '%s'\n", os.Getenv("S3AccessKeyID"))
	fmt.Printf("Secret Key: '%s' (Length: %d)\n", os.Getenv("S3SecretKey"), len(os.Getenv("AWS_SECRET_ACCESS_KEY")))
	fmt.Printf("Bucket:     '%s'\n", os.Getenv("S3BucketName"))
	fmt.Printf("Region:     '%s'\n", os.Getenv("S3Region"))
	fmt.Println("------------------------")

	// returning instance with env accessed vars
	return &ENVConfig{
		DbHost: dbHOST,
		DbUser: dbUSER,
		DbPass: dbPASS,
		DbPort: dbPort,
		DbName: dbName,
		ServerPort: serverPortStr,
		JwtSecret:JwtSecretKey,
		RedisDB: redisDb,
		RedisDBPassword: redisDbPass,
		RedisHost: redisDbHost,
		RabbitMQURL: rabbitmqURL,
		S3AccessKeyID: s3AccessKeyID,
		S3SecretKey: s3SecretKey,
		S3BucketName: s3BucketName,
		S3RegionName: s3RegionName,
	},nil
}


// load config type's instance all in once for use
func GetConfig() *ENVConfig {
	once.Do(func() {
		loadedConfig,err := LoadConfig()
		if err != nil {
			panic(err)
		}
		config = loadedConfig //* setting var to assign this struct instance values 
	})
	return  config
}