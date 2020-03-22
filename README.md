# Webterminal ![](https://img.shields.io/badge/language-golang-blue)  
A web terminal by golang
# Install
`go build .`  
If you do not enable go modules, run `export GO111MODULE=on`, then run `go build .`  
If you are in China, you can use a proxy download package, run `export GOPROXY=https://goproxy.cn`
# Usage
`./Webterminal`  
# API
  - GET / index (Have a test case)
  - GET /echo echo command (Test websocket)
  - GET /terminal open websocket and ssh (You have to submit ssh verification form via websocket within 10 seconds, Otherwise it will disconnect overtime)
  - POST /test verify the validity of the ssh form
  - POST /upload upload file (ssh private key)
# Note  
  ## Data exchange format
  ```golang
  // login form
  type Form struct {
    Host      string `json:"host"`
    Port      int    `json:"port"`
    User      string `json:"user"`
    AuthType  string `json:"auth_type"` // password or secretkey
    Password  string `json:"password"`
    SecretKey string `json:"key"` // upload file return filename
  }
  // exchange message
  type Message struct {
    Type string `json:"type"` // login or cmd or resize or heartbeat
    Cmd  string `json:"cmd"` // command 
    Cols int    `json:"cols"` // resize col
    Rows int    `json:"rows"` // resize row
    Form Form   `json:"form"` 
  }
  ```
# Test case
access / and click 'start' button
# License
![](https://img.shields.io/badge/License-MIT-blue.svg)
