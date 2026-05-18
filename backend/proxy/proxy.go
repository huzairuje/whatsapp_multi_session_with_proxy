package proxy

// Proxy holds configuration for one proxy endpoint.
// BaseURL is the host (e.g. "198.23.239.134"), Port ("6540"),
// Username and Password for authentication, if required.
type Proxy struct {
	BaseURL  string
	Port     string
	Username string
	Password string
}

// NewManager initializes and returns a new Manager.
// Before assigning or fetching user proxies, you can load a proxy list
// via Manager.LoadFromFile(...) or Manager.ResetProxies(...).
func NewManager() *Manager {
	return &Manager{
		assign: make(map[string]int),
	}
}
