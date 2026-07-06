package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// pipe which pipes data from one src to another - in goroutines

// func that reads from (file)src from where it reads bytes and writes via iowriter
func PipeWriter(filename string,writer io.WriteCloser) {
	
	// regarless of anything close writer always

	defer writer.Close() //defered call to close writer operations -> syscall to the kernel -> flush pending buf writes + release related fds
	// since they must be called in goroutine - lweight threads handles data exchange manaed by scheduler
	
	// read content from file
	buf := make([]byte,128)
	
	// open file ( only for reads only)
	file,err := os.Open(filename)
	if err != nil {
		return 
	}
	
	defer file.Close()

	// read from file
	for {
		rBy,err := file.Read(buf)
		if rBy > 0 {	
			// only fires when read bytes len is valid and more than zero
			_,err := writer.Write(buf[:rBy])		// since this a pipe writer it is writing to the pipe 
			if err != nil {
				return 
			}
			//! close writer when it has done reading and stream finished -> kernel finishes the pipe to prevent forever block/freeze of the application
			
		}


		// when it hits EOF -> end of stream read 
		if err == io.EOF{
			break // !break the loop which closes invokes defered call file.Close
		}

		if err != nil {
			return 
		}
	}
}


// oh....we need two from where it reads from pipe reader and one that writes into it

// func that reads cnt from pipe Reader, prints the desired txt string
func PipeReader(reader io.Reader,grepWord string) {

	// read,err := io.ReadAll(reader)
	

	// if it hits err as stream is ended
	// eof is reduntant on the io.Readall, only when stream read it makes sense
	// if err != nil  {
	// 	return
	// }


	// & upgrading to grep words out
	// todo - instead of printing full - just printing chunk which has "hello" text strinng 
	//** we know buffio.Scanner reads bytes from sources actually consumes it -> filter from there

	// grepWord := "func"
	scanner := bufio.NewScanner(reader)
	var accum strings.Builder //for tracing down grepped prompted text - "hello" hardcoded from now
	// for each scanner -> if returns true -> iterate on that loop chunk -> consume and filter down
	var counter int16
	for scanner.Scan() {
		// bug - scanner already gives line by line - alwayd - no need to break furthur
		// fix - get text line directly and filter out
		// filter text which user asked for
		// chunkByte := scanner.Bytes()
		
		line := scanner.Text()
		// reader :=bytes.NewReader(chunkByte)
		// buf := make([]byte,50) //buf size -- covers nearby relevant bytes

		// could break chunk into small buf reads  - 
		// for {
			// bufReadBytes,err := reader.Read(buf)
			
			// only when read bytes
			// if bufReadBytes > 0 {
				// chunkText := string(buf[:bufReadBytes])
				// if !strings.Contains(chunkText,"hello"){
					// continue  //skip iteration if does not include this word
				// } //if grep is "hello"
					//  strings builder for not just better output but better ux experience too
					// accumulating it
				if strings.Contains(line,grepWord) {

					// log.Fatalf("found match - %s",line)
					counter++
					// stores in the memory
					c := fmt.Sprintf("%d.",counter)
					accum.WriteString("\n")
					accum.WriteString(c)
					accum.WriteString(line)
				}
				// if contained the word :
				
			// }

			// if err == io.EOF{
			// 	break //break loop
			// }
			// if err != nil{
			// 	return //return function
			// }

			// find which contains "favorable" text - print those relevant text chunk lines only
	}// scanner loop
	if err := scanner.Err();err != nil {
		return
	}


	// get final text builded string
	str := accum.String()
	if str == "" {
		log.Printf("pipe stream finished with 0 matches from word - %s\n", grepWord)
	} else {
		fmt.Println("\n\n ------ Grep Result 🔎 ------", str)
	}
	// fmt.Print(str)
}