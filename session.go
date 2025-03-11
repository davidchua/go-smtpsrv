package smtpsrv

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
)

// A Session is returned after successful login.
type Session struct {
	connState *smtp.ConnectionState
	From      *mail.Address
	EmailData bytes.Buffer
	To        []*mail.Address
	handler   HandlerFunc
	body      io.Reader
	username  *string
	password  *string
}

// NewSession initialize a new session
func NewSession(state *smtp.ConnectionState, handler HandlerFunc, username, password *string) *Session {
	return &Session{
		connState: state,
		handler:   handler,
	}
}

func (s *Session) Mail(from string, opts smtp.MailOptions) (err error) {
	s.From, err = mail.ParseAddress(from)
	return
}

func (s *Session) Rcpt(to string) (err error) {
	toAddr, err := mail.ParseAddress(to)
	s.To = append(s.To, toAddr)
	return
}

func (s *Session) Data(r io.Reader) error {
	if s.handler == nil {
		return errors.New("internal error: no handler")
	}

	s.EmailData.Reset()
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)
	_, err := io.Copy(&s.EmailData, &buf)
	if err != nil {
		return err
	}
	s.body = tee

	c := Context{
		session: s,
	}

	return s.handler(&c)
}

func (s *Session) Reset() {
}

func (s *Session) Logout() error {
	return nil
}

func (s *Session) FormatRFC822() string {
	return fmt.Sprintf(
		"Date: %s\r\nFrom: %s\r\nTo: %s\r\n%s",
		time.Now().Format(time.RFC1123Z),
		s.From,
		s.formatRecipients(),
		s.EmailData.String(),
	)
}

func (s *Session) formatRecipients() string {
	toList := []string{}
	for _, v := range s.To {
		toList = append(toList, fmt.Sprintf("%s", v))

	}
	return fmt.Sprintf("%s", strings.Join(toList, ", "))
}
