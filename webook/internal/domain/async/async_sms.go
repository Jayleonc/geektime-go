package async

type Sms struct {
	Id      int64
	TplId   string
	Args    []string
	Numbers []string
}
