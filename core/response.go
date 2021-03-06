//    Gonews is a webapp that provides a forum where users can post and discuss links
//
//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>

//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as published
//    by the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.

//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.

//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gonews

import (
	"net/http"
	"sync"
)

// ResponseWriterExtraInterface is a response writer
// enhanced with various apis
type ResponseWriterExtra interface {
	http.ResponseWriter
	IsResponseWritten() bool
	SetSession(SessionWrapper)
	Session() SessionWrapper
	HasSession() bool
	GetCurrentSize() int
	SetLogger(LoggerInterface)
	Status() int
}

// ResponseWriterExtra can notify if a response has been written
type DefaultResponseWriterExtra struct {
	http.ResponseWriter
	Request            *http.Request
	session            SessionWrapper
	hasWrittenResponse bool
	currentSize,
	status int
	logger LoggerInterface

	sync.Once
}

// Session returns a SessionInterface
func (rw *DefaultResponseWriterExtra) Session() SessionWrapper {
	return rw.session
}

// SetLogger sets the logger
func (rw *DefaultResponseWriterExtra) SetLogger(logger LoggerInterface) {
	rw.logger = logger
}

// SetSession sets the session
func (rw *DefaultResponseWriterExtra) SetSession(session SessionWrapper) {
	rw.session = session
}

// HasSession returns true if rw has a session
func (rw *DefaultResponseWriterExtra) HasSession() bool {
	return rw.session != nil
}

// SetWrittenResponse returns true if a response has been written
func (rw *DefaultResponseWriterExtra) SetWrittenResponse() {
	rw.hasWrittenResponse = true
}

func (rw *DefaultResponseWriterExtra) error(messages ...interface{}) {
	if rw.logger != nil {
		rw.logger.Error(append([]interface{}{"ResponseWithExtra.Write"}, messages...)...)
	}
}

func (rw *DefaultResponseWriterExtra) debug(messages ...interface{}) {
	if rw.logger != nil {
		rw.logger.Debug(append([]interface{}{"ResponseWithExtra.Write"}, messages...)...)
	}
}

// Write writes in the response stream
func (rw *DefaultResponseWriterExtra) Write(b []byte) (size int, err error) {
	// save the session once
	rw.Once.Do(func() {
		if rw.HasSession() {
			err := rw.Session().Save(rw.Request, rw.ResponseWriter)
			if err != nil {
				rw.error("Error saving the session ", err)
			}
		} else {
			rw.error("Session not found, can't save... ")
		}
	})

	size, err = rw.ResponseWriter.Write(b)
	if err != nil {
		rw.error(err)
	}
	rw.currentSize += size

	return
}

// GetCurrentSize get size written in response
func (rw *DefaultResponseWriterExtra) GetCurrentSize() int {
	return rw.currentSize
}

// IsResponseWritten returns true if Write has been called
func (rw *DefaultResponseWriterExtra) IsResponseWritten() bool {
	return rw.hasWrittenResponse
}

// WriteHeader writes the status code
func (rw *DefaultResponseWriterExtra) WriteHeader(status int) {
	rw.ResponseWriter.WriteHeader(status)
	rw.status = status
}

// Status returns the current status
func (rw *DefaultResponseWriterExtra) Status() int {
	return rw.status
}
