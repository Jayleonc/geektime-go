package web

type Response struct {
	RequestId string      `json:"requestId"`
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
}

func (r Response) ok() {

}
