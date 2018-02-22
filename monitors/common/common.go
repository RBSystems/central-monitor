package common

type Report struct {
	Device             string
	Address            string
	Success            bool
	ErrorMsg           string
	ResolutionFunction func() (string, error)
}

type Device struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Type    string `json:"type"`
}
