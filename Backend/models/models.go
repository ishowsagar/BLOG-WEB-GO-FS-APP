package models

import (
	"time"
)

// @ types of all types of models used across the App

//  store all the models which needed to be migrated, later could be added more
type MigrationsStore struct {
	UserType User
	PostType Post
	CommentType Comment
}

//  user type struct
type User struct {
	ID uint `json:"id" gorm:"primaryKey"`
	Name string `json:"name" binding:"required"` //! json for type ingres, binding for gin req binder
	Email string `json:"email" binding:"required"` //! maybe would need form tags for binding
	Password string `json:"-" binding:"required"`
	// # adding more fields for more data in user struct for each profile
	Username string `json:"username" gorm:"default:'insta_user12345'"`
	Nickname string `json:"nickname" gorm:"default:'User'"`
	Bio string      `json:"bio" gorm:"default:'New to instagram, follow me'"` 
	CreatedAt time.Time `json:"created_at" time_format="2006-01-02"`
	// this is the power of migration, u can always add changes
	// & setting default values on these fields to be 0 -> reduce stress on COALESCE
	FollowersCount uint `json:"followers_count" gorm:"default:0"`
	FollowingCount uint `json:"following_count" gorm:"default:0"`
}

// bug - if no constraint are set - if user or post related somehwere data deleted - that must be deleted,eg - user dleteed then its likes n comments n associated posts
// fix - add gor:"gorm:constraint:OnUpdate:CASCADE,OnDelete:CASCADE"

// post type struct --> made by which UserID
type Post struct {
	ID uint `json:"id" gorm:"primaryKey"`
	// foregein key refs for corelation
	User User `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`// gorm ref
	UserID uint `json:"user_id" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	Title string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	LikeCount uint `json:"like_count" gorm:"-"`
	CreatedAt time.Time `json:"created_at" time_format="2006-01-02"`
	UpdatedAt time.Time `json:"updated_at" time_format="2006-01-02"`	
}


// comment type struct -> determined by which PostID,Made by which UserID
type Comment struct {
	ID uint `json:"id" gorm:"primaryKey"`
	 // for gorm refrence purpose
	User User `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	Post Post `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:PostID;references:ID"` //bug - if we delete post, we had to delete data associated with that ids
	UserID uint `json:"user_id" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	PostID	uint `json:"post_id"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:PostID;references:ID"`
	Content string `json:"content" binding:"required"`
	CreatedAt time.Time `json:"created_at" time_format="2006-01-02"`
	UpdatedAt time.Time `json:"updated_at" time_format="2006-01-02"`
}

//  type token struct for token
type Token struct {
	ID uint `json:"id" gorm:"primaryKey"`
	User User `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	UserID uint `json:"user_id" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	Expiry time.Time  `json:"expiry" time_format="2006-01-02"`
	Hash string `json:"hash"`
}


type Like struct {
	ID uint `json:"id" gorm:"primaryKey"`
	//* need which "user" is posting comment on which "post"
	User User `json:"-"`
	Post Post `json:"-"`
	//  references - gorm:"fk":(prefixMustBeTimeName)Field:refs:ID{fromThatTableSpecified}
	UserID uint `json:"user_id" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID"` 
	PostID	uint `json:"post_id"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:PostID;references:ID"`
	LikeCount uint `json:"like_count"`
    LikedAt time.Time `json:"liked_at" time_format="2006-01-02"` //bug - hit bug for not finding Timestamp field, as we were using : to assign field type but we had to use "=" espacially for this field
// bug another one ** => mistakenly created wrong field name with case err, BigCase treat as _, so TimeStamp treated as -> 't'ime_'s'tamp
// fix- added fresh right col, but need to drop the wrong col with function created in migrations
// 💪💪this is the power of migrations
}


//  type struct for follow data struct table
type Follow struct {
	ID uint `json:"id" gorm:"primaryKey"`
	// references - to set fk on them to
	Follower User `json:"-" gorm:"foreignKey:FollowerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Followee User `json:"-" gorm:"foreignKey:FolloweeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	// adding constraint to not to let anyone follow twice
	FollowerID uint `json:"follower_id"  gorm:"not null;uniqueIndex:idx_follower_followee"`
	FolloweeID uint `json:"followee_id"  gorm:"not null;uniqueIndex:idx_follower_followee"`
	FollowedAt time.Time `json:"followed_at" time_format="2006-01-02"`
}



// referencing models
// User User `json:"-"`
// type Profile struct {
// 	ID uint `json:"id" gorm:"primaryKey"`
// 	//& references for gorm
// 	User User
// 	Post Post
// 	// * relationships - which user is being references about
// 	UserID uint `json:"user_id" gorm:"constraint:onUpdate:CASCADE,onDelete:CASCADE;foreignKey:UserID:references:ID"`
// 	Bio string `json:"bio"`
// 	// todo - need to add fields for followers count and following count for an user related to user table
// 	FollowerCount uint `json:"follower_count"`
// 	FollowingCount uint `json:"following_count"`
// 	Pronouns string `json:"pronouns"`
// 	// todo - need postCount field which always add post_count by +1 whenever there is a new post in posts table
// 	Posts uint `json:"posts" gorm:"constraint:onUpdate:CASCADE,onDelete:CASCADE;"`
// }

