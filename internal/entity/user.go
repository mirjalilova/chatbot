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
	PhoneNumber string `json:"phone_number"`
}

type VerifyReq struct {
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
}

type LoginRes struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

type GetByPhone struct {
	Id string `json:"id"`
	Role string `json:"role"`
}

