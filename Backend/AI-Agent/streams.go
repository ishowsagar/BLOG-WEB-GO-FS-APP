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


	// * refined definition -> stdin,stdou,stderr are just streams of data attached to every process ( executed program comes with these 3 
	// scanner is just reading from source from where it can read bytes
	// stdout -> stream where bytes of data is sent to the terminal, where bytes are displayed in text 
	// stdin -> stream where bytes of data is captured from keyboard input from terminal as when user inputs into the terminal -> bytes are captured in stdin
	// stderr -> just a seperate stream which captures error bytes
	// " pipes are way to pipe down streams to another source/destination instead of terminal, bytes could be sent or captured from stdin and piped to another destination like scanner" 

	// contexualizaton part => every program/thing would have these streams opened -> that's why our programme could read bytes through stdin,send fmt.print() bytes via stdout,err via stderr
	// and ultimately -> piped to scanner which just needs a source from where it can read bytes is stream which we already have -> stdin.

		// 	stdin, stdout, and stderr are three byte streams automatically attached to every process by the operating system when the process is created.

		// stdout is an output stream where a program writes normal bytes.
		// stderr is another output stream dedicated to errors.
		// stdin is an input stream from which the program reads bytes.

		// By default, when a process is launched from a terminal, these streams are connected to that terminal:

		// stdin receives bytes coming from the terminal (usually typed by the user),
		// stdout sends bytes to the terminal for display,
		// stderr sends error bytes to the terminal separately
		
		// ! imp
		//! but these are just stream - raw streams, cause these stream are declared on one end ( this running go process ) that does not automaticaly means
		//! like stdin stream knows it can be used to read bytes into program,but where to read into program and from where becomes the question?

		// ** Ans - context of stream ends and piping becomes the answer to this question as every process comes with these streams, but these streams does not have
		// ** any prior context that we have way to send bytes and recieve data from
		// & eg - 1. stdout is attached to it, but does not know what to send and where to,so fmt() ( would need a name later) sends bytes to terminal from process's stdout stream,
		// & 2. sane way stdin stream recieves bytes but does not know what and from where, so but in case of this stdin is default set to be with temrinal,so it knows whatver is inputted into the ternial recieved in the stdin stream,and which becomes ioreader source for scanner implmentation
		// that's why program's stdin stream becomes the source for scanner to read bytes from 
		


		// ** i/o bytes abstraction 
		// so like it scopes down to one thing- ultimately all is being done is bytes, 
		// writes reads we need a way to read and write bytes effectively.
		//  we are dealing with read writes simply, thats why this would not be solution if we read bytes from dedicated roles/fnc
		//  since bytes could be anything so anything could be read and in chunks -
		//  so this abstraction makes sense-
		//  you would be either reading bytes from source ( http...networks), or writing to some destination

		// bytes 
		// 1. At the operating system level, everything eventually becomes bytes. Files, network packets, HTTP bodies, terminal input, images, JSON,
		//  and videos all reduce to streams of bytes.

		// 2.  Since every source ultimately produces bytes and every destination ultimately consumes bytes, 
		// Go doesn't abstract files or networks. Instead, it abstracts the movement of bytes.

		// 3. A Reader represents anything capable of producing a stream of bytes.
		//  A Writer represents anything capable of consuming a stream of bytes.

		// 4. This gives the entire Go ecosystem one common language. 
		// Packages don't need to know whether they're reading from a file, stdin, a socket, or an HTTP response.
		// They only care that bytes can be read or written. 


		// ioreader/writer 
		// This isn't just define a source from where bytes could be read or to write but it describes if a something is satisfying these interfaces universally
		//** that means those are satisfying go core philosphy around working with bytes :
		// 1. it considers something a reader source when it is universally satisfying condition that -> it could be used to read bytes anywhere as reader
		// 2. since it implements reader -> everywhere bytes read/writes abstraction comes into play -> this could be used there to read/write byte into

		// refined explanation :
		// io.Reader / io.Writer
			//
			// These don't know how to read or write bytes.
			// They only define the universal contract for producing or consuming bytes.
			//
			// Go's philosophy is that every system eventually communicates through bytes.
			// Instead of creating separate APIs for files, stdin, sockets, HTTP bodies,
			// buffers, etc., Go asks:
			//
			// "Can this type produce bytes?"
			//      ↓
			// It satisfies io.Reader.
			//
			// "Can this type consume bytes?"
			//      ↓
			// It satisfies io.Writer.
			//
			// Once a type satisfies these contracts, the entire Go ecosystem can work
			// with it without caring where the bytes come from or where they go.
			//
			// The actual logic for obtaining or consuming bytes lives inside the concrete
			// type (File, Socket, Buffer, HTTP Body, stdin, etc.), not inside the interface.

	// NewScanner ( scans for whatever the user inputs <- stdin) -> needs ioreader <- a sources from where it can read bytes~data from
	
	// ** proccess comes with stdin stream ( satisfies universally accepted bytes implementation philosphy and implements reader interface ) - global bytes reader source
	// by default - captures bytes from terminal and since piped into process <- accessed as ioreader => provides reader source from where user input bytes are read
	scanner := bufio.NewScanner(os.Stdin) //* os.Stdin implements ioreader <- cause data coming from the terminal is sourced via stdin which wraps data and act as reader

	// *scanner reads bytes data from stdin ioreader source <- whatever user types -> intercepted by the stdin reader -> which is piped down to scanner for capturing inside the program
	// so scanner is intialized for scanning user input

	// flow
	// since it provides gateway for interaction, anything could be printed or recieved { vague but for now}
	// cause we will print something, capture input and send out
	cyanPrompt := "\033[36m"
	resetPen   := "\033[0m"
	// 1. printing this to terminal <- since it is going out to the terminal ( fmt under the hood do it with os.Stdout{vague}) 
	fmt.Print(cyanPrompt+confirmationOutMsg+ resetPen) 
	
	// 2. capturing input - as scanner reads from ioreader <- stdin becomes the source from where it reads bytes (INPUT) from
	scanner.Scan() // scan untill input has not been recieved ( user must send input ( from keyboard) )

	inputTxt := scanner.Text()
	// res := fmt.Sprintf("recieved user response as '%s'.",inputTxt) // formatting meths uses pholders
	return inputTxt
	// **successfully implemented✅✅ -> scanner gets the pipe stream from stdin { bytes coming from the terminal } -> scans them and stores in txt 
	// **and when we do fmt.printf(output) -> we are actually streaming bytes~data to the terminal ->sending output to display.
}



