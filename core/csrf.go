//    Gonews is a webapp that provides a forum where users can post and discuss links
//
//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as published
//    by the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.
//
//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
