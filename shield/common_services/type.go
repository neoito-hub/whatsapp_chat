package common_services

import "gorm.io/gorm"

type HandlerPayload struct {
	Url         string
	RequestBody string
	UserID      string
	Db          *gorm.DB
	UserName    string
	Queryparams map[string]string
}

type HandlerResponse struct {
	Err    bool   `json:"err"`
	Data   string `json:"data"`
	Status int    `json:"status"`
}

type DBInfo struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
	Sslmode  string
	Timezone string
}
