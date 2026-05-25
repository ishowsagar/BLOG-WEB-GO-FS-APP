package main

import (
	"fmt"
	"log"

	"github.com/ishowsagar/go-blog-web-application/utils"
)

func main() {
	token, err := utils.GenerateToken(16)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}
	fmt.Println(*token)
}
