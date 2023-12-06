package domain

import "time"

type User struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string
	// YYYY-MM-DD
	Birthday time.Time
	AboutMe  string

	Phone string

	// UTC 0 的时区
	Ctime time.Time
	Utime time.Time

	WechatInfo
}
