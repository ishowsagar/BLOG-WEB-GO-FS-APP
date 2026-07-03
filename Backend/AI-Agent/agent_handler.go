package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fsnotify/fsnotify"
)

// builds prompt based on provided mode ["docs"...] and append content to it for sending request
func BuildPrompt(mode, content string) string {

	// * goal - return desired prompts based on mode selection

	// mode validation., must have 3 allowed modes only
	allowedModes := map[string]bool{
		// allowed modes only
		"review": true,
		"docs":   true,
		"qa":     true,
	}

	// check if it exists in allowed modes map -> if key exists -> mode is available to be prompted in
	_, available := allowedModes[mode]
	if !available {
		log.Fatalf(" this mode - '%s' is not available", mode)
	}

	// if provided mod exists -> generate prompts based off that
	var prompt string
	switch mode {
	case "review":
		prompt = "You are an expert code reviewer, review the provided code and give me the short review upto 100 words only but expert review"
	case "docs":
		prompt = "You are an expert code docs generator, analyse the provided code and document it upto 100 words only but expert documentation"
	case "qa":
		prompt = "You are an expert coding mentor, review the provided code and ask me max 5 question in a way it makes me click topic better "
	default:
		prompt = "You are an expert code reviewer, review the provided code and give me the short review upto 100 words only but expert review"
	}


	// since assigned only but need to actually do the intended job - return it
	return prompt +"\n\n" + content //* returning both with line seperation  
	
}

func BuildDeepRagPrompt(mode,ragChunkQueryText,QueryQuestion string) string {

	// * goal - return desired prompts based on mode selection

	// mode validation., must have 3 allowed modes only
	allowedModes := map[string]bool{
		// allowed modes only
		"review": true,
		"docs":   true,
		"qa":     true,
	}

	// check if it exists in allowed modes map -> if key exists -> mode is available to be prompted in
	_, available := allowedModes[mode]
	if !available {
		log.Fatalf(" this mode - '%s' is not available", mode)
	}

	// if provided mod exists -> generate prompts based off that
	var prompt string
	switch mode {
	case "review":
		prompt = "You are an expert code reviewer, review the provided code and give me the short review upto 100 words only but expert review"
	case "docs":
		prompt = "You are an expert code docs generator, analyse the provided code and document it upto 100 words only but expert documentation"
	case "qa":
		prompt = "You are an expert coding mentor, review the provided code and ask me max 5 question in a way it makes me click topic better "
	default:
		prompt = "You are an expert code reviewer, review the provided code and give me the short review upto 100 words only but expert review"
	}


	// since assigned only but need to actually do the intended job - return it
	return prompt + "\n\nRelevant code context:\n" + ragChunkQueryText + "\n\nUser question: " + QueryQuestion //* returning full context of rag + memory  
	
}

// builds prompt based on provided mode ["docs"...] and append content to it for sending request
func BuildDirPrompt(mode, content,userPrefsCmd string) string {

	// * goal - return desired prompts based on mode selection

	// mode validation., must have 3 allowed modes only
	allowedModes := map[string]bool{
		// allowed modes only
		"review": true,
		"docs":   true,
		"qa":     true,
	}

	// check if it exists in allowed modes map -> if key exists -> mode is available to be prompted in
	_, available := allowedModes[mode]
	if !available {
		log.Fatalf(" this mode - '%s' is not available", mode) //direct fatal return
	}

	// only then check if userPrfCmd { user special cmd } is not nil then replace with it

	dirContext := " and remember it contains all files data in one go, anaylyse files and related code, code is seperated by files boundaries marked with lines like starting with this file name and ends with this file name,for clearance"
	fileContextConfirmation := "and label all the files you reviewed"
	Overrider := "Must follow this special command, override prev if clashing with any ->"+userPrefsCmd
	// if provided mod exists -> generate prompts based off that
	var prompt string
	switch mode {
	case "review":
		prompt = "You are an expert code reviewer, review the provided code and give me the short review upto 100 words only but expert review" + dirContext + fileContextConfirmation + Overrider
	case "docs":
		prompt = "You are an expert code docs generator, analyse the provided code and document it upto 470 words only but human way documentation" + dirContext + fileContextConfirmation + Overrider
	case "qa":
		prompt = "You are an expert coding mentor, review the provided code and ask me max 5 question in a way it makes me click topic better " + dirContext + fileContextConfirmation + Overrider
	default:
		prompt = "You are an expert code reviewer, review the provided code and give me the short review upto 100 words only but expert review" + dirContext + fileContextConfirmation + Overrider
	}


	// since assigned only but need to actually do the intended job - return it
	return prompt +"\n\n" + content //* returning both with line seperation  
	
}

// builds prompt based on provided mode ["docs"...] and append content to it for sending request
func BuildGitPrompt(gitMode, content string) string {

	// * goal - return desired prompts based on mode selection


	commitContext := "Generate ONLY the commit message text, no explanation, no bullet points, just the message itself ready to paste into git commit -m"
	reviewContext := "response should not exceed 100 words, bullet points, ready to fire commands to fix issues"
	// if provided mod exists -> generate prompts based off that
	var prompt string
	switch gitMode {
	case "status":
		prompt = "You are an expert coding mentor, review the provided code and give me the short output summary what is current git status and what needs to be done"  
	case "add" :
		prompt = "You are an expert coding mentor, review the provided added files git ouput and " + commitContext
	case "commit":
		prompt = "You are an expert coding mentor, analyse the provided code and generate beautiful humanized relavant commit message" + commitContext 
	case "diff":
		prompt = "You are an expert coding mentor, review the provided code and read the full diff and provide useful insights on it and how to fix them " + reviewContext 
	}


	// since assigned only but need to actually do the intended job - return it
	return prompt +"\n\n" + content //* returning both with line seperation  
	
}


// func that checks if history already exists or not
func CheckConvoHistoryFILE(filename string)(historyExists bool,State error) {
	
	// checks if this file has stats -> eventually checks if file exists
	_,err := os.Stat(filename)

	// if hit err to get info and err was that file does not exists -> return false that it does not exists
	if err != nil && os.IsNotExist(err) {
		return false,nil // err nil cause its just state
	}

	// otherwise true if exists
	return true,nil
}


// writes history conversation to the history.json <- keep full history of the conversation
func HandleHistoryWrites(writeToFile string,history []byte)(error) {
	err := os.WriteFile(writeToFile,history,0644)
	if err != nil {
		return err
	}

	fmt.Println("conversation has been added to history.")
	return nil

}



func HandleHistoryReads(history []byte)([]*ContentsSliceKeyWrapperGem,error) {
	// parts := &PartsSliceKeyWrapperGem{
	// 	Text: content,
	// }
	// cnts := &ContentsSliceKeyWrapperGem{
	// 	Role: "s",
	// 	Parts: []*PartsSliceKeyWrapperGem{parts},
	// }
	// history := OutboundPayloadGem{
	// 	Contents: []*ContentsSliceKeyWrapperGem{
	// 		cnts,
	// 	},
	// }

	var payload []*ContentsSliceKeyWrapperGem
	err := json.Unmarshal(history,&payload)
	if err != nil {
		return nil,fmt.Errorf("failed to unmarshal history")
	}
	return  payload,nil
}


// func that keep running in bg and watch out for any changes either writes or deletes, if detected -> 
func WatchDirChanges(dirToWatchOut,fileSuffix,Mode,APIKEY,OutputFileName string) (watcherErr error) {

	// &watcher work flow
	// fsnotify recieves concurrent events sent from os
	// that's why intialized wactcher keep checking for incoming events, if they exists -> do the required job

	// fires go routine func which -> keep runnin in the background -> read if event is recieved on the chan

	// 1. create watcher - recieves events from the os {event :{name,op(meth)}}
	watcher,err :=fsnotify.NewWatcher()
	if err != nil {
		slog.Error("failed to intialize watcher","error",err)
		return err
	}
	// 2. add dir which has to be watched out for [write] changes
	if err:= watcher.Add(dirToWatchOut) ; err != nil {
		slog.Error("failed to intialize dir to watch out for","error",err)
		return err
	}

	// 3. infinite loop which keeps checking if there is -> incoming event recieved on the watcher's chan {recieve hora hai kya event}
	go func(dirWatcher *fsnotify.Watcher,fileSuffix,Mode,APIKEY,OutputFileName string) {

		// labelling loop for better visible layered escapes
		watcherLoop:
		for {
			// ! infinite loop -> keep running in background -> so it always keep running to recieve the events on the watcher's chan
			select {
				// select - execute the cases based satisfied cases otherwise notifies retry or chan full blockage for non-crashing concurrent operations
				case event,triggered := <- dirWatcher.Events :
					slog.Info("some event is recieved","Name",event.Name)
					if !triggered {
						slog.Error("void or nil event recieved;exiting out of loop safely","Name",event.Name)
						return //early return from the fnc
					}

					if event.Name == OutputFileName {
						slog.Error("returning early to prevent infinite call on agent itself")
						return
					}
					// * event validation - only watch out for write(op) event and on .go files only
					// if event is 'write' { when write is detected is recieved } and 
					// * since event { name,op} -> name tells us on which file event has been brewed, so -> we can check on it => if that has .go suffix, we proceed
					if event.Op.Has(fsnotify.Write) && strings.HasSuffix(event.Name,fileSuffix) {
						
						// cooldown period to avoid too many request hitting
						time.Sleep(4 * time.Second)
						// * means go file change has been detected
						slog.Info("write event detected🚨","Name",event.Name)
						
						//! since desired 'write' event on file (event.Name) has detected, need to -> invoke agent automatically to execute review cmd/ git later

						// ** calling agent for review of this file // - follows same procedure - client.Do(req) and recieves res -> write
						// bug - event.Name gives the full abosulte path, we just need to extract file name from it
						// fix - get filename from the full path so it works correctly or it will not find file and return err as specified file not found
						targetModifiedFile := event.Name
						
						// it gives full path as event.Name- C:\Users\asus\documents\GO_DEV\BLOG-WEB-GO-APP\Backend\AI-Agent\main.go
						// we need to cut it from "AI-AGENT\" -> including slash to get exact filename - but it would not cause err \ is showing new line err
						// sep := `AI-AGENT\` //or could have used filebase.base(normalStringPath)
						
						// bug - cannot use normally cut string as it contains forward slash '\' which causes an err when dumped in string value
						// fix - use them in template litreals ``
						// _,filename,found :=strings.Cut(targetModifiedFile,sep) //* prints main.go from full path as it splits from dir name as 'sep'
						// if !found{
						// 	slog.Error("could not finding target file","error","name mismatch or non existing file")
						// 	break watcherLoop
						// }

						client := http.Client{Timeout: 15 *time.Second }
						reqURL :="https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

						//1. read file content - from this target file (earlier it was provided by arg but here by th event itself)
						cnt,err := ReadFileContent("./",targetModifiedFile) // gives file name to get all content
						if err != nil {
							log.Fatalf("failed to read file content, err- %v",err)
						} 
						var parts []*PartsSliceKeyWrapperGem
						part := &PartsSliceKeyWrapperGem{
							// ! adding conditional mode prompt + content
							Text: BuildPrompt(Mode,cnt), 
						}
						parts=append(parts, part)

						contents := &ContentsSliceKeyWrapperGem{
							Role: "user", //* while requesting, telling it has user role - user req
							Parts: parts,
						}

						out := &OutboundPayloadGem{
							Contents: []*ContentsSliceKeyWrapperGem{contents},
						}

						// encoding out into bytes (ofc in []byte)
						outByteBuffer,err := json.Marshal(out)
						if err != nil {
							log.Fatalf("failed to encode outgoing payload,err -%v",err)
						}

						// wrapping payload into io reader -> let servers recieves data in chunks
						body := bytes.NewBuffer(outByteBuffer)  // buffer.buffer auto handles read/writes internally


						// 3. creating request for sending 
						req,err :=http.NewRequest("POST",reqURL,body)
						if err != nil {
							log.Fatalf("failed to request AI,err - %v",err)
						}

						req.Header.Set("Content-Type","application/json") // header key is **case-sensi 
						req.Header.Set("x-goog-api-key", APIKEY) // Pass your Gemini key here


						// added spinner
						// 1. Initialize the spinner
						// CharSets[14] is a cool dot-bouncing animation, but there are dozens of options
						s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)  
						
						var spinnerText string//notifi dynamically
						switch Mode {
						case "review" :
							spinnerText = "Agent reviewing changes..."
						case "docs" :
							// verbose for now
							// later - add timing funcitonality based spinners
							spinnerText = "Agent generating docs..."
						case "qa" :
							spinnerText = "Agent formulating qna..."
						}

						s.Suffix = spinnerText // Adds text next to the spinner
						s.Color("cyan") // Optional: give it some color!

						// 2. Start the spinner
						s.Start()

						//4. initiates the request
						res,err := client.Do(req)
						if err != nil {
							log.Fatalf("failed to get res, err - %v",err)
						}

						// 4. Stop the spinner immediately after the request finishes!
						s.Stop()
						//5. intercepting response
						// deferred calling when sorrounding everything else has fired -> invoke this to get response
						
						// ! response recieved in body - satisfies io Reader - let []byte chunks read by reciever 
						defer res.Body.Close() // also wraps []byte in ioreader <- 
						
						// !full res validation to avoid errors when [0] errorJson[1] is returned only
						if res.StatusCode != 200 {
							log.Fatalf("API returned a non-200 status code: %d", res.StatusCode)
						}


						// fetch retrieved data too in ioreader in chunks
						bodyReader := res.Body
						r,err := io.ReadAll(bodyReader) // read at once and tells how much is read
						if err != nil {
							log.Fatalf("failed to read response body byte data, err- %v",err)
						}


						var inboundResponse InboundPayloadGem
						err = json.Unmarshal(r,&inboundResponse)
						if err != nil {
							log.Fatalf("failed to unmarshal recieved response, err- %v",err)
						}
						
						// parsed response validation check
						if len(inboundResponse.Candidates) == 0 {
							log.Fatal("API returned a successful response, but the 'Candidates' array was empty.")
						}

						// as res is srved on [0] -> checking if it is coming nill
						if inboundResponse.Candidates[0].ContentWrapperGem == nil || len(inboundResponse.Candidates[0].ContentWrapperGem.Parts) == 0 {
							log.Fatal("API returned candidates, but the 'Parts' array was empty or nil.")
						}


						AIResponseContent := inboundResponse.Candidates[0].ContentWrapperGem.Parts[0].Text
						if len(AIResponseContent) == 0 {
							log.Fatal("empty response from AI, must have hit some unexpected error",)
						}


						// since choices is a slice -> access its element's msg
						err = WriteToFile(OutputFileName,AIResponseContent)
						if err != nil {
							log.Fatalf("failed to write to the file, err - %v",err)
						}

						// switch on *Mode val for response - as it telling to use switch instead
						switch Mode {
							case "docs" :
								fmt.Println("Agent has Successfully Generated documentation⚡")
							case "review" :
								fmt.Println("Agent has Successfully Analysed & Reviewed Changes⚡")
							case "qa" :
								fmt.Println("Agent has Successfully Formulated QNA⚡")	
						}//..switch

					}
				case err,triggered := <- dirWatcher.Errors :
					if !triggered {
						slog.Info("void or nil event recieved;exiting out of loop safely","error",err)
						return
					}
					slog.Info("an error has occurred","error",err)
					break watcherLoop // gets out of loop safely without breaking rest of code
			}//..select
		}
	}(watcher,fileSuffix,Mode,APIKEY,OutputFileName)//.. go-func

	return
}



// Call when needed to get git based res only <- pass selected mode and gitOut to work with
func GetGitResponse(clientSelectedGitMode,gitOutToSend,reqURL,API_KEY,writeToFileArg string,s *spinner.Spinner,client *http.Client)(string,error) {
	// build req
		var parts []*PartsSliceKeyWrapperGem
		part := &PartsSliceKeyWrapperGem{
			// ! adding conditional mode prompt + content
			Text: BuildGitPrompt(clientSelectedGitMode,gitOutToSend), 
		}
		parts=append(parts, part)

		contents := &ContentsSliceKeyWrapperGem{
			Role: "user", //* while requesting, telling it has user role - user req
			Parts: parts,
		}

		out := &OutboundPayloadGem{
			Contents: []*ContentsSliceKeyWrapperGem{contents},
		}

		// encoding out into bytes (ofc in []byte)
		outByteBuffer,err := json.Marshal(out)
		if err != nil {
			return "",fmt.Errorf("failed to encode outgoing payload,err -%v",err)
		}

		// wrapping payload into io reader -> let servers recieves data in chunks
		body := bytes.NewBuffer(outByteBuffer)  // buffer.buffer auto handles read/writes internally


		// 3. creating request for sending 
		req,err :=http.NewRequest("POST",reqURL,body)
		if err != nil {
			return "",fmt.Errorf("failed to request AI,err - %v",err)
		}

		req.Header.Set("Content-Type","application/json") // header key is **case-sensi 
		req.Header.Set("x-goog-api-key", API_KEY) // Pass your Gemini key here


		// added spinner
		// 1. Initialize the spinner
		// CharSets[14] is a cool dot-bouncing animation, but there are dozens of options
		
		var spinnerText string//notifi dynamically
		switch clientSelectedGitMode {
		case "status" :
			spinnerText = "Agent checking status..."
		case "commit" :
			// verbose for now
			// later - add timing funcitonality based spinners
			spinnerText = "Agent commiting changes..."
		case "diff" :
			spinnerText = "Agent checking repo changes..."
		}

		s.Suffix = spinnerText // assigns the spinner suffix text to be this 
		s.Color("cyan") // Optional: give it some color!

		// 2. Start the spinner
		s.Start()

		//4. initiates the request
		res,err := client.Do(req)
		if err != nil {
			return "",fmt.Errorf("failed to get res, err - %v",err)
		}

		// 4. Stop the spinner immediately after the request finishes!
		s.Stop()
		//5. intercepting response
		// deferred calling when sorrounding everything else has fired -> invoke this to get response
		
		// ! response recieved in body - satisfies io Reader - let []byte chunks read by reciever 
		defer res.Body.Close() // also wraps []byte in ioreader <- 
		
		// !full res validation to avoid errors when [0] errorJson[1] is returned only
		if res.StatusCode != 200 {
			return "",fmt.Errorf("API returned a non-200 status code: %d", res.StatusCode)
		}


		// fetch retrieved data too in ioreader in chunks
		bodyReader := res.Body
		r,err := io.ReadAll(bodyReader) // read at once and tells how much is read
		if err != nil {
			return "",fmt.Errorf("failed to read response body byte data, err- %v",err)
		}


		var inboundResponse InboundPayloadGem
		err = json.Unmarshal(r,&inboundResponse)
		if err != nil {
			return "",fmt.Errorf("failed to unmarshal recieved response, err- %v",err)
		}
		
		// parsed response validation check
		if len(inboundResponse.Candidates) == 0 {
			return "",fmt.Errorf("API returned a successful response, but the 'Candidates' array was empty.")
		}

		// as res is srved on [0] -> checking if it is coming nill
		if inboundResponse.Candidates[0].ContentWrapperGem == nil || len(inboundResponse.Candidates[0].ContentWrapperGem.Parts) == 0 {
			return "",fmt.Errorf("API returned candidates, but the 'Parts' array was empty or nil.")
		}


		AIResponseContent := inboundResponse.Candidates[0].ContentWrapperGem.Parts[0].Text
		if len(AIResponseContent) == 0 {
			return "",fmt.Errorf("empty response from AI, must have hit some unexpected error",)
		}


		// 5. write res for now <- if untill this everything works -> fire git to do requested work
		err = WriteToFile(writeToFileArg,AIResponseContent)
		if err != nil {
			return "",fmt.Errorf("failed to write to the file, err - %v",err)
		}
		
		
		// 6. notifying client with res 
		var resMsg string
		switch clientSelectedGitMode {
		case "status" :
			resMsg = fmt.Sprintf("Agent has done checking status & response awaits your essence in %s file📂",writeToFileArg)
		case "commit" :
			// verbose for now
			// later - add timing funcitonality based spinners
			resMsg = fmt.Sprintf("Agent has commited changes & response awaits your essence in %s file📂",writeToFileArg)
		case "diff" :
			resMsg = "Agent has analysed working tree and response awaits your essence in %s file📂 "
		}
		
		fmt.Println(resMsg)
		return  AIResponseContent,nil
}


// func that belongs to type AgentConfig which -> runs dir calls
func(acfg *AgentConfig) RunAgent() {
	if acfg == nil {
		return
	}

	// so we don't want nil struct to be called on
	// todo- add logic to call for resp and all
	

}


// func that writes slice [] of bytes data to the respective destination <- based on which io writer is called on 
func BytesWriter(writer io.Writer,bytesData []byte) (bytesWritten int,er error) {

	// io.writer is an interface -> it is an interface which stores common writer method ( cause commonly owned by all ), and we know interface is
	// implemented by type which -> has these methods on it.
	
	// so interfaces are satisfied and implemented by any type that has those method which are being invoked -> so if any type has same method -> underlying interface struct calls that method


	// cause any type that has methods belongs to it ( which interface expects ) -> satisfies writer interface
	// based on which writer would have been passed
	bytesWritten,err := writer.Write(bytesData)
	if err != nil {
		return 0,err
	}

	// if bytes would have been successfully written <- gives n 
	return bytesWritten,nil
}

// writes strings data based off called writer type
func StringsWriter(writer io.Writer,data string) (int,error) {

	// io.writer is an interface -> it is an interface which stores common writer method ( cause commonly owned by all ), and we know interface is
	// implemented by type which -> has these methods on it.
	
	// so interfaces are satisfied and implemented by any type that has those method which are being invoked -> so if any type has same method -> underlying interface struct calls that method


	// cause any type that has methods belongs to it ( which interface expects ) -> satisfies writer interface
	// based on which writer would have been passed
	bytesWritten,err := writer.Write([]byte(data))
	if err != nil {
		return 0,err
	}

	// if bytes would have been successfully written <- gives n 
	return bytesWritten,nil
}

// func that writes bytes which are<- read from ioreader source - not static bytes data
func StreamlinedWriter(writer io.Writer,source io.Reader) (int,error) {

	readBytes,err := io.ReadAll(source)
	if err != nil {
		return 0,err
	}

	// writer interface has method which belongs to it -> 'if any type satisfies and implements writer interface -> would call that write method idomatically'
	bytesWritten,err := writer.Write(readBytes)
	if err != nil {
		return 0,err
	}

	return bytesWritten,nil

}

// writes bytes read from source in chunks
func FlowBytes(writer io.Writer,reader io.Reader) (int64,error) {

	//! But this flow is still very static -> we need to wrap reader source in a way that ->
	//! it reads in chunks from reader, gives data in chunks/lines because io.copy has 32kb intial buffer <- which is real problem
	

	// bug - io.copy buffer size is 32kb -> breaking source bytes into chunks/lines untill data is not fully

	// reads bytes from source and writes to the destination based on which type has satisfied and implemented writer
	writtenBytes,err := io.Copy(writer,reader)

	if err != nil {
		return 0,err
	}

	return writtenBytes,nil
}


// writes data to the file - underneaths creates file and implies file writer 
func FileBytesWriter(data string,filename string)( error) {

	// bug - it creates file on every chunk and replace old chunk with new chunk and file
	// fix - need to pass flags which makes it so => if that exists ->append  to it, won't recreat

	file,err := os.OpenFile(filename,os.O_APPEND | os.O_CREATE | os.O_WRONLY,0644) // single pipe for optional
	if err != nil {
		return err
	}
	defer file.Close()
		
	_,err = file.Write([]byte(data))
	if err != nil {
		return err
	}

	return nil
}