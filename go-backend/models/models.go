package models

type UpdatePasswordRequest struct {
	Username    string `json:"username"`
	Email       string `json:"emailID"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
	ServerIP    string `json:"serverIP"`
	Database    string `json:"database"`
}

type ResetRequest struct {
	RequestID     int    `json:"requestID"`
	Username      string `json:"username"`
	ServerIP      string `json:"serverIP"`
	RequestType   string `json:"requestType"`
	RequestStatus string `json:"requestStatus"`
	RequestTime   string `json:"requestTime"`
}