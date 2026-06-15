package testor

import (
	"log/slog"
	"strings"
)

// ! note - '_test.go' are only limited to test environment,they are invisible to the rest of the codebase
// bug - func/utility is getting undefined imported from same package but from test files
// that's why that validator function was getting undefined
// fix - under same package,but moved out of test file env and added to the general .go file

// func that takes in password string and validates it accordingly
func ValidPasswordTesting(password,email string) (statusResponse string,validated bool) {
	
	// takes up value according to the condition passed
	var status string //** variable for status string

	// custom password validators for testing pass in isolate space - does not invoke any app service
	maxPasswordlength := 15 // max pass length alllowed
	// !if inputted pass is execeeding the allowed max password length -> return false with correct status 
	if len(password) > maxPasswordlength {
		status = "password length cannot be more than 15 characters❌."
		return status,false
	}

	if email == password {
		status = "password cannot be same as email❌."
		return status,false
	}

	// for splitting strings part from a central/ref point // takes in wholeString,and seperator identifier -> solits into two before and after that identifier
	mailIntial,_,hasFound := strings.Cut(email,"@") //some@gmail.com => identifier "@" -> splits from here, before and after @ becomes first and second return values and last being bool of success 
	
	// we would have extracted mainInitial part, putting some test on it to check if they dont match the name as that would be common password for guessing
	// if it has found, it contains @ part -> then split <- checked before splitting - otherwise other val would become nil
	if !hasFound{
		status = "please enter valid email❌."
		return status,false
	}
	// map of cases where - key being the common pass set to true for those password which are very insecure
	commonPasswords := map[string]bool {
		"12345678" :true,
		"abcd1234" :true,
		"qwertyui" :true,
		"johndoe" :true,
		"password" :true,
		"Password" : true,
	}

	// if password falls under common { small intentional list of common pass} - we should check it before any big test to return early
	_,hasfoundInsecurePassword := commonPasswords[password] // if inputted password exists in map { comma _ method},means that key exists
	// declare it insecure
	if hasfoundInsecurePassword {
		status = "very common and insecured password❌."
		return status,false
	}

	
	// seperate cases for case check too
	// first checking exact matching 
	if mailIntial == password {
		status = "password cannot be same as email intials❌."
		return status,false
	} 

	lowerMail := strings.ToLower(mailIntial)
	lowerPassword := strings.ToLower(password)
	// then checking if on same cassed
	if  lowerMail ==  lowerPassword{
		status = "found alternated casing; but still password cannot be same as email intials;❌."
		return status,false
	}



	// if pass does not voilate test condition -> mark passed
	status = "passed"
	slog.Info("password testing function executed successfully","result",status)
	return status,true
}

