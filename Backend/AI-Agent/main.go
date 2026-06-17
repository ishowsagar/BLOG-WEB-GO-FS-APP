package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// entry point for client calling deepseek

// @types declaration

// io.Reader/Writer => these are interfaces which both -> needs a method which intakes ( []byte) and do the desired work based on the context.
// ** io.Reader -> This interface is implemented & satisfied by the method which -> wraps []byte data and constructs a reader source from where data in bytes chunk could be read gracefully
// ** io.Writer -> This interfaces too implemented & needs a method which - intakes []byte data and write to the desired location ( disk,memory (buffer in memory only),http)

// // outbound message type struct
// type OutboundMessage struct {
// 	Role string `json:"role"`
// 	Content string `json:"content"`
// }

// // outbound request payload type struct
// type OutboundPayload struct {
// 	Model string `json:"model"`
// 	Messages []*OutboundMessage  `json:"messages"` //[] of messages type element
// }

// // inbound message type struct
// type InboundMessage struct {
// 	Role string `json:"role"`
// 	Content string `json:"content"`
// }

// // wrapper for inbound messages
// type InboundMessageWrapper struct{
// 	InboundMessage *InboundMessage `json:"message"`
// }

// // inbound message type struct
// type InboundPayload struct {
// 	Choices []*InboundMessageWrapper `json:"choices"`
// }

// ** gemini exclusive
// outbound request payload type struct
type OutboundPayloadGem struct {
	// contents [] -> stores el of type parts in struct -> which is [] stores -> text in struct
	Contents []*ContentsSliceKeyWrapperGem 	`json:"contents"`
}

// to keep conversation history, we have to keep track of all prev chats - req + res based off roles, and sending full for next res
type ContentsSliceKeyWrapperGem struct {
	// ! we have to add explicit role so gemini would know prev context whose cycle belong to whome
	Role string `json:"role"` // ! speaker {user-userReqs,model-Response} 
	Parts []*PartsSliceKeyWrapperGem	`json:"parts"`
}

type PartsSliceKeyWrapperGem struct {
	Text string `json:"text"`
}

type ContentWrapperGem struct {
	Parts []*PartsSliceKeyWrapperGem `json:"parts"`
}

// wrapper for inbound messages
type InboundCandidatesWrapperGem struct{
	ContentWrapperGem *ContentWrapperGem `json:"content"`
}

// inbound message type struct
type InboundPayloadGem struct {
	Candidates []*InboundCandidatesWrapperGem `json:"candidates"`
}


func main() {

	// bug - when this cannot find file by default - must specify path to the env file relative to this file where it is being loaded
	// fix - added relative path to serve it env
	loadErr := godotenv.Load("../.env")
	if loadErr != nil {
		log.Fatalf("failed to load env file, err- %v",loadErr)
	} 

	apiKEY := os.Getenv("GEM_KEY")


	// ** flag **//

	// !Q := why we using flags?
	//& Ans <- flags are just decriptive cmd line arg but very useful that define scope of arg + better control on it
	// based on provided flag - we could attach that context to the prompt to use can directly get desired response quickly
	// for eg. --mode review (review it), doc - document it etc... with q/a, quiz etc...

	// ** flag - defining cmd line arg with better control and flexibilty and human readble 
	// declared flags are invoked with doubledashes -- "decalred Flag key"
	// while defining a flag it-> takes in key under which it'd be defined, a default value, it's description when invoked help on it
	// they must be parsed before getting used -> attached to the application
	// !IMP - must be declared and parsed ( ready to be use in prod ) before manual args are being declared
	Mode := flag.String("mode","review","choose agent response mode for better desired output")
	Deeplearning := flag.Bool("deeplearning",false,"letting agent undergo beast mode with deep intelligence and conversation history")
	flag.Parse()

	// switch on *Mode val for AI meantime response - as it telling to use switch instead
	switch *Mode {
		case "docs" :
			fmt.Println("documentation is in process...")
		case "review" :
			fmt.Println("reviewing code...")
		case "qa" :
			fmt.Println(" started code analysis to ask question based on it...")	
	}//..switch

	// must use pointer * -> to get stored val on that addr
	// fmt.Println("flag - ",*Mode) // if not given address it would print addr, when addrs is given it prints the stored val on that addr

	// now mode is a defined flag, it holds address, so it must be used as address to attach it anywhere we want it to be 
	//! but flags must be used in this order - [0th arg] flags rest args - hut this will break os.args[positional flow and mix up]
	// ** solution - Use flag.Args()[positionAt] -> to set positioned cmd lien args same as what os.Args were doing but keep in mind -> positioned flag args must be defined after 'pasred' non-posi args   
	// used after program arg and before rest 
	// ** end //


	// fmt.Println("key being used:",apiKEY)

	// instead of harcoding it -> we would make this dynamic to let caller provide file name during code run
	
	// os.Args[posiAtCmdLine] -> let us define arg to be added dynamically, this let us set which arg would be defined and used based on position of definition
	// for ex - os.Arg[0] -> during running this file - first provided arg becomes this

	// !IMP -> first arg being at 0th index must be 'program name' like here - `go run ./` here ./ (binary) becomes first arg, 2nd being rest arg to be used 
	// programName := flag.Args()[0] //* ./ -> binary itself becomes it or we say path to the binary
	// fmt.Println(programName)

	// * flag automatically exlcudes binary and non positional from explicit args declaration, so that's why go run {.} {--mode docs} are first parsed flags
	// * then positional arg are comes into play from start 0th position which -> becomes as they are used for

	// flow -> flag.parse -> parses all the defined non-positional flags, so they are used as arg on cmd without taking position
	// then arg defined with -> flag.Args()[positionedAt] takes up the arg place they -> take exact place and provide val to the codebase

	filenameArg := flag.Args()[0]
	writeToFileArg := flag.Args()[1] //* where we wanna write response into which file
	if filenameArg == "" {
		log.Fatalf("could not find required arguement - '%s' that must be specified on '%s' position during execution.",filenameArg,"1st")
	}

	//1. read file content
	cnt,err := ReadFileContent("./",filenameArg)
	if err != nil {
		log.Fatalf("failed to read file content, err- %v",err)
	}



	// fmt.Println("os args",os.Args)

	// ** flow
	// this is the stanadard way of sending request to external api 
	// http.Client - http.Newreq cycle

	// store content as prompt for prompting AI

	// making a external http request with http.NewReq() -> sends client req directly when invoked
	// since here we are client, and req is made to api -> we dont need routern all for handeling -> all done by deepseek internal and sends res, from decoding incoming req to sendign encoded res 
	
	// ! payload is sent via io.Reader -> we need reader type -> as chunk is read in chunks
	
	//2. & client - have to make client who do that req
	client := &http.Client{
		Timeout:  27 * time.Second ,
	}
	
	// we are not using this anymore ->
	// ! since we need to send/recieve types of struct data that satisfies the server inbound types
	// returns an io reader content of type [] of byte wrapped reader -> let reciever read from here in chunks 
	// payloadReader := strings.NewReader(cnt) //* payload reader now wraps content in slice of byte data which let recievers get in chunks
	

	// utility
	// promptedMsg := &OutboundMessage{
	// 	Role: "system",
	// 	Content: "you are a code reviewer, tell me what i did in a beautiful summarised way",
	// }

	// refrenceFileContent := &OutboundMessage {
	// 	Role: "user",
	// 	Content: cnt,
	// }

	var parts []*PartsSliceKeyWrapperGem
	part := &PartsSliceKeyWrapperGem{
		// ! adding conditional mode prompt + content
		Text: BuildPrompt(*Mode,cnt), 
	}
	parts=append(parts, part)

	contents := &ContentsSliceKeyWrapperGem{
		Role: "user", //* while requesting, telling it has user role - user req
		Parts: parts,
	}

	out := &OutboundPayloadGem{
		Contents: []*ContentsSliceKeyWrapperGem{contents},
	}

	
	outByteBuffer,err := json.Marshal(out)
	if err != nil {
		log.Fatalf("failed to encode outgoing payload,err -%v",err)
	}

	// wrapping payload into io reader -> let servers recieves data in chunks
	body := bytes.NewBuffer(outByteBuffer)  // buffer.buffer auto handles read/writes internally

	reqURL :="https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

	// 3. creating request for sending 
	req,err :=http.NewRequest("POST",reqURL,body)
	if err != nil {
		log.Fatalf("failed to request AI,err - %v",err)
	}

	req.Header.Set("Content-Type","application/json") // header key is **case-sensi 
	req.Header.Set("x-goog-api-key", apiKEY) // Pass your Gemini key here


	// before sending req,manage history if enabled
	// ** conversation history with --mode deeplearning **//

	deeplearningEnabled := *Deeplearning == true
	// check if history already exists
	if deeplearningEnabled {
	fmt.Println("unlocking full potential of agent and powered by deep intelligence⚡")

	
	historyFile := "history.json"
	historyExists,_ :=CheckConvoHistoryFILE(historyFile)

	// if history does not exists already
	if !historyExists {
		// create new file
		err =os.WriteFile(historyFile,[]byte("[]"),0644)
		if err != nil {
			return
		}
	}

	// if yes,retrieve historyData in form of []byte -> unmarshal into pqyload -> append new message to it
	
	//* stores contents slice{user,model...}
	var historyGem []*ContentsSliceKeyWrapperGem
	historyData,err := os.ReadFile(historyFile) 
	if err != nil {
		log.Fatal(err.Error())
	}

	err = json.Unmarshal(historyData,&historyGem)
	if err != nil {
		log.Fatal(err.Error())
	}

	// append new user message data to it
	userparts := &PartsSliceKeyWrapperGem{
				Text: BuildPrompt(*Mode,cnt),
	}
	newUserHistory := &ContentsSliceKeyWrapperGem{
		Role: "user",
		Parts: []*PartsSliceKeyWrapperGem{
			userparts,
		},
	}

	// !stores content slice - where contentwrappers are pushes 
	historyGem = append(historyGem,newUserHistory)


	
	reqURL :="https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

	// ! sending full contents history 
	deepOutByteBuffer :=  &OutboundPayloadGem{
		Contents: historyGem,
	}
	var deepBodyBuf bytes.Buffer //storing bytes data
	
	// storing in bytes in buffer
	err = json.NewEncoder(&deepBodyBuf).Encode(deepOutByteBuffer)
	if err != nil {
		log.Fatal("failed to encode history")
	}
	// if history is encoded successfully, send to the gemini for collective response

	// 3. creating request for sending 
	deepReq,err :=http.NewRequest("POST",reqURL,&deepBodyBuf)
	if err != nil {
		log.Fatalf("failed to request AI,err - %v",err)
	}

	deepReq.Header.Set("Content-Type","application/json") // header key is **case-sensi 
	deepReq.Header.Set("x-goog-api-key", apiKEY) // Pass your Gemini key here

	// send to gemini 
	deepRes,err := client.Do(deepReq)
	if err != nil {
		return
	}
	
	// intercept response - get res - write data to the history with role "model" this time
	deepBody := deepRes.Body
	defer deepBody.Close() //deferred call to close reader

	
	// read incoming body data into byte  
	deepResByte,err := io.ReadAll(deepBody)
	if err != nil {
		return
	}

	var deepInbound *InboundPayloadGem
	err = json.Unmarshal(deepResByte,&deepInbound)
	if err != nil {
		return
	}

	
	// parsed response validation check
	if len(deepInbound.Candidates) == 0 {
		log.Fatal("API returned a successful deep response, but the 'Candidates' array was empty.")
	}
	
	if deepInbound.Candidates[0].ContentWrapperGem == nil || len(deepInbound.Candidates[0].ContentWrapperGem.Parts) == 0 {
		log.Fatal("API returned deep candidates, but the 'Parts' array was empty or nil.")
	}
	

	AIResponseContent := deepInbound.Candidates[0].ContentWrapperGem.Parts[0].Text
	if len(AIResponseContent) == 0 {
		log.Fatal("empty deep response from AI, must have hit some unexpected error",)
	}

	// if res is validated and good -> contruct content with gemini response with role attached
	deepAIResponse := deepInbound.Candidates[0].ContentWrapperGem.Parts[0].Text
	
	deepPartsGem := &PartsSliceKeyWrapperGem{
		Text: deepAIResponse, //* attaching deep gemini response to it
	}
	deepContentGem := &ContentsSliceKeyWrapperGem{
		Role: "model", //* attaching role
		Parts: []*PartsSliceKeyWrapperGem{deepPartsGem},
	}

	
	// ! appending res content to the history contents
	// appending this to the history - keeping both histories intact
	historyGem = append(historyGem, deepContentGem)
	
	//* now new history would be added to the history file but we also have to get back now full updated history and write to the file
	
	// ! grabbing full history {content : [usercontents,geminicontents]}
	updatedHistoryBytes,err := json.Marshal(historyGem)
	if err != nil {
		return
	}

	// writing to the both file - {content - write full encoded contents in history} , res to the output file
	err = WriteToFile(writeToFileArg,AIResponseContent)
	if err != nil {
		return
	}
	err = WriteToFile(historyFile,string(updatedHistoryBytes))
	if err != nil {
		return
	}
	fmt.Printf("deep response is recieved in %s file",writeToFileArg)
	
	}else {

	
	// ** end //

	//4. initiates the request
	res,err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to get res, err - %v",err)
	}
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
	err = WriteToFile(writeToFileArg,AIResponseContent)
	if err != nil {
		log.Fatalf("failed to write to the file, err - %v",err)
	}

	// switch on *Mode val for response - as it telling to use switch instead
	switch *Mode {
		case "docs" :
			fmt.Println("Documentation Success⚡")
		case "review" :
			fmt.Println("analysed Success⚡")
		case "qa" :
			fmt.Println("questions Success⚡")	
	}//..switch
	}//..else
}

