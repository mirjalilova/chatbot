package entity

type DashboardActiveUsers struct {
	Day           string `json:"day"`
	ActiveUsers   int    `json:"active_users"`
	RequestsCount int    `json:"requests_count"`
}
