// @Copyright (c) 2016 mparaiso <mparaiso@online.fr>  All rights reserved.

package gonews

import (
	"database/sql"

	"log"
	"net/http"
	"os"

	"fmt"

	"errors"

	"github.com/gorilla/sessions"
)

// Any is any value
type Any interface{}

// ContainerOptions are options provided to the container
type ContainerOptions struct {
	DataSource,
	Driver,
	Secret,
	Title,
	Slogan,
	Description,
	TemplateDirectory,
	TemplateFileExtension string
	Debug bool
	LogLevel
	Session struct {
		Name         string
		StoreFactory func() (sessions.Store, error)
	}
	ConnectionFactory func() (*sql.DB, error)
	LoggerFactory     func() (LoggerInterface, error)
	csrfGenerator     CSRFGenerator
	user              *User
}

// Container contains all the application dependencies
type Container struct {
	ContainerOptions  ContainerOptions
	db                *sql.DB
	logger            LoggerInterface
	threadRepository  *ThreadRepository
	userRepository    *UserRepository
	commentRepository *CommentRepository

	template TemplateEngine

	sessionStore sessions.Store
	request      *http.Request
	response     ResponseWriterExtra

	CSRFGeneratorProvider
	TemplateProvider
	SessionProvider

	user *User
}

func (c Container) Debug() bool {
	return c.ContainerOptions.Debug
}

func (c *Container) SetDebug(debug bool) {
	c.ContainerOptions.Debug = debug
}

// Request returns an *http.Request
func (c *Container) Request() *http.Request {
	return c.request
}

// SetRequest sets the request
func (c *Container) SetRequest(request *http.Request) {
	c.request = request
}

// SetResponse sets the response writer
func (c *Container) SetResponse(response ResponseWriterExtra) {
	c.response = response
}

// ResponseWriter returns the response writer
func (c *Container) ResponseWriter() ResponseWriterExtra {
	return c.response
}

// HasAuthenticatedUser returns true if a user has been authenticated
func (c *Container) HasAuthenticatedUser() bool {
	return c.user != nil
}

// SetCurrentUser sets the authenticated user
func (c *Container) SetCurrentUser(u *User) {
	c.user = u
}

// CurrentUser returns an authenticated user
func (c *Container) CurrentUser() *User {
	return c.user
}

// GetSecret returns the secret key
func (c *Container) GetSecret() string {
	return c.ContainerOptions.Secret
}

// GetConnection returns the database connection
func (c *Container) GetConnection() (*sql.DB, error) {
	if c.ContainerOptions.ConnectionFactory != nil {
		db, err := c.ContainerOptions.ConnectionFactory()
		if err != nil {
			return nil, err
		}
		c.db = db
	} else if c.db == nil {
		db, err := sql.Open(c.ContainerOptions.Driver, c.ContainerOptions.DataSource)
		if err != nil {
			return nil, err
		}
		c.db = db
	}
	return c.db, nil
}

// GetThreadRepository returns a repository for Thread
func (c *Container) GetThreadRepository() (*ThreadRepository, error) {
	if c.threadRepository == nil {
		db, err := c.GetConnection()
		if err != nil {
			return nil, err
		}
		c.threadRepository = &ThreadRepository{DB: db, Logger: c.MustGetLogger()}
	}
	return c.threadRepository, nil
}

// MustGetThreadRepository panics on error
func (c *Container) MustGetThreadRepository() *ThreadRepository {
	r, err := c.GetThreadRepository()
	if err != nil {
		panic(err)
	}
	return r
}

// GetUserRepository returns a repository for User
func (c *Container) GetUserRepository() (*UserRepository, error) {
	if c.userRepository == nil {
		db, err := c.GetConnection()
		if err != nil {
			return nil, err
		}
		logger, err := c.GetLogger()
		if err != nil {
			return nil, err
		}
		c.userRepository = &UserRepository{db, logger}
	}
	return c.userRepository, nil
}

// MustGetUserRepository panics on error or return a repository of User
func (c *Container) MustGetUserRepository() *UserRepository {
	r, err := c.GetUserRepository()
	if err != nil {
		panic(err)
	}
	return r
}

// GetCommentRepository returns the repository of comments
func (c *Container) GetCommentRepository() (*CommentRepository, error) {
	var (
		err    error
		db     *sql.DB
		logger LoggerInterface
	)
	if c.commentRepository == nil {
		db, err = c.GetConnection()
		if err == nil {
			logger, err = c.GetLogger()
			if err == nil {
				c.commentRepository = &CommentRepository{db, logger}
			}
		}
	}
	return c.commentRepository, err
}

// MustGetCommentRepository panics on error
func (c *Container) MustGetCommentRepository() *CommentRepository {
	if r, err := c.GetCommentRepository(); err != nil {
		panic(err)
	} else {
		return r
	}
}

// GetOptions returns the container's options
func (c *Container) GetOptions() ContainerOptions {
	return c.ContainerOptions
}

// GetLogger gets a logger
func (c *Container) GetLogger() (LoggerInterface, error) {
	if c.logger == nil {
		if c.ContainerOptions.LoggerFactory != nil {
			logger, err := c.ContainerOptions.LoggerFactory()
			if err != nil {
				return nil, err
			}
			c.logger = logger
		} else {
			logger := &log.Logger{}
			logger.SetOutput(os.Stdout)
			if c.ContainerOptions.Debug == true {
				c.logger = NewDefaultLogger(ALL)
			} else {
				c.logger = NewDefaultLogger(c.ContainerOptions.LogLevel)
			}

		}
	}
	return c.logger, nil
}

// MustGetLogger panics on error or return a LoggerInterface
func (c *Container) MustGetLogger() LoggerInterface {
	logger, err := c.GetLogger()
	if err != nil {
		panic(err)
	}
	return logger
}

// HTTPRedirect redirects a request
func (c *Container) HTTPRedirect(url string, status int) {
	if session, err := c.GetSession(); err == nil {
		session.Save(c.Request(), c.ResponseWriter())
	} else {
		c.MustGetLogger().Error("Container", err)
	}
	http.Redirect(c.ResponseWriter(), c.Request(), url, status)
}

// HTTPError writes an error to the response
func (c *Container) HTTPError(rw http.ResponseWriter, r *http.Request, status int, message Any) {
	c.MustGetLogger().Error(fmt.Sprintf("%s %d %s", r.URL, status, message))
	rw.WriteHeader(status)
	// if debug show a detailed error message
	if c.ContainerOptions.Debug == true {
		// if response has been sent, just write to output for now
		// TODO buffer response in order to handle the case where there is
		// 		an error in the template which should lead to a status 500
		if rw.(ResponseWriterExtra).IsResponseWritten() {
			http.Error(rw, fmt.Sprintf("%v", message), status)
			return
		}
		// if not then execute the template with the Message
		c.MustGetTemplate().ExecuteTemplate(rw, "error.tpl.html", map[string]interface{}{"Error": struct {
			Status  int
			Message interface{}
		}{Status: status, Message: message}})
		return
	}
	// if not debug show a generic error message.
	// don't show a detailed error message
	if rw.(ResponseWriterExtra).IsResponseWritten() {
		http.Error(rw, http.StatusText(status), status)
		return
	}
	c.MustGetTemplate().ExecuteTemplate(rw, "error.tpl.html", map[string]interface{}{"Error": struct {
		Status  int
		Message string
	}{Status: status, Message: http.StatusText(status)}})
}

// GetSessionStore returns a session.Store
func (c *Container) GetSessionStore() (sessions.Store, error) {
	if c.ContainerOptions.Session.StoreFactory == nil {
		return nil, errors.New("SessionStoreFactory not defined in Container.Options")
	}
	if c.sessionStore == nil {
		var err error
		c.sessionStore, err = c.ContainerOptions.Session.StoreFactory()
		if err != nil {
			return nil, err
		}
	}
	return c.sessionStore, nil
}
