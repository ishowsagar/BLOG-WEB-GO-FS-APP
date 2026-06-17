package main

import "log"

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
