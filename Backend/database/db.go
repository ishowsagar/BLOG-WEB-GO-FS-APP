package database

import (


	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @Types

// BaseDBModel type that stores DB of type *sql.DB
type BaseDBModel struct {
	DB *gorm.DB
} 


//  function that returns instance of baseDbModel that stores db conn of type *sql.DB
func ConnectToDatabase(dbConnectionString string) (*BaseDBModel,error) {


	// send credentials for postgres db connection
	db,err := gorm.Open(postgres.Open(dbConnectionString),&gorm.Config{})
	// retrieve db
	if err != nil {
		// when returning err --> log where it is actually called
		return nil,err
	}
	
	// configuring db
	// db.SetMaxIdleConns(5)
	// db.SetMaxOpenConns(10)
	// db.SetConnMaxIdleTime(5 * time.Minute)
	
	// early ping testing
	// err = PingTestingDb(db)
	// if err != nil {
	// 	defer db.Close() // after pinging
	// 	slog.Error("failed to ping db","error",err)
	// 	return nil,err
	// }

	// returning Instance of BaseDBModel to store retrieved db conn
	return &BaseDBModel{
		DB: db,
	},nil
}

