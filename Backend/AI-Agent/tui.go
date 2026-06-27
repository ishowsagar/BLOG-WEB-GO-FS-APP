package main

import (
	"log/slog"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// @ handles all the terminal user interface work
// all are under package main cause since we already inside mod, don't wanna clash with it;atleast for now

// type struct that stores information about model that we'd be running as tui
type TUIModel struct {
	Choices []string // list of available choices to choose from
	Cursor int // int as in current positon of user cursor from choices ;choices element posi decides that
	Selected string // traces user prompted selection - //!map for multi selection but int just for single selection
	Prompt string // storng prompts 
}

// func that fires TUI -> returns user selection, handel err
func IntializeTUI(prompt string,choicesSlice []string) (string,error){

	// ** since choices is all that given to user as prompted choice under message - dynamic now-> based on passed slice of string -> it will show choices to choose
	// intializes tui model which is pointer recieved on methods to update events ( tui like events )
	tuiModel := &TUIModel{
		Choices: choicesSlice,
		Prompt: prompt, // user selected prompt print by view as returned string from view printe by bubbletea
	}

	// now since TUIModel has all the methods that tea.Model interface exoects -> program ready to be fired
	TUI := tea.NewProgram(tuiModel)

	// calling to fire the TUI - handeling all cases - view {cursor based}, update {update cursor and selection}
	finalModel,tuierr := TUI.Run()
	if  tuierr != nil {
		slog.Error("failed to load TUI","error",tuierr)
		return "",tuierr
	}

	// once it ran and user did it -> we get finalModel -> type assert againt tui to get selected
	userSelection := finalModel.(*TUIModel).Selected
	// ! bug - pre append color in view, not here -> causing wrong string to be selected and from chocices as this comes pre appended
	// * fix -  remove it
	return userSelection,nil // gives selected choice - dynamic

	// just now we bave to make sure prompted message is dynamic and rest cursor is handled by upodate which doe snot nees a change ig

}

// init the tui framework into action, return nil as just like uppgrader's in ws for checkOrigin that returns true for upgrading the conn
func(t *TUIModel) Init() tea.Cmd{

	return nil
	// view ->  renders and intializes tui

}


// method that returns string that would be rendered on the console and every keystroke is cursor is handled by bubbletea
func(t *TUIModel) View() string {
	
	// render accumulated string
	lineBreak := "\n"
	var builder strings.Builder // strings builder which builds string in memory and accumulate all and with string() gives the final string

	
	// 1. showing this string first - as builder accumulate and this is how app will progress

	//bug - string concatenation inside builder is wrong you must call them consecutively to align them
	// fix - write progressively for concat

	// this is what prompted to when agent runs
	// ** since we have access to reciever type -> that's the best use of it -> pull prompt from there 
	builder.WriteString(lineBreak)
	builder.WriteString(proBlue)
	builder.WriteString(t.Prompt)
	builder.WriteString(lineBreak)

	// 2. looping over choices to check if current index matches tea's cursor

	for currentChoiceIndex,currentchoice := range t.Choices{
		// checking if current choice's index matches cursor's then we 'prefix - >' onto that
		// otherwise prefix with spaces
		if currentChoiceIndex == t.Cursor {
			builder.WriteString( Green +"> ") // manually concatenating cursor to choices selection
			builder.WriteString(currentchoice) 
			builder.WriteString(Reset+"\n") 
		}else {
			builder.WriteString(" ") // if cursor not there, remember cursor is tracked by framework -> use info -> if matches cursor -> if not -> no cursor show
			builder.WriteString(currentchoice) 
			builder.WriteString("\n") 
		} // builder builds strings and as upon called writes progressively
	}

	// ! ah catch -> since keypress update causes -> loop to look for cursor which matches current choice index -> and string get build from it
	// ** whatever is returned by the view -> printed to the console by the bubble tea -> that's how interfaces shines
	return builder.String()
}


// method that belongs to TUI models which -> updates ui and do actual selections - based off updating cursor positon ( here it is done which runs view to check that - so when this upfdate sthat run to chekc whats the curent posi -> render that string builded tuui)
func(t *TUIModel) Update(keypressEventMsg tea.Msg) (tea.Model,tea.Cmd) {

	//** it updates the cursor position based on keypress/userInput -> triggers new view -> 
	// cause we update the cursor and based off cursor it chnages selection -
	//  ohhh we need to do the selection part too ig
	
	
	//& everykeypress is interceptd by tea as keyMsg, so we check which we recicved and based off that -> update cursor to update view ( only visual update) 
	switch keypressEventMsg := keypressEventMsg.(type) { // strict type check injection for //!type assertion <- otherwise it won't load that type
	
	// if that is a key press, could be any
	case tea.KeyMsg :	// check this first so keypres..msg becomes available
		
		// switching on if it was a keyMsg and getting what was that actual msg with .String() on recievedKeyPress
		switch keypressEventMsg.String() {
			
			// ! now check for cases as this gives us which keypressed that msg was --
		case "up" :
			// t should be pointer for actual changes
			if t.Cursor > 0 {
				t.Cursor-- // decrement 
			}
		case "down" :
			// only increment till choices length , e.g for 2 only 2, so if we keep updatig it would update wrong tui never updated
			if t.Cursor < len(t.Choices) {
				t.Cursor++ // increment as down is considered positive in TUI
			}
		case "enter" :
			// this would update selection, other were updating cursor posiiton <- obvs firstly those cases fire and when keypress is intercepted as enter -> update selection -> might need to show and store that later 
			t.Selected = t.Choices[t.Cursor] // store selected in model from -> based off choices which choice it matches
			// bug - selected not happened
			// fix - must add t.Quit cmd to let it know about this event
			return t,tea.Quit
		
		// bug - since bubble takes over terminal - had to intercept all keypresses to launch events
		// fix - add missing keypresses and later if needed, intercept and decipher here to run required cmd to run operation that's what tea.Cmd is what for
		case "ctrl+c","q" :
			// ! concate shorcuts with + ; as it comes event as  combined event
			return t,tea.Quit  
		}
		
	}

	return t,nil // now t satisfied the t.Model as it has all three methods that method had on it

}

