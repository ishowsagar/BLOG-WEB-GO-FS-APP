package main

import (
	"bytes"
	"io"
	"os/exec"
)

//goal is to run terminal commands and write concurrent/final output to the console - ls,pwd,... basic commands implementation

// & Since os/exec searches for binary executable in the env variables that would be listed as path -> but one thing that has to be moticed
// => it is running external services binary executables which exists in the path as env variables
// ** for native shell calls, it would not be the need of the hour to call shell process fromt the os's exec since its most idiomatic way to use it for external binary program executables
// since also,the executables that has to be run exists as path in env variables to fire those executables, these do not exists like that.

// it would not be idiomatic to execute these builtin programs from thom exec since they all are available in the syscall(need to learn more about that).

// & syscall -> any proccess that intends to do something either reader/writes it knows from where or where to via os's provided fd but actual thing is not done by process
// ** this work is done in syscall interface in 'kernel space' ( not literal go interface but mechanism ) -> its a space ( called as kernel space under syscall boundary ) where processes redirects information about the job and resource ( this lives in user space - user prompted action and job description and all) to the kernel space ( a special zone where kernel has access to the hardware, to do read/writes all those things internally)
// ** --> information from process ( all those things lives in user space ) is passed to the 'kernel space' -> like proccess retrieved context fd of the resource/thing and what to either read from or write data to fd -> kernel has access to hardware and disks which eventually do the job-> return result
// ** result is then returned back to the user space - to process eventually

// boundary that keeps kernel space sandbox environment seperated from user space (where user prompted actions/process live where programs run by users lives) -> syscall boundary

// ! everything is wrapped under this interface boundary -> syscall is intiated to do the task <- os level go core things that are done, all are done in syscall wrappers
// ** syscall wrapper nothing but space where kernel is working to do the passed job

// func that runs executables from *os level , args... for rest of all string args
func ExecutableRunner(programExecutable string,args ...string)(io.Reader,error) {

	// Start process


	// *from $path => exec.Command searches for the provided binary executable in env that exists as a path -> if found -> runs the prompted command from it
	//** every binary executable is just a runnable program with resources -> had to be run to execute them -> so, when this exec.Command fires up ->
	// ** looks in for that path in the env variables (exists as a path for that program to be invoked from anywhere)  which matches the provided binary executable proram to be run
	
	// so exec.Command finds the path listed in the env variables to run that executable binary program to run the prompted command

	// exec.Command returns -> process cmd struct -> which stores information like all three streams and env all...-> methods defined on it to run desired command 
	cmd:= exec.Command(programExecutable,args...) //--> runs child process in the console 

	//** it returns instance of type *exec.cmd -> stores information about the cmd process
	// by default instance comes with attached streams to the console <- that's how default is done -> as when cmd instance is created -> it attaches those streams & information to the console destination.  

	// cmd output goes to the stdout pipe of the cmd

	// **oh...by default it would have printed output to the console (as by default dest is console ) but since we retrieved the pipe -> output is piped here and could be sent anywhere
	outPipe,err := cmd.StdoutPipe() // gives the pipe whose output is redirected to the pipe
	if err!= nil {
		return nil,err
	}

	// running the command outputs to the pipe

	// when cmd is ran -> it keeps sending output and wait untill it is finished -> so have to read concurrently from the pipe before it finishes and exits
	// run -> runs to the full completion and waits untill that -> pipe ends -> exits 
	// but start immediatly returns data -> cause pipe is retrieved from process -> stdout goes to the pipe -> then concurrent read from pipe is possible
	
	// ** when cmd is started -> it fires the process -> writes out to the stdout pipe (cause since stdout is wired to the pipe, overidden from default)
	// ** so it keeps redirecting output to the pipe <- from where concurrent reader is reading from the source as that pipe untill it finishes
	err = cmd.Start()
	if err != nil {
		return nil,err
	}

	// after running get the output from the pipe
	outputBytes,err := io.ReadAll(outPipe)
	if err != nil {
		return nil,err
	}

	// let it finish gracefully - os level cleanup 
	err = cmd.Wait()
	if err != nil {
		return nil,err
	}

	// creating buffer from bytes and returning it as reader to read bytes from source <- comes handy when needed direct sources
	buffer := bytes.NewBuffer(outputBytes)
	return buffer,nil
}