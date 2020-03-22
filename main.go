package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	gossh "golang.org/x/crypto/ssh"

	"Webterminal/constant"
	"Webterminal/ssh"
	"Webterminal/utils"
	"Webterminal/ws"
)

var routers map[string]http.HandlerFunc

const uploadFileKey string = "Filedata"

// Index home
func Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, struct{}{})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Echo echo ws
func Echo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Upgrade websocket ...")
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(p))
		upperStr := strings.ToUpper(string(p))
		if err := conn.WriteMessage(messageType, []byte(upperStr)); err != nil {
			fmt.Println(err)
			return
		}
	}
}

// Verifiy Read and verify
func Verifiy(conn *websocket.Conn) (*ssh.Client, error) {
	// disconnect if there is no answer within 10 seconds after connecting
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	msgType, p, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("conn ReadMessage failed ...", err)
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return nil, err
	}
	if msgType != websocket.TextMessage {
		fmt.Println("conn ReadMessage msgType not TextMessage...")
		conn.WriteMessage(websocket.TextMessage, []byte("Conn Read Message not Text !"))
		return nil, err
	}
	msgObj := constant.Message{}
	err = json.Unmarshal(p, &msgObj)
	if err != nil {
		fmt.Println("conn ReadMessage msgType not TextMessage...")
		conn.WriteMessage(websocket.TextMessage, []byte("Conn Read Message Format Error !"))
		return nil, err
	}
	if msgObj.Type != constant.MsgLogin {
		conn.WriteMessage(websocket.TextMessage, []byte("Conn Read Message Type Not Is Login !"))
		return nil, err
	}
	form := msgObj.Form
	var key string
	if form.SecretKey != "" {
		pwd, err := os.Getwd()
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Failed to get current folder"))
			return nil, err
		}
		key = path.Join(pwd, "tmp", form.SecretKey)
	}
	client, err := ssh.NewSSHClient(form.User, form.Password, form.Host, key, form.Port, []string{})
	if err != nil {
		fmt.Println("ssh NewSSHClient failed ...", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Auth failed !"))
		return nil, err
	}
	return client, err
}

// ReceiveMessage Before read hook message dispatch
func ReceiveMessage(p []byte, n int, tempP []byte, resizeCh chan<- map[string]int, heartBeatCh chan<- struct{}, timeoutCh chan<- struct{}) (int, error) {
	message := &constant.Message{}
	if err := json.Unmarshal(p[:n], message); err != nil {
		fmt.Println("Postread json unmarshal failed ...", err)
		return 0, err
	}
	switch message.Type {
	case constant.MsgCmd:
		timeoutCh <- struct{}{}
		copy(tempP, []byte(message.Cmd))
		return len([]byte(message.Cmd)), nil
	case constant.MsgResize:
		// window size change message
		if message.Cols > 0 && message.Rows > 0 {
			resizeCh <- map[string]int{
				"cols": message.Cols,
				"rows": message.Rows,
			}
		}
		return 0, nil
	case constant.MsgHeartBeat:
		heartBeatCh <- struct{}{}
		return 0, nil
	}
	return 0, nil
}

// ReceiveResize Listen for resize messages on the pipeline
func ReceiveResize(session *gossh.Session, c <-chan map[string]int) {
	for {
		select {
		case cr, ok := <-c:
			if !ok {
				fmt.Println("resize chan out ...")
				return
			}
			cols := cr["cols"]
			rows := cr["rows"]
			fmt.Println("change size")
			if err := session.WindowChange(rows, cols); err != nil {
				fmt.Println("resize failed ...", err)
				return
			}
		}
	}
}

// ReceiveHeartbeat Listen for heartbeat messages on the pipeline (the client is considered disconnected if there is no ping message within 15 seconds)
func ReceiveHeartbeat(rw *ws.WsWrapper) {
	for {
		select {
		case _, ok := <-rw.HeartbeatCh:
			if !ok {
				fmt.Println("heartbeat chan out ...")
				return
			}
		case <-time.After(15 * time.Second):
			rw.Write([]byte("Lost heartbeat ..."))
			fmt.Println("lost heartbeat ...")
			rw.QuitChan <- struct{}{}
			return
		}
	}
}

// ClientTimeout Judge that the client times out, disconnects if no message is read for 10 minutes
func ClientTimeout(rw *ws.WsWrapper) {
	for {
		select {
		case _, ok := <-rw.TimeoutChan:
			if !ok {
				fmt.Println("timeout chan out ...")
				return
			}
		case <-time.After(10 * time.Minute):
			rw.Write([]byte("conn timeout ..."))
			fmt.Println("conn timeout ...")
			rw.QuitChan <- struct{}{}
			return
		}
	}
}

// CloseChan Close chan
func CloseChan(rw *ws.WsWrapper) {
	close(rw.ResizeCh)
	close(rw.HeartbeatCh)
	close(rw.TimeoutChan)
	close(rw.QuitChan)
}

// SSHHandle handle
func SSHHandle(conn *websocket.Conn) {
	defer func() {
		fmt.Println("conn close")
		conn.WriteMessage(websocket.TextMessage, []byte("Conn Close !"))
		conn.Close()
	}()

	// create ssh connect
	client, err := Verifiy(conn)
	if err != nil {
		fmt.Println("conn websocket failed ...", err)
		return
	}
	fmt.Println("connect succ ...")
	defer client.Close()

	conn.SetReadDeadline(time.Time{})

	// create ssh session
	session, err := client.NewSession()
	if err != nil {
		fmt.Println("new session failed ...", err)
		return
	}
	defer session.Close()

	// terminal options
	modes := ssh.GetDefaultTerminalOption()

	var h, w int
	if client.Rows != 0 {
		h = client.Rows
	}
	if client.Cols != 0 {
		w = client.Cols
	}
	// Request a pseudo terminal
	if err := session.RequestPty("xterm-256color", h, w, modes); err != nil {
		fmt.Println("request pty failed ...", err)
		return
	}

	// create WsWrapper set after read hook
	rw := ws.NewWsWrapper(conn)
	rw.SetPostReadHandle(ReceiveMessage)

	// listen resize
	go ReceiveResize(session, rw.ResizeCh)
	// listen heartbeat
	go ReceiveHeartbeat(rw)
	// determine client timeout
	go ClientTimeout(rw)
	defer CloseChan(rw)

	// docking standard input and output
	session.Stdin = rw
	session.Stdout = rw
	session.Stderr = rw

	if err := session.Shell(); err != nil {
		fmt.Println("session shell failed ...", err)
		return
	}
	// if err := session.Wait(); err != nil {
	// 	fmt.Println("session wait failed ...", err)
	// 	return
	// }
	<-rw.QuitChan

}

// Terminal ssh
func Terminal(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// defer conn.Close()
	fmt.Println("Upgrade websocket ...")

	go SSHHandle(conn)
}

// Test ssh verify test
func Test(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		resp := constant.MakeResponse(400, nil, "Bad Request", false)
		json.NewEncoder(w).Encode(resp)
		return
	}
	auth := constant.Form{}
	json.NewDecoder(r.Body).Decode(&auth)
	var key string
	if auth.SecretKey != "" {
		pwd, err := os.Getwd()
		if err != nil {
			resp := constant.MakeResponse(500, nil, "Failed to get current folder", false)
			json.NewEncoder(w).Encode(resp)
			return
		}
		key = path.Join(pwd, "tmp", auth.SecretKey)
	}
	client, err := ssh.NewSSHClient(auth.User, auth.Password, auth.Host, key, auth.Port, []string{})
	if err != nil {
		resp := constant.MakeResponse(500, nil, err.Error(), false)
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer client.Close()
	resp := constant.MakeResponse(200, nil, "Connect success", true)
	json.NewEncoder(w).Encode(resp)
	return
}

// UploadFile upload file
func UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		resp := constant.MakeResponse(400, nil, "Bad Request", false)
		json.NewEncoder(w).Encode(resp)
		return
	}
	file, header, err := r.FormFile(uploadFileKey)
	if err != nil {
		resp := constant.MakeResponse(500, nil, "Parse form failed", false)
		json.NewEncoder(w).Encode(resp)
		return
	}
	fmt.Println("upload file name : ", header.Filename)
	pwd, err := os.Getwd()
	if err != nil {
		resp := constant.MakeResponse(500, nil, "Failed to get current folder", false)
		json.NewEncoder(w).Encode(resp)
		return
	}
	fileFolder := path.Join(pwd, "tmp")
	if _, err := os.Stat(fileFolder); err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(fileFolder, os.ModeDir)
			if err != nil {
				resp := constant.MakeResponse(500, nil, "Failed to create folder", false)
				json.NewEncoder(w).Encode(resp)
				return
			}
		}
	}
	fileName := fmt.Sprintf("%s-%s", utils.GetRandomString(5), strconv.FormatInt(time.Now().Unix(), 10))
	filePath := path.Join(fileFolder, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		resp := constant.MakeResponse(500, nil, "Failed to create file", false)
		json.NewEncoder(w).Encode(resp)
		return
	}
	_, err = io.Copy(dst, file)
	if err != nil {
		resp := constant.MakeResponse(500, nil, "Failed to save file", false)
		json.NewEncoder(w).Encode(resp)
		return
	}
	data := map[string]string{
		"filename": fileName,
	}
	resp := constant.MakeResponse(200, data, "File uploaded successfully", true)
	json.NewEncoder(w).Encode(resp)
	return
}

// CheckOrigin Handling cross domains
func CheckOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		// w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func main() {
	routers = make(map[string]http.HandlerFunc)
	routers["/"] = Index
	routers["/echo"] = Echo
	routers["/terminal"] = Terminal
	routers["/test"] = Test
	routers["/upload"] = UploadFile

	for k, v := range routers {
		http.Handle(k, CheckOrigin(http.HandlerFunc(v)))
	}
	// static file
	fsh := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fsh))

	fmt.Println("Listen :8999 ...")
	err := http.ListenAndServe(":8999", nil)
	if err != nil {
		fmt.Println(err)
	}
}
