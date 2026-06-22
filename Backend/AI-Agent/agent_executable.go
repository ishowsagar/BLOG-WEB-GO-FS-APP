package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

// func that takes in git command which -> return ioreader which -> stores output bytes
func 	RunGitOperator(agrs ...string) (io.Reader,error) {

	// 1. trigger any git executable command
	cmd := exec.Command("git",agrs...) // some commands need more args so rest goes here if provided in case of commit like what it bugged the system

	cmd.Stderr = os.Stderr

	// 2. Get pipe which recieves the ouput in bytes ( in chunks )
	outputPipe,err := cmd.StdoutPipe()
	if err != nil {
		return nil,err
	}
	
	// 3. start command which -> pipes down response to the pipe
	err = cmd.Start()
	if err != nil {
		return nil,err
	}
	
	// 4. Read bytes from that pipe source
	out,err := io.ReadAll(outputPipe)
	if err != nil {
		return nil,err
	}
	// 5.let it finish piping down the executable response
	err = cmd.Wait()
	if err != nil {
		return nil,err
	}	

	// 6. return the response as ioreader buf -> so any reader can recieves bytes data from source directly
	buf := bytes.NewBuffer(out)
	return buf,nil
}