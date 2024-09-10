package api

type PutObjectResponse struct {
	URL  string `json:"url"`
	ETag string `json:"etag"`
	Size int64  `json:"size"`
}

type Address struct {
	AddressID string `json:"id"`
	Receiver  string `json:"receiver"`
	Contact   string `json:"contact"`
	Address   string `json:"address"`
}

type RegisterResponse struct {
	Token string `json:"token"`
}
