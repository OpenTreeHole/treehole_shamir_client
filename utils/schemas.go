package utils

type PGPMessageResponse struct {
	UserID     int    `json:"user_id"`
	PGPMessage string `json:"pgp_message" gorm:"column:key"`
}

type UserShare struct {
	UserID int   `json:"user_id"`
	Share  Share `json:"share"`
}

type UploadSharesRequest struct {
	IdentityName string      `json:"identity_name"`
	Shares       []UserShare `json:"shares"`
}

type UserShareError struct {
	UserShare
	error
}
