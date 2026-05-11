package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/ishowsagar/go-blog-web-application/models"
	"github.com/ishowsagar/go-blog-web-application/utils"
)

// @types

//  type UserDBModel that stores db of type *sql.DB
type UserDBModel struct {
	db *sql.DB
}

// @ Interface -> Stores all the methods which belongs to the *UserDbModel type
type UserStore interface {
	CreateUser(user *models.User) (*models.User,error)
	GetUserByEmail(email string) (*models.User,error)
} 

// @ methods belongs to the type -> UserDBModel which stores db of type *sql.DB
func(u *UserDBModel) CreateUser(user *models.User) (*models.User,error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	
	query := `
		INSERT INTO users(name,email,password,username,nickname,bio,created_at,followers_count,following_count)
		VALUES($1,$2,$3,$4,$5,$6,$7,0,0)
		RETURNING id,name,email,password,username,nickname,bio,created_at,followers_count,following_count
	`

	resRow := u.db.QueryRowContext(ctx, query,
		user.Name,
		user.Email,
		user.Password,
		user.Username,
		user.Nickname,
		user.Bio,
		time.Now(),
	)

	// scan returned row into the passed user to populate ID and defaults
	err := resRow.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Username,
		&user.Nickname,
		&user.Bio,
		&user.CreatedAt,
		&user.FollowersCount,
		&user.FollowingCount,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}



func(u *UserDBModel) GetUserByEmail(email string) (*models.User,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		select
			 id,name,email,password,created_at
		from
			 users
		where
			 email=$1
	`
    res	:= u.db.QueryRowContext(ctx,query,email)
	var user models.User
	err := res.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			//  explicitly return this err
			return nil,sql.ErrNoRows
		}
		return nil,err
	}
	
	return &user,nil

}


// method that updates already existed user password - need newhash to be stored
func(u *UserDBModel) ResetUserPassword(newPassHash ,email string) (bool,error){
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	query := `
		Update
			users 
		set
			password=$1
		where
			email=$2
	`
    _,err	:= u.db.ExecContext(ctx,query,newPassHash,email)
	if err != nil {
		return false,err
	}
	
	return true,err
}


func(u *UserDBModel) GetUserByUserID(userID uint)(*models.User,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	// u can also use transaction tx single atomic when needed

	query := `
		Select
			 id,name,email,password,created_at,
			 COALESCE(followers_count, 0),COALESCE(following_count, 0),username,nickname,bio
		from
			users
		where
			id=$1
	`

	resRow:= u.db.QueryRowContext(ctx,query,userID) // providing user id to fetch data of that entry
	
	var user models.User
	err := resRow.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.FollowersCount,
		&user.FollowingCount,
		&user.Username,
		&user.Nickname,
		&user.Bio,

		
	)
	
	// bug - since these 2 fields are null~nil be default- use coalesce(field,defaultval) otherwise it wil fail
	//done - fixed it with selecting field as COALESCE(field,0) not concrete field
	if err!= nil {
		// there was err but query was successfull,but no result,nil struct literal return
		if err == sql.ErrNoRows {
			return nil,err
		}
		return nil,err
	}

	return &user,nil
}