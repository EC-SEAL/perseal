package model

type DashboardResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	ErrorMessage string `json:"error"`
}

type Redirect struct {
	SessionID   string `json:"sessionId"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Module      string `json:"module"`
}
