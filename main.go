// Copyright © 2016 Apigee Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
