package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// func that prints current dir (from execution) items to the console
func ls() {

	// get current working dir
	currentdir, err := os.Getwd()
	if err != nil {
		return
	}
	fmt.Printf("getWd - %s", currentdir)

	// read dir - if really dir - stat does syscall to grab metadata from the kernel -> cheap payhead -> everything about it in stat

	// ** oh...since go is parent process running this program -> it asks OS about file descriptors of various things to work with ( not seperate process is fired but running process asks OS about fd)
	entryInfo, err := os.Stat(currentdir)
	if err != nil {
		return
	}

	// if it really dir
	if entryInfo.IsDir() {

		// in this case too -> go process asks OS what fd is assigned to this currentdir -> so to read from here <- by kernel
		dir, err := os.ReadDir(currentdir)
		if err != nil {
			return
		}

		// this gives us all entries in []*DirEntry <- each el is dirEntry -> could be file or dir
		for _, dirEntry := range dir {
			// iterate through each entry -> write to the console

			// get underlying information about the dir entry
			fileinfo, err := dirEntry.Info() // this again triggers os.Stat syscall to grab (inode- this is inode) metdata related to it
			if err != nil {
				return
			}

			filesize := fileinfo.Size()

			dirEntryLister := map[bool]string{
				true :"folder", // if true when dir is true -> folder
				false :"file",// false file
			}

			dirEntryType := dirEntry.IsDir() //returns true when dir, false when file
			foundType := dirEntryLister[dirEntryType]

			n := fmt.Sprintf("%s - %s", foundType,fileinfo.Name())
			s := fmt.Sprintf("filesize - %dkb", filesize/1024)

			src := strings.NewReader("\n" + "📂 " + n + "  " + s)
			buf := make([]byte, 8)

			for {
				rby, err := src.Read(buf)
				if rby > 0 {

					chunk := buf[:rby]
					// * each entry carries metadata -> that's from where name is coming from... -> load that too
					os.Stdout.Write(chunk)
					time.Sleep(time.Millisecond * 40)
				}

				if err == io.EOF {
					break
				}
				if err != nil {
					return
				}
			}
		}

		fmt.Println("\nDir loaded🚨")
		// ** Result 1 => it worked as intended -> for every child proccess it asks os for its fd -> pases to kernel -> syscall -> return res to process back
		// ** and  for native working -> os.Getwd -> making sys calls to the kernel <- by firing process which asks OS about fd of this dir Too -> pass to kernel for returned res -> same triggered for inode( retriggered on each iterated dir entry) -> syscall via that entry fd
		
		// in this way we could make it so - reads from anywhere as another reader, but write to pw - make it dynamic & flexible across systems
	
	}
}