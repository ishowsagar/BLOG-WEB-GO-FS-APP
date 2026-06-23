package main

import (
	"bufio"
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
	"github.com/joho/godotenv"
)

// entry point for client calling deepseek

// @types declaration
// type struct for storing agent needed data
type AgentConfig struct {
	Mode               string // agent prompt mode selection
	Deeplearning       bool   // deeplearing bool for turining on memory or not
	Target             string // TargetDirEntry - for selection dir mode, if given dir -> executes dir entries and prompt based off mode + dir entries
	Git                string // git selected mode for working with git automations - works with non dir and supports review,status,commit
	FilenameArg        string // file for context prompt in normal mode
	writeToFileArg     string //  output file name
	selectedDirPathArg string // dir path for scan context
	dirSubFileTypesArg string // dir sub entries files selection for scan
}

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
	Contents []*ContentsSliceKeyWrapperGem `json:"contents"`
}

// to keep conversation history, we have to keep track of all prev chats - req + res based off roles, and sending full for next res
type ContentsSliceKeyWrapperGem struct {
	// ! we have to add explicit role so gemini would know prev context whose cycle belong to whome
	Role  string                     `json:"role"` // ! speaker {user-userReqs,model-Response}
	Parts []*PartsSliceKeyWrapperGem `json:"parts"`
}

type PartsSliceKeyWrapperGem struct {
	Text string `json:"text"`
}

type ContentWrapperGem struct {
	Parts []*PartsSliceKeyWrapperGem `json:"parts"`
}

// wrapper for inbound messages
type InboundCandidatesWrapperGem struct {
	ContentWrapperGem *ContentWrapperGem `json:"content"`
}

// inbound message type struct
type InboundPayloadGem struct {
	Candidates []*InboundCandidatesWrapperGem `json:"candidates"`
}

// type with prefix data
type InboundChunkPayloadGemWrapper struct {
	Data *InboundPayloadGem `json:"data"`
}

func main() {

	// inside main, we need to execute that fnc -> which watch out for events <- recieved in its watcher chan from the os
	// need a way to integate,it invokes on that dir and do the work and also need a way to trigger agent from inside the watcher

	// writer testing

	// ohhhhh..os.stdout satisfies writer interface and it wrote tp console dest...-> this is how fmt working under the hood, piping provided bytes to the console
	// n,err := BytesWriter(os.Stdout,[]byte("Hello terminal\n"))
	// if err != nil {
	// 	slog.Error("failed to write bytes","error",err)
	// 	return
	// }

	// reading collected file chunk
	// content,err := ReadFileContent(".","main.go")
	// if err != nil {
	// 	return
	// }

	// // wrapping data into bytes and creating a reader source - from where bytes can be read
	// reader := strings.NewReader("hello im here can you see me") //\forward slash n for line breaks - but not needed when streaming via scanner stripped bytes

	// // wraps the reader source in a way that -> it reads from source in chunks
	// scanner := bufio.NewScanner(reader)
	// // var chunkSizeAccumulator int // unidiomatic way, define empty var with var declarement with :=
	// chunkSizeAccumulator := 0
	// // * scan gets the bytes -> looping to get all chunks bytes from reader in loop -> loop make it possible for read and write chunks -> live write
	// for scanner.Scan() {
	// 	//! scan gives each of chunk bytes data -
	// 	eachLine := scanner.Text() //returns string from read chunk from reader
	// 	n,err := StringsWriter(os.Stdout,eachLine)
	// 	if err != nil {
	// 		// todo - later add own logger by writer which writes err to the console
	// 		slog.Error("failed to write bytes stream","error",err)
	// 		return
	// 	}

	// 	// on successfull write -> accumulate chunk written for total
	// 	chunkSizeAccumulator += n
	// 	time.Sleep(time.Millisecond * 250 )
	// }
	// slog.Info("successfully wrote streamed content to the respective destination","bytesWritten",chunkSizeAccumulator)
	// // successfully working ✅ as intended it to work -> it is reading chunks wrapped in scanner and writing each stripped line to the writer dest✅
	// os.Exit(0)

	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	)	

	// spinner - add start and stop with pre suffix for usint it
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	// bug - when this cannot find file by default - must specify path to the env file relative to this file where it is being loaded
	// fix - added relative path to serve it env
	loadErr := godotenv.Load("../.env")
	if loadErr != nil {
		log.Fatalf("failed to load env file, err- %v", loadErr)
	}

	// apiKEY := os.Getenv("GEM_KEY")
	apiKEY := os.Getenv("GEMINI_API_KEY")

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
	// Mode := flag.String("mode", "review", "choose agent response mode for better desired output")
	// Deeplearning := flag.Bool("deeplearning", false, "letting agent undergo beast mode with deep intelligence and conversation history")
	// Target := flag.String("type", "file", "choose agent either reads content from file or directory")
	// Git := flag.String("git", "status", "choose what agent to do with respect to git automations") //* for capturing git command from user
	// flag.Parse()

	// switch on *Mode val for AI meantime response - as it telling to use switch instead
	// switch *Mode {
	// case "docs":
	// 	fmt.Println("documentation is in process...")
	// case "review":
	// 	fmt.Println("reviewing code...")
	// case "qa":
	// 	fmt.Println(" started code analysis to ask question based on it...")
	// } //..switch

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

	// args validation to prevent crash
	// declaredArgs := flag.Args() // only gives us provided args

	//! upgrade -  Switching to user prompted selections and replacing them from args
	// ** Reason -> imagine asking user to provide all the things upfront when they all needed is to run program first and then provide needy things when prompted
	// ** this makes it robust and better user experience and better domain expansion and showing purpose of each arg

	// * arg replacer - user input { process comes with stdin <- pulling bytes from there }
	// now all these variables would be filled by user input

	// get these from proccess's stdin stream capturing user input bytes

	var hasDeeplearningModeSelected bool

	// & user input access
	s.Suffix = "Booting up agent✨..."
	s.Color("cyan")
	s.Start()
	time.Sleep(time.Second * 2)

	s.Stop() // stop after 2 sec delay
	fmt.Println("---- Agent booted up successfully 🚀 -----")

	// user provides input <-= captured by stdin from terminal -> reader streaming to process
	modeSelection := "Select Agent Response Mode - [review,docs,qa] : "
	selectedMode := CaptureInputFromShell(modeSelection)

	DeeplearningSelection := "Do you want to learn with memory intelligence (premium✨) - [y,n] : " //turn into bool based off ans
	selectedDeeplearningMode := CaptureInputFromShell(DeeplearningSelection)

	// returns true if var's val matches provided value
	hasDeeplearningModeSelected = strings.EqualFold(selectedDeeplearningMode, "y")

	TargetSelection := "Select what agent could access- [file,dir] for scans : "
	selectedTarget := CaptureInputFromShell(TargetSelection)

	// todo - add if target is file -> ask for file to scan for + validations for all fields later
	// todo 2 - dynamically ask for input based off prior inputs

	var selectedFile string
	if TargetSelection == "file" {
		fileSelection := "Provide file name : "
		selectedFile = CaptureInputFromShell(fileSelection)
	}

	// bug => need guard for git - it fires git mode cause we selects none or ""
	// fix - asks for git operations mode only if prior user wants git mode -> also don't prompt if dir mode is selected

	gitAutomationSelection := "Do you want git automations - [y/n] : "
	selectedGitAutomation := CaptureInputFromShell(gitAutomationSelection)

	// tracks selected modes
	var selectedGitOperation string //tracks git modes
	var hasGitEnabled bool          // true if git enabled

	// if user wants git mode
	if selectedGitAutomation == "Y" || selectedGitAutomation == "y" {
		GitOperationSelection := "Select what agent git operation do automatically - [Status,Commit,Review,None] : "
		selectedGitOperation = CaptureInputFromShell(GitOperationSelection)
		hasGitEnabled = strings.EqualFold(selectedGitAutomation, "y")
	} else {
		selectedGitOperation = "none"
	}

	outputFileSelection := "Enter filename where agent write response to (auto-created if not exists already) : "
	selectedOutputFile := CaptureInputFromShell(outputFileSelection)

	dirPathSelection := "Provide dir path to scan for (relative path only 🚨) : "
	selectedDirPath := CaptureInputFromShell(dirPathSelection)

	subfileTypeSelection := "Select which files should be focusedly scanned (premium✨) : "
	selectedSubFileType := CaptureInputFromShell(subfileTypeSelection)

	s.Suffix = "gathering user selections✨..."
	s.Color("cyan")
	s.Start()
	time.Sleep(time.Second * 2)

	s.Stop() // stop after 2 sec delay
	agentConfig := &AgentConfig{
		Mode:               selectedMode,
		Deeplearning:       hasDeeplearningModeSelected,
		Target:             selectedTarget,
		FilenameArg:        selectedFile,
		writeToFileArg:     selectedOutputFile,
		Git:                selectedGitOperation,
		dirSubFileTypesArg: selectedSubFileType,
		selectedDirPathArg: selectedDirPath,
	}
	fmt.Printf(`User selections :
		mode : %s,
		deeplearning? : %v,
		target : %s,
		git : %s,
		dirPath : %s,
		subFileType : %s,
		outputFile :%s,
		filenameArg :%s,
	`, agentConfig.Mode, agentConfig.Deeplearning, agentConfig.Target, agentConfig.Git, agentConfig.selectedDirPathArg, agentConfig.dirSubFileTypesArg, agentConfig.writeToFileArg, agentConfig.FilenameArg)

	confirmation := "Do you want to proceed with this ? (y/n) :"
	// adding confirmation so user can rollback or commit input
	userConfirmation := CaptureInputFromShell(confirmation)
	proceed := userConfirmation == "Y" || userConfirmation == "Yes" || userConfirmation == "YES" || userConfirmation == "YeS" || userConfirmation == "y"

	// if any of condition is matched -> proceed to do work -> invoke agent otherwise rollback to empty state and restart the operation
	// todo - might need robust way to handle this better
	if proceed {
		s.Suffix = "agent thinking..."
		s.Color("red")
		s.Start()
		time.Sleep(time.Second * 4)
		s.Stop()
		// rest logic upon retrieval of user input
	} else {
		s.Suffix = "agent rolling back changes..."
		s.Color("cyan")
		s.Start()
		time.Sleep(time.Second * 4)
		s.Stop()
		agentConfig = nil
		fmt.Println("operation aborted gracefully🛑")
		return
	}

	// os.Exit(0)

	// fmt.Println(declaredArgs)
	// var argCount int
	// if len(declaredArgs) < 4 {

	// 	// implementing by reading about streams attached to proccesses, piping and os

	// 	for i,currentArg := range declaredArgs {
	// 		argCount += i
	// 		fmtdArg := fmt.Sprintf("found '%s' arg at '%d' position",currentArg,i)
	// 		fmt.Println(fmtdArg)
	// 	}
	// first finds what args are not passed
	// capture them progessively with prompted message -> capture inputs and store them

	// integrate them into the program
	// ** we would capture args as stdin piped streamed to program,as talker to user
	// todo - later make this robust when spinned up => prompts for every arg with detailed msg

	// fmt.Println("args count",argCount)
	// os.Exit(0)
	// bug - always access after validation
	// fix - as they are already declared, they are being just accessed here

	// ** Upragde - agrs have been upgraded into user input based selection
	// FilenameArg := flag.Args()[0]        //* tracks which file is being selected for reading context from single file only for -> deep + normal <- "none" dir mode
	// writeToFileArg := flag.Args()[1]     //* where we wanna write response into which file
	// selectedDirPathArg := flag.Args()[2] //* giving context which dir to read form <- dymamically giving info
	// dirSubFileTypesArg := flag.Args()[3] //* which specific types to ask for in the selected dir
	// if FilenameArg == "" {
	// 	log.Fatalf("could not find required arguement - '%s' that must be specified on '%s' position during execution.", FilenameArg, "1st")
	// }

	// & universal utility
	// ** flow
	// this is the stanadard way of sending request to external api
	// http.Client - http.Newreq cycle

	// store content as prompt for prompting AI

	// making a external http request with http.NewReq() -> sends client req directly when invoked
	// since here we are client, and req is made to api -> we dont need routern all for handeling -> all done by deepseek internal and sends res, from decoding incoming req to sendign encoded res

	// ! payload is sent via io.Reader -> we need reader type -> as chunk is read in chunks

	//2. & client - have to make client who do that req
	client := &http.Client{
		Timeout: 1 * time.Minute,
	}
	reqURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"
	chunkReqURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:streamGenerateContent?alt=sse"

	// testing changes if it shows this line being written to the main.go

	// todo - add mode enable selectors in one place later

	// ** Git automations **//
	allowedGitModes := map[string]bool{
		"status": true,
		"diff":   true,
		"commit": true,
		// ! risky add push later
	}

	//! client selected git mode early validation - only allow available modes only

	clientSelectedGitMode := agentConfig.Git

	// todo - make mode more robust and clear by refactoring into selection first and then firing operations based off that

	// git mode is enabled when this non-position parsed flag is non-empty cause if not specified -> becomes empty
	gitEnabled := hasGitEnabled && agentConfig.Target != "dir" && agentConfig.FilenameArg != "none" && selectedGitOperation != "none"

	// when these conditons met => git enabled cause we don't want other cases to be fired regardlessly
	if gitEnabled {

		// bug - early returns without mode injected return could cause program to crash on any mode when git flag not passed
		// fix - render early return only inside git mode
		// comma,ok method to check if this key exists in the map
		_, available := allowedGitModes[clientSelectedGitMode]
		if !available {
			slog.Error("this mode is avaiable in premium subscription only💲", "status", "unauthorized🛑")
			return
		}
		// * now client would only prompt git if selected mode is available
		// slog.Info("git mode has been selected⚡","mode",clientSelectedGitMode)

		// 1. run git operator by passing client selected command

		// todo - later we get it from client to inject here

		// todo - before commiting need a git operator operation to -> add desired files first { later user picked only }

		// 1. add files - later implement which files to add for commit

		// get all files which user wanna commit
		repo, err := os.ReadDir(agentConfig.selectedDirPathArg)
		if err != nil {
			slog.Error("failed to read repo", "error", err)
			return
		}

		var trackStagefiles []string // holds required staging files - added names of t
		for _, repoFile := range repo {

			// only read file,skip dits
			if repoFile.IsDir() != true {
				trackStagefiles = append(trackStagefiles, repoFile.Name())
			} else {
				// else is it is dir -> skip it
				continue
			}
		}

		// **once it collects all the required staging files, provide to the add git operator
		// ! Imp - need a way to pass whole slice elements as files instead of running git adder on each - flatten out slice of dirEntry files

		// bug - was constructing slice, but had to append fileName slice to append them into one so can be vardiaced
		// fix - append filesSlice as vardiac -> so it could append it to the slice and later when again vardiaced -> pass it as all el instead of one
		addfilesArgs := append([]string{"add", "-f"}, trackStagefiles...) // appending whole arr
		// ! imp -> since args are vardiac -> decoded as flattened out already by ..., so if we pass it as whole it will do the job

		_, err = RunGitOperator(addfilesArgs...)
		if err != nil {
			slog.Error("failed to run git operator", "cmd", "add", "error", err)
			return
		}

		// bug - git add does not send anything or even send useful information - need to send something that makes sense what is staged ( added ) and ready to be commit - full context window of the work
		// fix - call git diff -> send that as context for getting the commit msg
		diffFileRead, err := RunGitOperator("diff", "--staged")
		if err != nil {
			slog.Error("failed to run git operator", "cmd", "diff", "error", err)
			return
		}

		diffFileReadBytes, err := io.ReadAll(diffFileRead)
		if err != nil {
			msg := "failed to read added file output from reader"
			slog.Error(msg, "error", err)
			return
		}

		// 2. send files git ouput to get response - get commit message auto generated { might need prompt work}
		commitMsg, err := GetGitResponse(clientSelectedGitMode, string(diffFileReadBytes), reqURL, apiKEY, agentConfig.writeToFileArg, s, client)
		if err != nil {
			msg := "failed to get commit message"
			slog.Error(msg, "error", err)
			return
		}

		// 3. get commit msg and display to user and ask for confirmation
		fmt.Printf("Agent generated commit message ✨ :%s", commitMsg)
		confirmationMsg := "Commit with this message? (Yes/No) :"
		ConfirmationRes := CaptureInputFromShell(confirmationMsg)

		// recieved input validation - only accepts yes or no - otheriwse default to no
		allowedConfirmationsOnly := map[string]bool{
			"Y":   true,
			"N":   true,
			"Yes": true,
			"No":  true,
		}

		// checking if res exists in allowedConfirmations
		_, ok := allowedConfirmationsOnly[ConfirmationRes]
		if ConfirmationRes == "" || !ok {
			slog.Error("empty or unknown response from client, it must be one of them [Y,N,Yes,No]")
			return
		}

		// 4.  if incoming is validated and yes,do commit
		switch ConfirmationRes {
		//Agent spinner notifies it is doing its work
		case "Yes", "Y":
			// commit
			// fixings => not all git commands works on single command, like for commit had to add '-m' flag explicitly
			// now it runs smoothly commit command
			buffReader, err := RunGitOperator(clientSelectedGitMode, "-m", commitMsg) //* now pass as much as args git executable commands expects
			if err != nil {
				slog.Error("failed to run git operator", "error", err)
				return
			}

			// 2. returns ioreader { source of bytes from where respective methods can read bytes from} <- read ouput
			gitOutByte, err := io.ReadAll(buffReader)
			if err != nil {
				slog.Error("failed to read git output from bytes.Buffer reader", "error", err)
				return
			}

			// 3. build content to send to gem
			gitOut := string(gitOutByte)
			_, err = GetGitResponse(clientSelectedGitMode, gitOut, reqURL, apiKEY, agentConfig.writeToFileArg, s, client)
			if err != nil {
				slog.Error("failed to get commit details", "error", err)
				return
			}

		// otherwise tear down the operation
		case "N", "No":
			// todo unadd added files dynamically later
			s.Suffix = "Agent rolling back changes..."
			s.Start()

			rmRead, err := RunGitOperator("reset", "Head")
			if err != nil {
				slog.Error("failed to unstage all changes", "error", err)
				return
			}
			rmByteRes, err := io.ReadAll(rmRead)
			if err != nil {
				slog.Error("failed to read bytes from reader", "error", err)
				return
			}

			// rolling back changes - no need to invoke ai - do custom res
			unstagedChanges := string(rmByteRes)

			err = WriteToFile(agentConfig.writeToFileArg, unstagedChanges)
			if err != nil {
				slog.Error("failed to write res to the file", "error", err)
				return
			}
			s.Stop()
			fmt.Print("All your changes have been rolled back to original previous state✨")

		} //..switch
	}

	// ** end **//

	// ** DIR READ - dir files ocntent read and sent to gemini **//

	// no deep learning on dir mode for now - > token window would be bigger with bigger payload head each time ( for now)

	// if these condition meet -> then only proccess the dir work <- intentionally doing now so first make it work -> then dynamic + robust later

	// when Target is "dir" -> we need to get context of which dir + file type inside the dir
	dirSelection := agentConfig.Target == "dir" &&
		agentConfig.Deeplearning == false &&
		agentConfig.selectedDirPathArg == "." &&
		agentConfig.dirSubFileTypesArg == ".go" //! go for now , for other have to add explicit validtion,so early temp return check here

	// if target dir is selected and context of dirPath ( where to look in files ) and files type context is given -> then only do the dir operation
	if dirSelection {
		// slog.Info("entered dir mode...")
		// checking dir mechanism and related branches
		// slog.Info("dir mode is selected","selectedDir",selectedDirPathArg,"specificFileSelection",dirSubFileTypesArg)
		// 1. check if provided DirPath points to the dir

		// ** invoke spinner to tell client dir is being scanned...

		s.Suffix = " Agent is scanning and understanding directory structures..." // Adds text next to the spinner
		s.Color("red")                                                            // Optional: give it some color!

		// 2. Start the spinner
		s.Start()
		var dirCount int
		var nestedDirEntriesCount int
		info, err := os.Stat(agentConfig.selectedDirPathArg)
		if err != nil {
			return
		}
		if info.IsDir() == true {
			dirCount += 1
			dir, err := os.ReadDir(agentConfig.selectedDirPathArg)
			if err != nil {
				return
			}

			for _, entry := range dir {
				if entry.IsDir() == true {
					dirCount += 1
					nestedDir, err := os.ReadDir(entry.Name())
					if err != nil {
						return
					}
					for _, entry := range nestedDir {
						nestedDirEntriesCount++
						fmt.Printf("nestedDir file is found!,filename - %s", entry.Name())
					}
				}
			}
		}
		s.Stop() //stop spin as now it would have scanned it

		fmt.Println("📂Dir count - ", dirCount)
		fmt.Println("📌Nested dir's entries count - ", nestedDirEntriesCount)

		if nestedDirEntriesCount > 0 {
			fmt.Println("nested dir file count", nestedDirEntriesCount)
		}

		fileInfo, err := os.Stat(agentConfig.selectedDirPathArg)
		if err != nil && os.IsNotExist(err) {
			slog.Debug("file/dir does not exists", "error", err)
			return
		}

		slog.Info("File metadata", "isDir?", fileInfo.IsDir(), "name", fileInfo.Name())

		isdir := fileInfo.IsDir() == true //* if this conditions meets -> it is dir,true cause it could return false too that for file but we dont want that
		// var dirSubFilesEntriesCount int
		// var goFilesCount int
		// var bufReader bytes.Buffer //* store byte data to it
		var goFilesDataAccumulator string

		// 2, if provided path points to the dir, access files of the dir using os.ReadDir()
		if isdir {

			// Reader/writer interface => all they need is a 'source', a source which -> from where it can read bytes data from, as bytes data exists in []byte form
			// that's why -> they need a source which knows how to send []of byte data <- from where it can read bytes from...that's why when io.reader is wanted ->
			// => all it wants -> a source ( any source it could be http(body) - network, stringReader which wraps []bytes and becomes a io.reader source,same for buffer.Buffer
			// that's why it don't care about ioreader source as it is serving data, it could be served from anywhere-> cause that func only needs chunks bytes data to work with but does not care where it is coming from.

			slog.Info("directory found & has started analysis...")

			var accumulator strings.Builder //strings concatenator

			// spinner - call start and stop with declared suffix => use it where needed ,_ just start from code and stop when done
			s.Suffix = "Reading parent dir..." //todo - might use formatted string for dynamic loggin
			s.Color("cyan")
			s.Start()
			// if it is confirmed it is a director -> reading files of that dir via os.ReadDir (needs dir path as name for locating it)
			parentDir, err := os.ReadDir(agentConfig.selectedDirPathArg)
			if err != nil {
				slog.Error("failed to read dir entries", "error", err)
				return
			}

			// dir scan flow becomes -
			// 1.check if current path/name points to dir , get its info (os.stat)
			// 2. if info tells it is a dir -> readDir -> range & loop over it -> acccumulate content
			// 3. if info tells current dirEntry is a dir -> again -> readDir -> range & loop over it -> acccumulate content
			// 4. call a.String to get final string
			// 5. pass full content to gemini for prompting and getting response
			// 6. but nested dir got again nestedDir -> stop it there

			slog.Info("entering into dir...")
			s.Stop()

			for _, currentDirEntry := range parentDir { // first always is index,thenDataPiece
				// gives us []*DirEntry which -> stores in that dir

				// & if current Entry is not a dir -> read file content & accumulate
				// Dir entries could be nested dir or normal files -> so checking if it is not dir -> do same
				if currentDirEntry.IsDir() != true {
					// * checking if this current iteration (as it would be file's whose file has this suffix) -> if yes then only proceed

					// bug - 503 cause we were sending unneccassary data
					// 4. Read each file content with specifics file type selected
					if !strings.HasSuffix(currentDirEntry.Name(), agentConfig.dirSubFileTypesArg) || currentDirEntry.Name() == ".git" || currentDirEntry.Name() == "node_modules" {
						continue //* we just need to skip current iteration, continue to the next iteration, don't execute code furthur
					}
					// yes if it has suffix,it won't skip and run this block of code as upper condition is ignored as it has suffix
					eachGoFileContent, err := ReadFileContent(agentConfig.selectedDirPathArg, currentDirEntry.Name())
					if err != nil {
						slog.Error("failed to read file content", "error", err)
						return
					}

					// accumlate each chunk of string - use strings.builder for better strings concatenation
					accumulator.WriteString(fmt.Sprintf("---- file starts from here, name ---%s\n", currentDirEntry.Name()))
					accumulator.Write([]byte(eachGoFileContent))
					accumulator.WriteString(fmt.Sprintf("\n---- file ends here, name ---%s\n", currentDirEntry.Name()))
				} //file

				
				// & if currentDirEntry is literally a dir

				s.Suffix = "Scanning nested dir..." //todo - might use formatted string for dynamic loggin
				s.Color("cyan")
				s.Start()
				if currentDirEntry.IsDir() == true {
					// read and loop over dir again and get each file content accumulated
					nestedDir, err := os.ReadDir(currentDirEntry.Name())
					if err != nil {
						slog.Error("failed to read dir entries", "error", err)
						return
					}

					slog.Info("found nested dir 🚨", "dirPath", currentDirEntry.Name(), "type", currentDirEntry.Type())
					for _, nestedDirEntry := range nestedDir {
						// reading content of each again

						// & checking again if it a file only
						if nestedDirEntry.IsDir() != true {

							// todo - limiting to don't go furthur <- might make it premium feature later as it eats more credits
							nestedDirEntryFileContent, err := ReadFileContent(currentDirEntry.Name(), nestedDirEntry.Name())
							if err != nil {
								slog.Error("failed to read nested dir file's content", "error", err)
								return
							}

							// if it has successfully got the content -> add to the accumulator
							dirOpener := fmt.Sprintf("---- Nested dir -%s files starts from here, name ---%s\n", currentDirEntry.Name(), nestedDirEntry.Name())
							dirCloser := fmt.Sprintf("---- Nested dir -%s files ends here, name ---%s\n", currentDirEntry.Name(), nestedDirEntry.Name())
							accumulator.WriteString(dirOpener)
							accumulator.Write([]byte(nestedDirEntryFileContent))
							accumulator.WriteString(dirCloser)
						} //..nested-file

						if nestedDirEntry.IsDir() == true {
							slog.Info("found another nested dir inside nested dir", "name", nestedDirEntry.Name())
							slog.Info("skipping it 🚨")
							continue //skip
						}

					} //..range
				} //..nested-dir
				s.Stop()

				// goFilesCount++ //track count
				// slog.Info("found go file","name",currentDirEntry.Name())
				// dirSubFilesEntriesCount ++
				// slog.Info("successfully accessed dir sub file","currentFileName",currentDirEntry.Name())

			} //..parent-dir
			goFilesDataAccumulator = accumulator.String() //* gives final string
		}

		// * we need to store accumulated data somewhere first to query it once
		// clientData,err := json.Marshal(goFilesDataAccumulator)
		// if err != nil {
		// 	slog.Error("failed to encode given client data","error",err)
		// 	return
		// }

		// ouptutFile := "go_content.md"
		// ioreader := strings.NewReader(goFilesDataAccumulator) //wraps strings data in []byte and satisfies ioreader -> source which gives []byte data

		// // it needs a ioreader source - from where it can read bytes
		// clientDataBytes,err := io.ReadAll(ioreader)
		// if err != nil {
		// 	slog.Error("failed to read client data bytes into reader","error",err)
		// 	return
		// }

		// err = os.WriteFile(ouptutFile,clientDataBytes,0644)
		// if err != nil {
		// 	slog.Error("failed to write client data","error",err)
		// 	return
		// }

		// now at this point we have both type of data -> bytes + string

		s.Suffix = "Agent is reviewing all files..." //todo - might use formatted string for dynamic loggin
		s.Color("cyan")
		s.Start()
		time.Sleep(time.Second * 2)
		s.Stop()

		//5. create req with http.NewReq() - need to send data of outboundPayloadGem only
		dirParts := &PartsSliceKeyWrapperGem{
			Text: BuildDirPrompt(agentConfig.Mode, goFilesDataAccumulator),
		}
		dirContents := &ContentsSliceKeyWrapperGem{
			Role:  "user",
			Parts: []*PartsSliceKeyWrapperGem{dirParts},
		}
		dirOutbound := &OutboundPayloadGem{
			Contents: []*ContentsSliceKeyWrapperGem{dirContents},
		}

		out, err := json.Marshal(dirOutbound)
		if err != nil {
			slog.Error("failed to marshal client data", "error", err)
			return
		}

		// ! guard check for payload =:
		totaOutBytesLength := len(goFilesDataAccumulator)

		fmt.Println("payload size", totaOutBytesLength)

		s.Suffix = "Almost there, thanks for waiting💪..."
		s.Color("red")
		s.Start()
		time.Sleep(time.Second * 2)
		s.Stop()
		// readerBuf act as both reader and writer -> used as reader as bytes read into buffer and now it serves bytes data as source reader
		bufRead := bytes.NewBuffer(out)

		// this method wants a ioreader -> as learned just means -> source from it can read bytes data
		dirReadReq, err := http.NewRequest("POST", chunkReqURL, bufRead)
		if err != nil {
			slog.Error("failed to request client dir request", "error", err)
			return
		}

		// header set for api access
		dirReadReq.Header.Set("Content-Type", "application/json")
		dirReadReq.Header.Set("x-goog-api-key", apiKEY)

		// 🛡️ Explicitly notify proxies not to buffer this payload
		// dirReadReq.Header.Set("Accept", "text/event-stream")
		// dirReadReq.Header.Set("Connection", "keep-alive")

		// 1. Initialize the spinner
		// CharSets[14] is a cool dot-bouncing animation, but there are dozens of options
		// s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		// s.Suffix = "Agent is about to finish thinking..." // Adds text next to the spinner
		// s.Color("cyan")                                   // Optional: give it some color!

		// 2. Start the spinner
		// s.Start()
		// 6. do the requesr
		dirResp, err := client.Do(dirReadReq)
		if err != nil {
			slog.Error("failed to send client dir request", "error", err)
			return
		}

		// 4. Stop the spinner immediately after the request finishes!
		// s.Stop()

		// api validation if it has returning err as status code is not 200 means not okay - don't execute furthur - applied in all same way
		if dirResp.StatusCode != 200 {
			bodyBytes, _ := io.ReadAll(dirResp.Body)
			log.Fatalf("API returned a non-200 status code: %d, response: %s", dirResp.StatusCode, string(bodyBytes))
		}

		// now when chunked response arrives it arrives in chunks body with data prefix <- must unload into this
		dirBody := dirResp.Body //* body carries res <- reader
		defer dirBody.Close()

		// 7. retrieve resp + validate + decode + get response -> write
		// incoming res comes in bytes

		// s.Suffix = "gathering response..." //todo - might use formatted string for dynamic loggin
		// s.Color("cyan")
		// s.Start()

		// bug - this is useless and breaker as now we are not reading untill its finished in all in once but in chunks
		// fix - remove this entirely since body is also a reader source -: scanner needs a source to read bytes from reader in chunks and scans each chunks progessively
		// read incoming body data into byte
		// dirBodyBytes, err := io.ReadAll(dirBody) // read all bytes data
		// if err != nil {
		// 	return
		// }

		// break into gem inbound res
		// var dirInbound *InboundPayloadGem
		// err = json.Unmarshal(dirBodyBytes, &dirInbound)
		// if err != nil {
		// 	return
		// }

		// extract ai response from inbound after validation
		// parsed response validation check
		// if len(dirInbound.Candidates) == 0 {
		// 	log.Fatal("API returned a successful deep response, but the 'Candidates' array was empty.")
		// }

		// if dirInbound.Candidates[0].ContentWrapperGem == nil || len(dirInbound.Candidates[0].ContentWrapperGem.Parts) == 0 {
		// 	log.Fatal("API returned deep candidates, but the 'Parts' array was empty or nil.")
		// }

		// s.Stop()

		// s.Suffix = "Almost done..." //todo - might use formatted string for dynamic loggin
		// s.Color("green")
		// s.Start()
		// 7. get response content
		// AIDirResponseContent := dirInbound.Candidates[0].ContentWrapperGem.Parts[0].Text
		// if len(AIDirResponseContent) == 0 {
		// 	log.Fatal("empty deep response from AI, must have hit some unexpected error")
		// }

		// todo - remove later - add same source for stream write to file too, keeping to keep flow intact

		// bug - this is no longer valid as this does not comes as collected final res but chunked body reader which has each chunk bytes coming untill not finished with gemini
		// fix - remove this and intercept body read into scanner -> so scanner collects each chunk res directly and writes to the respective dest
		// if res is validated and good -> contruct content with gemini response with role attached
		// dirAIResponse := dirInbound.Candidates[0].ContentWrapperGem.Parts[0].Text

		writtenBytesAccumulator := 0

		// ** Wrapping content into scanner -> so writer only writes chunks read from reader src
		// todos new - implement writer to write thst to anywhere - for now console + file only
		//1. wrap content into scanner - cause scanner reads chunks from source in loop

		// bug - it is still writing instant responses as res is already got and full -> need to render body into buf directly
		// fix - read into scanner directly from res body --> get chunks from body stream directly
		// buf := bytes.NewBuffer([]byte(dirAIResponse))
		scanner := bufio.NewScanner(dirBody)

		// opening file ( created on fly if not existed early on , appends content on the fly )
		file, err := os.OpenFile(agentConfig.writeToFileArg, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			slog.Error("failed to open file", "error", err)
			return
		}
		defer file.Close()

		//scanner cap bottleneck fixing
		maxCap := 1024 * 1024
		buf := make([]byte,0,64 * 1024)
		scanner.Buffer(buf,maxCap)

		//2. loop over scanner.scan which gives -> read chunk from reader source which
		for scanner.Scan() { // returns true if successfully read chunk from reader source

			// fix - open once before it starts, and when loop ends then only close file after err is checked

			//3.-> gives access to each chunk string with string method
			// eachline := scanner.Text()

			// get chunked bytes from body
			var chunkInboundGem *InboundPayloadGem
			eachbyte := scanner.Bytes()

			// bug - incoming data comes with 'data' prefix we have to remove it
			// fix - by remvoing prefix we can now unload into inboundgem wrapper directly
			line := bytes.TrimPrefix(eachbyte, []byte("data: ")) // sse comes with trailing space after colon

			// bug - sse often comes with notifiers payload like gemini sends data:[done] <- need to validate that incoming is just valid payload for unmarshaling
			// fix - validate it


			if len(eachbyte) == 0 || bytes.Equal(eachbyte, []byte("[DONE]")) {
				continue 
			}
			if string(line) == "" || string(line) == "[DONE]" {
				continue //skip this iteration in the scanner which <- reads from body source directly in chunks
				// scanner reads chunk from reader in chunks and scans them out to get each chunk in loop
			}

			// validate and decode chunk
			err = json.Unmarshal([]byte(line), &chunkInboundGem)
			if err != nil {
				slog.Error("failed to decode incoming bytes data", "error", err)
				return
			}

			if len(chunkInboundGem.Candidates) > 0{
			// iterating candidates
			for _,candidate := range chunkInboundGem.Candidates {
				if candidate.ContentWrapperGem == nil {
					continue 
				}

				// bug - removing hardcoded layer of parts, it might come in any other than at 0th posi
				// fix - iterate over parts to get the not nil part's text
				// if current candidate is not nil -> safely get them
				for _,part := range candidate.ContentWrapperGem.Parts {
					// but check first if current part has empty text or not
					if part.Text == "" {
						continue
					}

					// if text is not empty and successfully retrieved from iteratted current part 
					chunkStrippedText := part.Text
					
					// write to the console and file later
					// ! do not accumulate -> write progressively
					n, err := StringsWriter(os.Stdout, chunkStrippedText)
					if err != nil {
						slog.Error("failed to write response", "error", err)
						return
					}
					writtenBytesAccumulator += n
					
					// 8.write response to the file - of this chunked text
					_, err = file.Write([]byte(chunkStrippedText))
					if err != nil {
						slog.Error("failed to write client data to the file", "error", err)
						return
					}
					
					} 
					}
			}//range-candidate
		} // if
				
		// 		if  chunkInboundGem.Candidates[0].ContentWrapperGem != nil && len(chunkInboundGem.Candidates[0].ContentWrapperGem.Parts) > 0 {

		// 		chunkStrippedText := chunkInboundGem.Candidates[0].ContentWrapperGem.Parts[0].Text
					
		// 		// write to the console and file later
		// 		// ! do not accumulate -> write progressively
		// 		n, err := StringsWriter(os.Stdout, chunkStrippedText)
		// 		if err != nil {
		// 			slog.Error("failed to write response", "error", err)
		// 			return
		// 		}
		// 		writtenBytesAccumulator += n
				
		// 		// 8.write response to the file - of this chunked text
		// 		_, err = file.Write([]byte(chunkStrippedText))
		// 		if err != nil {
		// 			slog.Error("failed to write client data to the file", "error", err)
		// 			return
		// 		}
				
		// 		} 
		// 		s.Stop()
		// 	}

		// }

		//! need to if loop returns early cause no panic if nothing is left to read more
		// true -> when non-eof err but that means it returns false on eof err <- eof means end of file read but it means here that res body is empty now nothing to read more
		if err = scanner.Err(); err != nil {
			slog.Error("scanner hit unexpected error", "error", err)
			return
		} //returns non-eof err <- does not point eof response

		// 4. write in same way to file

		//9. send responsifying response
		// switch on *agentConfig.Mode val for response - as it telling to use switch instead
		switch agentConfig.Mode {
		case "docs":
			fmt.Println("Successfully analysed repository & Documentation Success⚡")
		case "review":
			fmt.Println("Successfully analysed repository & review Success⚡")
		case "qa":
			fmt.Println("Successfully analysed repository & questions generation is Success⚡")
		} //..switch

		return
		// now, all data is stored in output file in bytes form

		// err = WriteToFile(ouptutFile,goFilesDataAccumulator)
		// if err != nil {
		// 	slog.Error("failed to write client data","error",err)
		// 	return
		// }
		// slog.Info("Dir analysis is completed & accumulated all files content","ouptutFile",ouptutFile)
		// successfully analysed and worked withd dir✅
		// successfully retrieved all go files data ✅

		// dir acess flow
		// 1. check if it is file or dir with os.Stat watch out for err as they don't always mean it does not exists unless it is explicity checked too
		// 2. if it is a dir, read it with os.ReadDir -> actually it asks for name which -> is basically relative path where are we checking inside
		// if we are checking inside a dir and obvs path becomes current dir -> then os.Stat checks if it is dir or file
		// 3. loop over read dir -> dir has DirEntry elements as it is a slice of []*DirEntry (dir) -> loop over and access each file indivisually
		// early return for testing above functionality\
	}
	//  ** DIR END **//

	// before sending req,manage history if enabled
	// ** conversation history with --mode deeplearning **//
	deeplearningEnabled := agentConfig.Deeplearning == true
	// check if history already exists
	if deeplearningEnabled {
		slog.Info("entered dir mode...")
		fmt.Println("unlocking full potential of agent and powered by deep intelligence⚡")

		//1. read file content
		cnt, err := ReadFileContent("./", agentConfig.FilenameArg)
		if err != nil {
			log.Fatalf("failed to read file content, err- %v", err)
		}

		historyFile := "history.json"
		historyExists, _ := CheckConvoHistoryFILE(historyFile)

		// if history does not exists already
		if !historyExists {
			// create new file
			err := os.WriteFile(historyFile, []byte("[]"), 0644)
			if err != nil {
				return
			}
		}

		// if yes,retrieve historyData in form of []byte -> unmarshal into pqyload -> append new message to it

		//* stores contents slice{user,model...}
		var historyGem []*ContentsSliceKeyWrapperGem
		historyData, err := os.ReadFile(historyFile)
		if err != nil {
			log.Fatal(err.Error())
		}

		err = json.Unmarshal(historyData, &historyGem)
		if err != nil {
			log.Fatal(err.Error())
		}

		// append new user message data to it
		userparts := &PartsSliceKeyWrapperGem{
			Text: BuildPrompt(agentConfig.Mode, cnt),
		}
		newUserHistory := &ContentsSliceKeyWrapperGem{
			Role: "user",
			Parts: []*PartsSliceKeyWrapperGem{
				userparts,
			},
		}

		// !stores content slice - where contentwrappers are pushes
		historyGem = append(historyGem, newUserHistory)

		reqURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

		// ! sending full contents history
		deepOutByteBuffer := &OutboundPayloadGem{
			Contents: historyGem,
		}
		var deepBodyBuf bytes.Buffer //storing bytes data

		// storing in bytes in buffer
		err = json.NewEncoder(&deepBodyBuf).Encode(deepOutByteBuffer)
		if err != nil {
			log.Fatal("failed to encode history")
		}
		// if history is encoded successfully, send to the gemini for collective response

		// reader -> gets [] byte-> goes into writer (bytes) <- based off use/context it writes bytes there
		// writer is satisfied by that method which takes bytes data and write, based of context where it is called on

		// writer -> based of context which writer is passed,they write provided bytes to their respective destinations like http's writer to http, stdout writer to console.

		// 3. creating request for sending
		deepReq, err := http.NewRequest("POST", reqURL, &deepBodyBuf)
		if err != nil {
			log.Fatalf("failed to request AI,err - %v", err)
		}

		deepReq.Header.Set("Content-Type", "application/json") // header key is **case-sensi
		deepReq.Header.Set("x-goog-api-key", apiKEY)           // Pass your Gemini key here

		// 1. Initialize the spinner
		// CharSets[14] is a cool dot-bouncing animation, but there are dozens of options
		// s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Agent is thinking..." // Adds text next to the spinner
		s.Color("cyan")                    // Optional: give it some color!

		// 2. Start the spinner
		s.Start()

		// send to gemini
		deepRes, err := client.Do(deepReq)
		if err != nil {
			return
		}

		// 4. Stop the spinner immediately after the request finishes!
		s.Stop()

		if deepRes.StatusCode != 200 {
			log.Fatalf("API returned a non-200 status code: %d", deepRes.StatusCode)
		}

		// intercept response - get res - write data to the history with role "model" this time
		deepBody := deepRes.Body
		defer deepBody.Close() //deferred call to close reader

		// read incoming body data into byte
		deepResByte, err := io.ReadAll(deepBody)
		if err != nil {
			return
		}

		var deepInbound *InboundPayloadGem
		err = json.Unmarshal(deepResByte, &deepInbound)
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
			log.Fatal("empty deep response from AI, must have hit some unexpected error")
		}

		// if res is validated and good -> contruct content with gemini response with role attached
		deepAIResponse := deepInbound.Candidates[0].ContentWrapperGem.Parts[0].Text

		deepPartsGem := &PartsSliceKeyWrapperGem{
			Text: deepAIResponse, //* attaching deep gemini response to it
		}
		deepContentGem := &ContentsSliceKeyWrapperGem{
			Role:  "model", //* attaching role
			Parts: []*PartsSliceKeyWrapperGem{deepPartsGem},
		}

		// ! appending res content to the history contents
		// appending this to the history - keeping both histories intact
		historyGem = append(historyGem, deepContentGem)

		//* now new history would be added to the history file but we also have to get back now full updated history and write to the file

		// ! grabbing full history {content : [usercontents,geminicontents]}
		updatedHistoryBytes, err := json.Marshal(historyGem)
		if err != nil {
			return
		}

		// writing to the both file - {content - write full encoded contents in history} , res to the output file
		err = WriteToFile(agentConfig.writeToFileArg, AIResponseContent)
		if err != nil {
			return
		}
		err = WriteToFile(historyFile, string(updatedHistoryBytes))
		if err != nil {
			return
		}
		fmt.Printf("deep response is recieved in %s file", agentConfig.writeToFileArg)

		return
	}

	// ** normal file prompting starts here**//

	// check empty first if works otherwise check on set like if it "none" dont do this

	// using none as it would be stillnon nil -> so check if not "none" - then only proceed
	// if agentConfig.FilenameArg != "none" && *Deeplearning == false && *Target != "dir" {

	slog.Info("entered normal prompting mode...")

	// fmt.Println("os args",os.Args)

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

	//1. read file content
	cnt, err := ReadFileContent("./", agentConfig.FilenameArg)
	if err != nil {
		log.Fatalf("failed to read file content, err- %v", err)
	}
	var parts []*PartsSliceKeyWrapperGem
	part := &PartsSliceKeyWrapperGem{
		// ! adding conditional mode prompt + content
		Text: BuildPrompt(agentConfig.Mode, cnt),
	}
	parts = append(parts, part)

	contents := &ContentsSliceKeyWrapperGem{
		Role:  "user", //* while requesting, telling it has user role - user req
		Parts: parts,
	}

	out := &OutboundPayloadGem{
		Contents: []*ContentsSliceKeyWrapperGem{contents},
	}

	outByteBuffer, err := json.Marshal(out)
	if err != nil {
		log.Fatalf("failed to encode outgoing payload,err -%v", err)
	}

	// wrapping payload into io reader -> let servers recieves data in chunks
	body := bytes.NewBuffer(outByteBuffer) // buffer.buffer auto handles read/writes internally

	// 3. creating request for sending
	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		log.Fatalf("failed to request AI,err - %v", err)
	}

	req.Header.Set("Content-Type", "application/json") // header key is **case-sensi
	req.Header.Set("x-goog-api-key", apiKEY)           // Pass your Gemini key here

	// added spinner
	// 1. Initialize the spinner
	// CharSets[14] is a cool dot-bouncing animation, but there are dozens of options
	// s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Agent is thinking..." // Adds text next to the spinner
	s.Color("cyan")                    // Optional: give it some color!

	// 2. Start the spinner
	s.Start()

	//4. initiates the request
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to get res, err - %v", err)
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
	r, err := io.ReadAll(bodyReader) // read at once and tells how much is read
	if err != nil {
		log.Fatalf("failed to read response body byte data, err- %v", err)
	}

	var inboundResponse InboundPayloadGem
	err = json.Unmarshal(r, &inboundResponse)
	if err != nil {
		log.Fatalf("failed to unmarshal recieved response, err- %v", err)
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
		log.Fatal("empty response from AI, must have hit some unexpected error")
	}

	// since choices is a slice -> access its element's msg
	err = WriteToFile(agentConfig.writeToFileArg, AIResponseContent)
	if err != nil {
		log.Fatalf("failed to write to the file, err - %v", err)
	}

	// switch on *agentConfig.Mode val for response - as it telling to use switch instead
	switch agentConfig.Mode {
	case "docs":
		fmt.Println("Documentation Success⚡")
	case "review":
		fmt.Println("analysed Success⚡")
	case "qa":
		fmt.Println("questions Success⚡")
	} //..switch
	// }// == normal prompting if condition

	// bug - but we need a way to make this agent running forever and don't let main func exit
	// * Watcher ( detects file writes change - call agent to review it) -> review for now
	slog.Info("starting watcher", "dir", agentConfig.selectedDirPathArg, "suffix", agentConfig.dirSubFileTypesArg)
	go WatchDirChanges(agentConfig.selectedDirPathArg, agentConfig.dirSubFileTypesArg, agentConfig.Mode, apiKEY, agentConfig.writeToFileArg)
	slog.Info("Agent is watching your changes🤖...")

	// & new - if you add empty select at the end of your entry point -> it blocks main from exiting and keep go routines running forever - go Watcher
	select {}
	// ** deep end //

}
