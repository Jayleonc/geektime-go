package async

import (
	"time"
)

type Task struct {
	Id           string
	Name         string
	Type         string
	Parameters   string
	RetryCount   int
	Status       int
	ErrorMessage string
	CTime        time.Time
	UTime        time.Time
}
