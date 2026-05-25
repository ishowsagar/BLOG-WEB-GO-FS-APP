package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// @types

// type struct that stores client of type redis.Client -> client for read/writes
type RedisCacheClient struct {
	// this client is source of truth for all operations
	Client *redis.Client //&  this redis client is the one which -> holds all the operations for redis caching
}

// type struct that holds type of user form data is being stored
type UserCachePayload struct {
	Email string
	Password string
}

//  func that returns the instance of type -> RedisCacheClient -> has client for read/writes source of truth for all operations
func NewRedisCacheClient(redisClientPassword string,db int,redisHostAddress string) *RedisCacheClient {
	return &RedisCacheClient{
		Client: redis.NewClient(&redis.Options{
			Addr: redisHostAddress, // hardcoding addr for now, fix later with env protected var
			Password: redisClientPassword,
			DB: db,
			
		}),
	}
}



// fnc that belongs to the RedisCacheClient which -> store key-val pairs data
func(r *RedisCacheClient) SetCachedToken(userID uint,tokenExpiry time.Time,tokenHash string) (error) {

	// need to generate key-val data variables to store into cache db
	key := fmt.Sprintf("auth:token:user:%d",userID) // key be8ing userID
	
	// mapped data of token - same like struct 
	payload := map[string]interface{} {
		"hash" : tokenHash,
		"expiry" : tokenExpiry,
	}

	// marshal payload into -> [] of bytes data
	body,err := json.Marshal(payload)
	if err != nil {
		return err
	}

	intialContext := context.Background()
	ttl := time.Until(tokenExpiry) //* how much time left it to go expire

	// sets into db the key-val pair that we are setting here
	err = r.Client.Set(intialContext,key,body,ttl).Err()
	if err!= nil {
		return err
	}
	return nil

}

// fetch stores value of stored token from -> key
func(r *RedisCacheClient) GetCachedToken(userID uint) (tokenHash string,tokenExpiry time.Time,cachingErr error) {

	//  setting return type more verbose for more intuition and clarity 👆👆

	key := fmt.Sprintf("auth:token:user:%d",userID)


	
	// meths return err if *lookup does not exists
	ctx := context.Background()
	
	// client holds all the operations -> getting result <- of type []byte
	res,notFoundErr := r.Client.Get(ctx,key).Result()
	if notFoundErr != nil {
		return "",time.Time{},notFoundErr
	}

	// if key-pair found -> return value stored in it. store into same type of data struct what it was used to store into when setting in the cache db
	var payload struct { // same map like struct field and json tags for population
		Hash string `json:"hash"`
		Expiry time.Time `json:"expiry"`
	}

	// unmarshal payload incoming on this request from -> []byte data 
	// this method unmarshals the passed [slice] of byte data into a struct,and unloads~populates into specified var that holds same type of data 
	err := json.Unmarshal([]byte(res),&payload) // unmarshal into this var since val in setCacheToken Method was set in []byte type
	if err != nil {
		return "",time.Time{},err
	}

	// if successfull now -> payload holds data unmarshaled from redis result

	// return type of each key's val and ov=bvs err
	return payload.Hash,payload.Expiry,nil
}



//  func that belongs to the type -> RedisCacheClient -> deletes key-val pair stored in cache db
func(r *RedisCacheClient) DeleteCachedToken(userID uint) error {

	key := fmt.Sprintf("auth:token:user:%d",userID)
	if err := r.Client.Del(context.Background(),key).Err() ; err != nil {
		return err
	}

	//  if no err hit -> successfully deleted from cached db 
	return nil
}


// set the user cached in redis db - pass user of type models.user with hashed pass 
func(r *RedisCacheClient) SetCachedUser(ctx context.Context,userEmail string,userPassword string) (error) {

	// redis data is stored in key-val pair and must being in form of type [slice] of byte

	// set key which would store some val - must have same pattern key (%d - int placeholderVal,%s for string val)
	key := fmt.Sprintf("form:login:user:email:%s",userEmail) //* key being the email

	// val stored in variable struct - all data which need to store for fetching fast
	payload := UserCachePayload{
		Email: userEmail,
		Password: userPassword,
	}

	// marshal into [] of byte type of data to which will be stored in db
	marshaledPayload,err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ttl := time.Hour * 24 //* after 24 this will be expired automatically

	// set key-pair in cache with .set method from *redis'client
	err = r.Client.Set(ctx,key,marshaledPayload,ttl).Err()
	if err != nil {
		return err
	}

	return nil
} 


// get user stored in cached db
func(r *RedisCacheClient) GetCachedUser(ctx context.Context,email string) (cachedUserEmail string,cachedUserPassword string,cachingErr error) {

	// provide key to client's get method to get data stored in that key
	key := fmt.Sprintf("form:login:user:email:%s",email)

	// get result -> gives in form of type []of bytes
	marshaledRes,err := r.Client.Get(ctx,key).Result()
	if err != nil {
		return "","",err
	}

	// unmarshal into struct data - (unmarshing[slice]bytes(data)into struct,&populate into? this var)
	var cachedUserResponse UserCachePayload
	
	err = json.Unmarshal([]byte(marshaledRes),&cachedUserResponse)
	if err != nil {
		return "","",err
	}

	// note - yellow err are formatting errs

	// if successfully unmarshed from []bytes into struct data
	// now var holds all those stored vals on that key
	
	// return appropriate fields from it
	return cachedUserResponse.Email,cachedUserResponse.Password,nil
}
