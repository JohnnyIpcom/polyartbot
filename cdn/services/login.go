package services

type LoginService interface {
	Login(username string, password string) bool
	Logout(username string) bool
}

type loginService struct {
	username string
	password string
}

func NewLoginService() LoginService {
	return &loginService{
		username: "johnnyipcom",
		password: "fDds34s565k",
	}
}

func (l *loginService) Login(username string, password string) bool {
	return l.username == username && l.password == password
}

func (l *loginService) Logout(username string) bool {
	return true
}
