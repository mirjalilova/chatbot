package entity

type CreateUser struct {
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type UserInfo struct {
	ID          string `json:"id"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	CreatedAt   string `json:"created_at"`
}

type UpdateUser struct {
	Id          string `json:"id"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
}

type UpdateUserBody struct {
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
}

type UserList struct {
	Users []UserInfo `json:"users"`
	Count int        `json:"count"`
}

type LoginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRes struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}
