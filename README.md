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
  
# License
![](https://img.shields.io/badge/License-MIT-blue.svg)
