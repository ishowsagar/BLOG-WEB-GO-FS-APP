package testor

import (
	"fmt"
	"testing"
	"time"
)

//@ Unit testing => testing a single unit\component out of the system to check its working in a isolated space.
// this file name must end with - "_test"
//*types

// type struct for user details data for unit testing
type UserDetails struct {
	email string
	password string 

	//bug 1 - since test does not just normal testing, we perform tests expecting some expected outcomes, so we must consider them for case
	// fixed - so when cases would be created, each case struct data must had expected results attached -> so once we ran them all in r.Run
	//test =  we can assess them and based on expectation - filter results
	expectedStatus string
	expectedValid bool

}




// ** unit testing function **//
// bug 2 - testing operator function ( which is actually invoking tests) must start with "Test" initial
// fixed - now test package would know and recognize to auto do testing based off blueprint we wrote
func TestValidPassword(t *testing.T) {

	// 1..X -> mental mapping coding logic + makes it feel inituitive
	
	// 1. Cases Slice -  of input parameters (e.g user details, slice of elements structs with user detials data)
	caseSameInput  := UserDetails{
		email: "ayush@gmail.com",
		password: "ayush@gmail.com",
		
		// Test - we pass expected results cause we would not know before hands - these should fail or pass on the fly
		expectedStatus: "password cannot be same as email❌.",
		expectedValid: false,
	}	

	caseSameIntial :=UserDetails{
		email: "denver@gmail.com",
		password: "denver",
		expectedStatus: "password cannot be same as email intials❌.",
		expectedValid: false,
		
	}

	// caseUnknownEmailAddress :=UserDetails{
	// 	email: "denver%gmail.com",
	// 	password: "huskeyTheDuskey",
	// 	expectedStatus: "unknown email address❌.",
	// 	expectedValid: false,
	// }

	caseCommonPassword  :=UserDetails{
		email: "abc@gmail.com",
		password: "abcd1234",
		expectedStatus: "very common and insecured password❌.",
		expectedValid: false,
	}

	caseOverlengthedPassword  :=UserDetails{
		email: "abc@gmail.com",
		password: "thisisverylongpasswordandsecurethiswillbemybestpassword",
		expectedStatus: "password length cannot be more than 15 characters❌.",
		expectedValid: false,
	}

	caseValid  :=UserDetails{
		email: "ayushkumar1798@gmail.com",
		password: "viking@hero",
		expectedStatus: "passed",
		expectedValid: true,
	}

	cases := []UserDetails{
		caseSameInput,
		caseSameIntial,
		caseCommonPassword,
		caseOverlengthedPassword,
		caseValid,

	} //..cases slice


	// & by looking at the cases -> we can add more testing paramter on validator like when both fields are same,common pass,initial passwords are used
	
	// 2. loop over slice elements -> run testing function and each ietration invokes tests
	for _,eachCase := range cases{
		// 3. invoke testing function on every current iteration which -> does press test on each current ietration
		
		testIdentifier := fmt.Sprintf("user-%s-%d",eachCase.email,time.Now().Unix())
		// using t.Run method to run -> subtests on these cases, need each subtest identifier for running it
		t.Run(testIdentifier,func(t *testing.T) {
			// testidentifier -  being the identifier

			// invoke testing func on each case
			resultingStatus,hasTestBeenValidated := ValidPasswordTesting(eachCase.password,eachCase.email) //being explicit as we can
			
			
			//!IMP - asserting errors only if expected results does not match outcome { we want outcome matching res, validator is just for invocation}
			if resultingStatus != eachCase.expectedStatus || hasTestBeenValidated != eachCase.expectedValid {
				
				// would print err of type t.Errof() and return immediatly -> true if successful / false if failed the test,not test failed itself 
				t.Errorf("successfully ran test for '%s' user, expected(%s,%t) result(%s,%t) ",eachCase.email,
				eachCase.expectedStatus,caseCommonPassword.expectedValid,
				resultingStatus,hasTestBeenValidated,
				) //erf auto return with formatted err string return

				// !note - unit tests are for testing a func/comp, so checking if it is working perfectly to be integrated
				// ** clarity -> we still would be using the proven "tested func/comp" cause it now works for all cases ✅
			}
		})
	}

} 




