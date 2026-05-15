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


// func that belongs to the UserDBModel type -> fetches user by provided name 
func(u *UserDBModel) FindUsersByName(name string) ([]*models.User,error) {

	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()

	// u can also use transaction tx single atomic when needed
	// bug - it was not fetching matching results due to a common sql syntax mistake of like 
	// fixed - do it correct by using operators which finds entries with these input in it '%X%'
	
	// bug again - putting wildcard pattern recognition ('%X%') operations failed on query
	// fixed - operation must be done in replacers placeholders arguements in the query call
	query := `
		Select
			 id,name,email,password,created_at,
			 COALESCE(followers_count, 0),COALESCE(following_count, 0),username,nickname,bio
		from
			users
		where
			name like $1
	`

	// fixed - finally it is working now ✅,note :- You should do operation in args not in raw query itself
	resRow,err:= u.db.QueryContext(ctx,query,"%"+name+"%") // providing user id to fetch data of that entry
	
	if err!= nil {
		return nil,err
	}

	// go through each resulted row
	var usersFound []*models.User
	for resRow.Next() {
		// scan each field and populate data if not error 
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
		if err != nil {
			return nil,err
		}
		// after each iteration, append successfull hit into the []
		usersFound = append(usersFound, &user)
	}

// return it
	return usersFound,nil
}


func(u *UserDBModel) FetchUserDetailsFromPosts(limit int) (*string,error) {
	ctx,timeout := context.WithTimeout(context.Background(),utils.DbTimeoutDuration)
	defer timeout()


	// * select (whatever u want from joined tables) -> when joined on common entries both tqbles join and we select (selective row) from one big joined table
	query := `
		Select
			u.name
		from
			posts p
		inner join 
			users
		On
			u.id = p.user_id
		limit $1
	`

	res := u.db.QueryRowContext(ctx,query,limit)

	//* we would get specific result we store into specfic variable that store that type of data
	var resultStorer struct {
		name string
	}
	err := res.Scan(
		&resultStorer.name,
	)

	if err!= nil {
		// if scan gave no result -> nil result
		if err == sql.ErrNoRows {
			return nil,sql.ErrNoRows
		}	
		return nil,err
	}

	return &resultStorer.name,nil
	
}