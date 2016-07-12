package gonews_test

import (
	"fmt"
	"strings"
	"testing"

	"database/sql"
	"net/http"

	gonews "github.com/mparaiso/gonews/internal"
)

/*
	Scenario:

	Given a submission form validator

	If a valid submission form is validated
	It should return no errors
*/
func TestSubmissionFormValidator(t *testing.T) {
	// Given a submission form validator
	req, err := http.NewRequest("POST", "http://foo.com/submit", strings.NewReader("request body"))
	if err != nil {
		t.Fatal(err)
	}
	validator := &gonews.SubmissionFormValidator{CSRFGenerator: TestCSRFProvider{}, Request: req}
	submissionForm := &gonews.SubmissionForm{Name: "Submission form", CSRF: TestCSRFProvider{}.Generate("", ""), Title: "The Title", URL: "foo.bar.com"}
	// If a valid submission form is validated
	err = validator.Validate(submissionForm)
	// It should return no errors
	if err != nil {
		t.Fatal(err)
	}
}

func ExampleIsURL() {
	for _, url := range []string{
		"foo.com",
		"at.baz.co.uk/foo.com/?&bar=booo",
		"http://baz.com/bar?id=bizz",
		"http://presentation.opex.com/index.html?foobar=biz#baz",
	} {
		fmt.Println(gonews.IsURL(url))

	}

	for _, url := range []string{
		"foo",
		"biz/baz",
		"something.com/ with space",
	} {
		fmt.Println(gonews.IsURL(url))
	}

	// Output:
	// true
	// true
	// true
	// true
	// false
	// false
	// false
}

func Test_UserValidator_Validate_valid_user(t *testing.T) {
	user := &gonews.User{Username: "Bill_Doe", Email: "bill.doe@acme.com"}
	user.MustCreateSecurePassword("password")
	userValidator := &gonews.UserValidator{}
	if err := userValidator.Validate(user); err != nil {
		t.Fatalf("user should be valid, got '%v'", err)
	}
}

func Test_UserValidator_Validate_invalid_user(t *testing.T) {
	user := &gonews.User{Username: "Bill_Doe", Email: "bill.doe@acme"}
	user.MustCreateSecurePassword("password")
	userValidator := &gonews.UserValidator{}
	if err := userValidator.Validate(user); err == nil {
		t.Fatalf("user should be invalid, got '%v'", err)
	}
}

func Test_RegistrationFormValidator_valid_registrationForm(t *testing.T) {
	form := &gonews.RegistrationForm{CSRF: "csrf-token", Username: "johnny_doe", Password: "password", PasswordConfirmation: "password", Email: "johnny_doe@acme.com"}
	r := new(http.Request)
	r.RemoteAddr = "some-addr"
	validator := gonews.NewRegistrationFormValidator(r, TestCSRFProvider{}, TestUserFinder{})
	errors := validator.Validate(form)
	if errors != nil {
		t.Fatal("There should be no error got : ", errors)
	}
}

// CSRFProvider provide csrf tokens
type TestCSRFProvider struct{}

func (TestCSRFProvider) Generate(userID, actionID string) string {
	return "csrf-token"
}
func (TestCSRFProvider) Valid(token, userID, actionID string) bool {
	return token == "csrf-token"
}

type TestUserFinder struct{}

func (TestUserFinder) GetOneByEmail(email string) (*gonews.User, error) {
	if email == "john_doe@acme.com" {
		return &gonews.User{}, nil
	}
	return nil, sql.ErrNoRows
}
func (TestUserFinder) GetOneByUsername(username string) (*gonews.User, error) {
	if username == "john_doe" {
		return &gonews.User{}, nil
	}
	return nil, sql.ErrNoRows
}
