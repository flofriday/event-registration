package eventregistration

import "time"

type User struct {
	UUID        string
	FirsName    string
	LastName    string
	Email       string
	Phone       string
	RegisterdAt time.Time
}
