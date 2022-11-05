package httpreq

import "fmt"

type UnexpectedStatusCodeError struct {
	StatusCode int
}

func (e UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %v", e.StatusCode)
}
