package types

// RateLimit is returned from some http requests
type RateLimit struct {
	Remain, Limit, Reset int
	Set                  bool
	Policies             []string
}
