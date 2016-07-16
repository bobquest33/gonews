package gonews

import (
	// "fmt"
	"golang.org/x/net/xsrftoken"
)

// TODO fix tokens and sessions

// DefaultCSRFProvider implements CSRFProvider
type DefaultCSRFGenerator struct {
	Session SessionWrapper
	Secret  string
}

// Generate generates a new token
func (d *DefaultCSRFGenerator) Generate(userID, actionID string) string {
	t := xsrftoken.Generate(d.Secret, userID, actionID)
	// tokenNameInSession := fmt.Sprintf("%v-%v", userID, actionID)
	// d.Session.Set(tokenNameInSession, t)
	return t
}

// Valid valides a token
func (d *DefaultCSRFGenerator) Valid(token, userID, actionID string) bool {
	//tokenNameInSession := fmt.Sprintf("%v-%v", userID, actionID)
	// t := fmt.Sprint(d.Session.Get(tokenNameInSession))
	// d.Session.Set(tokenNameInSession, nil)
	// if t != token {
	// 	return false
	// }
	return xsrftoken.Valid(token, d.Secret, userID, actionID)
}