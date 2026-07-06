package main

import (
	"io"
	"os"
)

// shell => idea behind building a mini shell is to -> learm deeper about core concepts like writer,reader,proccesses,pipes etc...


func main() {


// color codes for console typography
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
		
	blue      = "\033[94m" // Professional Slate Blue
	proGreen  = "\033[92m" // Crisp Bright Green
	proRes    = "\033[0m"  // Clean State Reset
)

	// ** but it would not run 'ls' -> as this is a builtin program, not external binary which found in the path to run the executable
	// so to run native shell process -> need to tell to run shell process -> cmd /C "processName"


	// ** when any binary executable program is run,it is running on computer's disk -> when child process is spawmed to run that provided executable binary
	// ** it would not be finding exe~cutable from computer's thousand files but from the env -> where binary paths are already set and listed -> checks which matches firstly -> runs it
	
	// & $Path -> just env variable which stores path to the executable binary <- seperates from window's other files runnables


	// it depends on current running shell -> cause since running native shell process -> need to run exactly those that's why dir worked with cmd /C but not sh -c ls
	// resReader,err := ExecutableRunner("cmd","/C","dir")
	// if err != nil {
		//  so when it says it could not run executable due to no path found for provided program/command -> it means either program does not exists or if exists,
		//  its path is not configured in the env variables -> as it searches from the env path list, not from whole disk
	// 	log.Fatalf("failed to run exectuable, err - %v",err)
	// 	return
	// }


	// ** File descriptors => every file or folder has no meaning in their own just like raw bytes read ( we did prev)...when a process is fired -> it asks os what file descriptor is assigned to this file/dir/any resource
	// ** it asks the os about that file/dest/resource as those things are already baked into the os -> os tells the "int~number" <- which is also known as fileDescriptor
	// ** which is actually a number mapping to a resource in the os...-> ultimately writes,reads,pipes all are just fired from process's, retrived their file desriptors -> maps to some resource -> writes/reads/fncstioalites to them -> os handles that which is being written where as os tells about the mapping


	// even pipes reads, file writes <- process asks os what file descriptor is assigned to the that context file/resource/reader anything, even stdrr/in/out comes with their fd
	// so once fd is known to the process -> it writes/reads fromt that fd which is mapped to some resource.


	// file descriptors are assigned to every program/child process/thing/file/folder/resource -> so everything already pre baked into OS and has assigned a file descriptor number
	// so, process {it could be any} asks os about this referred resource -> os gives file descriptor of the resource -> so when process is fired on something, it uses that file descriptor to do read/write to that resource as per functinoality.

	// it successfully piped the started process output to the pipe - read from it -> returned reader <- source from where read bytes -> read bytes -> write those bytes
		// 04-07-2026  11.54 AM  2,529 executable.go
		// 04-07-2026  11.57 AM  909 main.go
		// 23-06-2026  10.42 AM  18 makefile


	// get bytes data from the returned reader

	// in this case -> process is reading bytes from source,it does not know about source of bytes but os does , so process asks which fd is assigned to the reader
	// so proccess eventually gets the fd,reads from the resource mapped fd. 
	// out,err := io.ReadAll(resReader)
	// if err != nil {
	// 	log.Fatalf("could not read output from reader, err - %v",err)
	// 	return
	// }
	
	
	// // write to the console
	// os.Stdout.Write(out) // this process too asks os for fd -> where to write -> kernel -> user space -> res

	// test 1 -> Implementing ls with native os calls, non binary exectubales to run them
	// done


	// test 2 -> Implementing cd -> change dir & ls (ls - later) 
	
	// instead of harcoded -> user prompted action
	
	// parentDir := os.Args[1]
	
	// changes cursor to the provided dir
	
	// ** kernel does change dir according to provided fd and all things passed from user space but it does not change parent process state ( dir ) - it exists, ceases that pwd operation too
	// **child processes cannot change parent processes state -- that's why forked child process could not change pwd when it was changed by the kernel
	// err := os.Chdir(parentDir) //* process asks OS fd of this parentDir too - it does not what parentDir is untill it asks OS about fd of this resource/thing -> pass to kernel for result
	// if err != nil {
	// 	return
	// }


	// go routines exists in user space only cause,since goruntime exists in user space, it allocates low memory (~2kb) threads in queue, with execution position to track it...
	// ** thus go routines are light weight (near abt ~2kb ) threads, go runtime scheduler decides which go routines go where,as all is done in user space where runtime exists
	// ** unless actual I/O operations are not done kernel does not involve -> cause that then triggers syscalls -> kernel does work -> return res to the user space (go runtime) 
	// since they share same process memory state -> read writes happens concurrently
	// go func() 

	// **now pwd (present working directory aka pwdir) changes to this provided dir
	// ls on this dir - as now pwd changes
	// ls()



	// Test 3 - go version -> running go binary executable <- searches in path to execute the user cmd
	// readSrc,err :=ExecutableRunner("go","version")
	// if err != nil {
	// 	log.Fatalf("failed to run executable error - %v",err)
	// }
	// readBytes,err := io.ReadAll(readSrc)
	// if err != nil {
	// 	log.Fatalf("failed to read bytes error - %v",err)
	// }

	// _,err = os.Stdout.Write(readBytes)
	// if err != nil {
	// 	log.Fatalf("failed to write out bytes error - %v",err)
	// }

	// ** Result 3 -> Successfully ran go command as exec runs external binary executable from path stored in env variable as "go" and user cmd -> get pipe out -> read


	// test 4 -> implementing cd => actually changes dir to the desired path
	// err := os.Chdir(os.Args[1]) 
	// if err != nil {
	// 	log.Fatalf("failed to change dir, err - %v",err)
	// }

	// ** Result -> since builtin cannot be called on parent process -> only forked child process inside kernel could be the ones that could be aletered 
	

	// & Phase 2 -> pipes 

	// idea behind pipe is to pipe content/payload from one process to another where there is no shared memory in general.
	// Go routines are fired - threads to move data - application must not be exited as this is done seperately to pipe data

	// implementation - creating a custom pipe which pipes data to read from pipe and write to the pipe...as this is done via go routines calls, for non-blocked write so one end writes and other reads 
	// io.Pipe -> gives two pipe ends : pipeReader and pipeWriter both satiesfies r/w interfaces and closers too
	// now two seperate go routines must be fired -> to where one reads from pr and one writes to the pw.


	// maybe we would need mutex for lock/unlock for more go routines


	// when we take a look at io.Pipe() -> it returns two things - pipeReader and pipeWriter - cause pipe reads from one src and writes to one dest for its piping work. 
	// * io.Pipe() -> they both satiesfies ioReader/writer interfaces - but since data is speedly moved it moves one lane -> in go routine -> one reads from writer for fast-non blocked unbuffered exchange of data.
	
	// this gives us pioes where one reads ( pipe reader ) and one for writing into the pipe (pipe writer)
	// needs two seperate goroutines which -> writes into pipeWriter which saties iowriter & reads from pipeReader satiesfying ioreader 
	pr,pw := io.Pipe()
	go PipeWriter("main.go",pw)
	
	// writer and reader closer are just interfaces which have close method which -> closes either read/writer operations cause :
	// 1. Since everything that involves os resources either in read/writer -> have fd ( file descriptors ) -> Process aks OS for fd of them ->
	// **cause fd resources are being consumed by kernel -> it must be closed explicitly by telling kernel to defer close
	// 2. Since also kernel never firsts writes/reads into fd, it holds operation call into temp buffer, so first it done into buffer to close the streams and write/reads to the disk locally after pending buffers are flushed
	// ! because if kernel directly does the operation on the disk -> a unexpected crash could cause issues -> no backup -> defer makes no sense as defer is just then call to do the job on the disk/os resources by flushing pending temp buffers 
	
	// both things -> releasing of fd after temp buffered the data -> are done with single .close call <- invoked on any reader/writer call
	// ** ohhhhh.. defer is a syscall to close operation : flushing pending buffer writes,realising acqiuired fds... <- fd is passed to the kernel by the process as it asks OS fd of the resource -> sys call -> flush buffers and release fd


	// bug - goroutines deadlock bugg as when both hit err or done select has left with no make backgroung go routines - reduntant workflow
	// fix - making sure to always defer close writer and firing pipeReader which reads from pipeReader without go routine - let work wrtier as this won't block writer go routine as it long as writer is runnig and file is read -> it reads too naturally block it indirectly
	
	 PipeReader(pr,os.Args[1])// when read would be finished -> writer must be closed - closed via writeCloser which satiesfied by any writer who write is then closed 
	//  pipe phase -> pipe is working by implementing ioPipe where go routines writes to the pw and reads from it ( readall - kept reader untill stream not finished )

	// ** Result of pipe -> Pipe gives us a writer and reader, writer must be called in go routine for non blocked write & read from pipe as one end non blocked write and read from src untill writer closes...
	// that's how reader just reads from the pipe when data is written to the pipe write's end.
}