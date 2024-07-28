package entities

import "fmt"

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u *User) String() string {
	return fmt.Sprintf("login: %s, password: %g", u.Login, u.Password)
}
