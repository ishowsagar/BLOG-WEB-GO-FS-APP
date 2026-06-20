package main

import (
	"bufio"
	"fmt"
	"os"
)

// @ Streams -> everything either going as output to the terminal or incoming as input to the program from the terminal is streamed via something.
// stdin -> scanner takes in ioreader stdin source <- streams pipe to the scanner <- so it reads bytes from the stdin
// stdout -> vague as not done yet but it also streaming bytes data ouput to the terminal with iowriters.

// Stdout and stdinput -> these are the standard streams, which interacts with shell from the program or to the program
// stdout =>  Data sent from program to the shell/terminal for display - output ( all things under the hood goes in bytes) going to the terminal
// stdin => data sent to the program by the shell/terminal - user input/ stdin coming the terminal

// these are streams of data, data is recieved and sent in bytes therefore implementing ioreader/writers.

// & Both are implemented to work with []bytes data, bytes data is being exchanged,

// reads stdin from the terminal amd prints it to the terminal back by printing via stdout as it goes out to the terminal
func CaptureInputFromShell(confirmationOutMsg string) (string) {


	// NewScanner ( scans for whatever the user inputs <- stdin) -> needs ioreader <- a sources from where it can read bytes~data from
	scanner := bufio.NewScanner(os.Stdin) //* os.Stdin implements ioreader <- cause data coming from the terminal is sourced via stdin which wraps data and act as reader

	// *scanner reads bytes data from stdin ioreader source <- whatever user types -> intercepted by the stdin reader -> which is piped down to scanner for capturing inside the program
	// so scanner is intialized for scanning user input

	// flow
	// since it provides gateway for interaction, anything could be printed or recieved { vague but for now}
	// cause we will print something, capture input and send out

	// 1. printing this to terminal <- since it is going out to the terminal ( fmt under the hood do it with os.Stdout{vague}) 
	fmt.Print(confirmationOutMsg) 
	
	// 2. capturing input - as scanner reads from ioreader <- stdin becomes the source from where it reads bytes (INPUT) from
	scanner.Scan() // scan untill input has not been recieved ( user must send input ( from keyboard) )

	inputTxt := scanner.Text()
	// res := fmt.Sprintf("recieved user response as '%s'.",inputTxt) // formatting meths uses pholders
	return inputTxt
	// **successfully implemented✅✅ -> scanner gets the pipe stream from stdin { bytes coming from the terminal } -> scans them and stores in txt 
	// **and when we do fmt.printf(output) -> we are actually streaming bytes~data to the terminal ->sending output to display.
	}