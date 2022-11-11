package utils

type CustomerError struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data, omitempty"`
}

func (p *CustomerError) Error() string {
	return p.Msg
}

func NewCustomerError(code int, msg string, data interface{}) error {
	err := &CustomerError{
		Code: code,
		Msg:  msg,
	}

	if data != nil {
		err.Data = data
	}

	return err
}
