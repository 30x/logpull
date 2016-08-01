package main

import (
  "fmt"

  "github.com/30x/logpull/pkg/server"
)

func main() {
  server, err := server.NewServer()
  if err != nil {
    fmt.Printf("Error making server: %v", err)
  }

  err = server.Start()
  if err != nil {
    fmt.Printf("Error starting server: %v", err)
  }

  return
}
