package common

type Report struct {
	Device             string
	Address            string
	Success            bool
	ErrorMsg           string
	ResolutionFunction func() (string, error)
}
