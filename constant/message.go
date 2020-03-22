package constant

// Define data exchange format client => server
const (
	// login
	MsgLogin = "login"
	// command
	MsgCmd = "cmd"
	// window size change
	MsgResize = "resize"
	// Heartbeat detection
	MsgHeartBeat = "heartbeat"
)

// Form Login form
type Form struct {
	Host      string `json:"host" yaml:"host"`
	Port      int    `json:"port" yaml:"port"`
	User      string `json:"user" yaml:"user"`
	AuthType  string `json:"auth_type" yaml:"auth_type"`
	Password  string `json:"password" yaml:"password"`
	SecretKey string `json:"key" yaml:"key"`
}

// Message Data exchange form
type Message struct {
	Type string `json:"type"`
	Cmd  string `json:"cmd"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
	Form Form   `json:"form"`
}

// NewMessage Return a Message object
func NewMessage(type_ string, data ...interface{}) Message {
	var msg Message
	switch type_ {
	case MsgCmd:
		msg = NewCmd(data[0].(string))
	case MsgResize:
		msg = NewResize(data[0].(int), data[1].(int))
	case MsgLogin:
		msg = NewLogin(data[0].(Form))
	case MsgHeartBeat:
		msg = NewHeartBeat()
	}
	return msg
}

// NewCmd Return type 'cmd' Message object
func NewCmd(cmd string) Message {
	return Message{
		Type: MsgCmd,
		Cmd:  cmd,
	}
}

// NewResize Return type 'resize' Message object
func NewResize(cols, rows int) Message {
	return Message{
		Type: MsgResize,
		Cols: cols,
		Rows: rows,
	}
}

// NewLogin Return type 'login' Message object
func NewLogin(form Form) Message {
	return Message{
		Type: MsgLogin,
		Form: form,
	}
}

// NewHeartBeat Return type 'heartbeat' Message object
func NewHeartBeat() Message {
	return Message{
		Type: MsgHeartBeat,
	}
}
