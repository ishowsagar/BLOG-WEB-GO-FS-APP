package utils

import (

	"golang.org/x/crypto/bcrypt"
)

// types

// password type struct which stores both PlainTextPassword and hashed pass
type Password struct {
	PlainTextPassword string
	Hash []byte
}


//  func that returns hashed byte -> in form of crypted pass
func SetHashedPassword(plainTextPassword string) []byte {
	hash,err := bcrypt.GenerateFromPassword([]byte(plainTextPassword),15) // store hash in bytes
	if err != nil {
		return nil
	}
	return hash
}

//  func that checks if passed hashed and normal passes match --> useful when login -> check passed pass string and hash stored in db matches
func CheckHashedPass(clientPass string,storedHash []byte)(bool,error) {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash),[]byte(clientPass))
	if err != nil {
		return false,err
	}
	return true,nil
}