package main

import (
  "fmt"

  "github.com/30x/logpull/pkg/server"
)

func main() {
  server, err := server.NewServer()
  if err != nil {
    fmt.Printf("Error making server: %v", err)
    return
  }

  err = server.Start()
  if err != nil {
    fmt.Printf("Error in server: %v", err)
  }

  return
}
