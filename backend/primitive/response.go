package primitive

import "net/url"

type Devices struct {
	PushName   string `json:"pushName"`
	Platform   string `json:"platform"`
	User       string `json:"user"`
	Server     string `json:"server"`
	IsLoggedIn bool   `json:"isLoggedIn"`
}

type DevicesWithProxy struct {
	PushName   string   `json:"push_name"`
	Platform   string   `json:"platform"`
	User       string   `json:"user"`
	Server     string   `json:"server"`
	ProxyURL   *url.URL `json:"proxy_url"`
	IsLoggedIn bool     `json:"is_logged_in"`
}

type StatusResponse struct {
	ID       string `json:"id"`
	PushName string `json:"pushName"`
	IsLogin  bool   `json:"isLogin"`
}
