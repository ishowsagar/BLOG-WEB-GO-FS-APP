package initializers

import (
	"context"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// @ this func is basically for testing application services before firing up whole app
func VerifyInfraStack(gormDB *gorm.DB,redisClient *redis.Client,rabbitMQBrokerServiceURl string)  {

	log.Println("========== 🚀PRE-TESTING APP SYSTEM INFRASTRUCTURE STACK🛫==========")

	// context
	ctx,timeout := context.WithTimeout(context.Background(),time.Second * 6) // max 6 sec timeout otherwise fails
	defer timeout()
	
	// 1 - pinging underlying sql.DB from gorm connection
	sqlUnderlyingGormDB,err := gormDB.DB()
	if err != nil {
		log.Fatalf("❌ DATABASE ERROR: Could not retrieve SQL instance: %v", err) 
		// since fatalF return immeditaly so no need of return even it won't execute furthur return will be unreachable
	}
	if err = sqlUnderlyingGormDB.PingContext(ctx) ; err != nil {
		log.Fatalf("❌ DATABASE ERROR: Connection check failed! App crashing at boot time. Error: %v", err)
	}

	log.Println("DATABASE: Connection is working fine and successfully pinged database✅.")


	// 2 - pinging Redis cache cluster 
	if err = redisClient.Ping(ctx).Err() ; err != nil {
		log.Fatalf("❌ REDDIS DB CONNECTION ERROR : cache clust unreachable! App crashing at the boot times. Error :%v",err)
	}
	log.Println("✅ REDIS CACHE: Operational and ping-pong verified.")


	// 3 - Verifying rabbitMqBroker connection
	brokerConn,err := amqp091.Dial(rabbitMQBrokerServiceURl)
	if err != nil {
		log.Fatalf("❌ RABBIT MQ BROKER CONNECTION ERROR : Message broker broker offline! App crashing at boot time. Error: %v", err)
	}
	defer brokerConn.Close()

	log.Println("✅ RABBITMQ: Event messaging pipeline successfully connected.")

	log.Println("============ 🎉SYSTEM INFRASTURE STACK WORKING FINE⚡: LAUNCHING APP🚀 ============")
}