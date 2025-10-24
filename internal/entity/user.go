package entity

type CreateUser struct {
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	Avatar      string `json:"avatar"`
}

type UserInfo struct {
	ID          string  `json:"id"`
	FullName    *string `json:"full_name"`
	PhoneNumber string  `json:"phone_number"`
	Role        string  `json:"role"`
	Avatar      *string `json:"avatar"`
	CreatedAt   string  `json:"created_at"`
}

type GetMe struct {
	FullName    *string `json:"full_name"`
	Role        string  `json:"role"`
	Avatar      *string `json:"avatar"`
	Language    string  `json:"language"`
	LimitIsOver bool    `json:"limit_is_over"`
}

type UpdateUser struct {
	Id          string `json:"id"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	Avatar      string `json:"avatar"`
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
	Id   string `json:"id"`
	Role string `json:"role"`
}
