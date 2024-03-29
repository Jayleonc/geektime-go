package vo

type UserEditReq struct {
	// 改邮箱，密码，或者能不能改手机号

	Nickname string `json:"nickname"`
	// YYYY-MM-DD
	Birthday string `json:"birthday"`
	AboutMe  string `json:"aboutMe"`
}

type UserSignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type UserLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SendSMSLoginReq struct {
	Phone string `json:"phone"`
}

type LoginSMSReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}
