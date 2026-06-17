package main

// ! go does let us create multi main*entry packages,when parent dir is diffren
import (
	"fmt"
	"io"
	"os"
)

//important concepts
// io.Reader type interface methods -> let reciever side read data in chunks [slice of bytes]
// buffer - just allocated slice of byte data with allocated warehouses as bytes data is raw data, it could be anything from img to text
//  reader/writer -> uses/takes in  buffers to let reciever consume the data in chunks or write... it is used in all scenerios - networks/http requests,
// terminal out - stdOut, writing to local files...

// ai agentic workflow - we'll be using deepseek api for this
// 1. read file content
// 2. send prompt to the ai
// 3. get response
// 4. writer response to local file/chat interface


func ReadFileContent(fileRelPath,fileToReadFrom string)(string,error) {

	// &os package helps to interact natively with operating system and its components gracefully
	var targetFile string
	
	// dynamically getting file - "./+agent.txt" -> first part becomes filePath+fileName
	targetFile = fmt.Sprintf("%s%s","./",fileToReadFrom)
	// fmt.Println("reading from - ",fileToReadFrom,"...")
	// if target is empty -> defaults
	if targetFile =="" {
		targetFile = "./agent.txt" // path rel to from where go binary is ran from e.g here - find out
	}
	// 1. retrieve file
	openedFile,err := os.Open(targetFile)
	if err != nil {
		return "",fmt.Errorf("failed to open file, err -%v",err)
	}
	defer openedFile.Close()
	
	// 2. create buffer -  buffered slice of byte type elements, which allocates warehouses for storing raw bytes data 
	
	//! buffer act as plate bearer - plate is emptied after one chunk (say) is read in bytes ->
	buffer := make([]byte,500) //500 kb buffer which stores data of type bytes 
	var totalbytes int 
	var content string

	// we will 'read the content into buffer' - since buffer takes in bytes data, file feeds in that data to the buffer and tells how much is writter/populated
	//* read in loop as chunks bytes would be handled in chunks
	// since [] of bytes stores data in byte as slice elements, each has its position , bytes written = max filled slice
	for {
		// ** chunk read into buffer ( say 'plate' ) flow
		// - chunk of bytes data from file is read into buffer -> returns how much is read, since buffer stores it at a time
		// - retrieves that stringified content from the buffer upto readed bytes n { tells till how many bytes are wrriten,1 byte consume 1 space in slice}
		// - that gives us -> exact buffer slice position till it is written,we extract from it and convert it to string
		currentChunkWrittenBytes,err := openedFile.Read(buffer) //returns how much bytes data is successfully written to buffer
		if err != nil {
			// file is read fully when it returns end of file (eof) err - signals done reading chunks
			if err == io.EOF {
				// ! we don't wanna return to crash and return early 
				// * but since loop operation has completed - break out of loop safely
				break
			}
			return "",fmt.Errorf("failed to populate file content into the 'buffer', err -%v",err)
		}
		// accumulating chunks - it is read into buffer { storing data into empty slice of bytes elements}
		// bug - we are overwriting file, we are only retreiving and accounting for how much bytes have been written,but we are losing track of actual content
		// fix - we need to track each chunk ( as cursor moves forward to chunk, reads one at a time) and store in the buffer upto n length each time
		// fixed - we store in each time it reads chunk into content
		totalbytes += currentChunkWrittenBytes
		
		// & this is how, each time it reads chunk into buffer -> returns n upto what slice length it has wrote bytes into buffer, acculamte it in content variable to store chunks
		// 3.converting succcessfully read bytes data into strings 
		content = string(buffer[:currentChunkWrittenBytes])
		// * since buffer would have been populated by , but we don't know how much file has written to the buffer which accepts bytes data 
		//* so we are taking buffer upto that slice length till it had successfully written bytes data
	}

		// fmt.Printf("total bytes read from file -%v\n",totalbytes)
		// fmt.Printf("file content -%s\n",content)

		// this method writes content of type slice of bytes data and cretes file if not does not earlier
		
		return  content,nil
}

// func which writes content to the specified file ; if not exists -> creates
func WriteToFile(fileName,content string)(error) {
		err := os.WriteFile(fileName,[]byte(content),0644) //last arg is permission set for operation
		if err != nil {
			return fmt.Errorf("failed to create file, err -%v",err)
		}
		
		// fmt.Printf("reading from file -%s/n, content - %s/n",fileName,content)
		// fmt.Printf("successfully generated '%s' file and feeded AI response to it.",fileName)ho
		return nil
}



