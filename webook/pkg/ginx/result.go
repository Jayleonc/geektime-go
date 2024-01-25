package ginx

import "C"
import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type Page struct {
	List      interface{} `json:"list"`
	Count     int64       `json:"count"`
	PageIndex int         `json:"pageIndex"`
	PageSize  int         `json:"pageSize"`
}

func (r *Response) OK() *Response {
	r.Code = 200
	return r
}

func (r *Response) ReturnError(code int) *Response {
	if r.Code == 0 {
		r.Code = 555
	} else {
		r.Code = code
	}
	return r
}

func OK(c *gin.Context, res Response) {
	c.JSON(http.StatusOK, res.OK())
}

func PageOK(c *gin.Context, page Page, msg string) {
	var resp Response
	resp.Data = page
	resp.Msg = msg
	OK(c, resp)
}

func Error(c *gin.Context, code int, msg string) {
	var res Response
	res.Msg = msg
	c.JSON(http.StatusOK, res.ReturnError(code))
}

// Return 调用方，负责构建 Response
// err != nil, 返回 Response.Code = 200 的响应
// err == nil，默认返回 Response.Code = 555 的响应，表示调用方并没有设置错误 Code
func Return(c *gin.Context, res Response, err error) {
	if err != nil {
		c.JSON(http.StatusOK, res.ReturnError(res.Code))
	}
	c.JSON(http.StatusOK, res.OK())
}
