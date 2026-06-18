package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

// builds prompt based on provided mode ["docs"...] and append content to it for sending request
func BuildDirPrompt(mode, content string) string {

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

	dirContext := " and remember it contains all files data in one go, anaylyse files and related code, code is seperated by files boundaries marked with lines like starting with this file name and ends with this file name,for clearance"
	fileContextConfirmation := "and label all the files you reviewed"
	// if provided mod exists -> generate prompts based off that
	var prompt string
	switch mode {
	case "review":
		prompt = "You are an expert code reviewer, review the provided code and give me the short review upto 100 words only but expert review" + dirContext + fileContextConfirmation
	case "docs":
		prompt = "You are an expert code docs generator, analyse the provided code and document it upto 100 words only but expert documentation" + dirContext + fileContextConfirmation
	case "qa":
		prompt = "You are an expert coding mentor, review the provided code and ask me max 5 question in a way it makes me click topic better " + dirContext + fileContextConfirmation
	default:
		prompt = "You are an expert code reviewer, review the provided code and give me the short review upto 100 words only but expert review" + dirContext + fileContextConfirmation
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



