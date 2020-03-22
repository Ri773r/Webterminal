package constant

// Response Define response format
type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
	Succ bool        `json:"succ"`
}

// MakeResponse Return Response object
func MakeResponse(code int, data interface{}, msg string, succ bool) Response {
	return Response{
		Code: code,
		Data: data,
		Msg:  msg,
		Succ: succ,
	}
}
