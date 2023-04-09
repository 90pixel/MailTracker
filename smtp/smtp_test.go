package smtp

import (
	"io"
	"reflect"
	"testing"

	"github.com/emersion/go-smtp"
	"github.com/kylegrantlucas/discord-smtp-server/discord"
)

func TestNewBackend(t *testing.T) {
	type args struct {
		discordToken string
		username     string
		password     string
	}
	tests := []struct {
		name    string
		args    args
		want    *Backend
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBackend(tt.args.discordToken, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBackend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBackend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackend_Login(t *testing.T) {
	type fields struct {
		discordClient *discord.Client
		username      string
		password      string
	}
	type args struct {
		state    *smtp.ConnectionState
		username string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    smtp.Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				discordClient: tt.fields.discordClient,
				username:      tt.fields.username,
				password:      tt.fields.password,
			}
			got, err := b.Login(tt.args.state, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Backend.Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Backend.Login() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackend_AnonymousLogin(t *testing.T) {
	type fields struct {
		discordClient *discord.Client
		username      string
		password      string
	}
	type args struct {
		state *smtp.ConnectionState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    smtp.Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				discordClient: tt.fields.discordClient,
				username:      tt.fields.username,
				password:      tt.fields.password,
			}
			got, err := b.AnonymousLogin(tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("Backend.AnonymousLogin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Backend.AnonymousLogin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_Mail(t *testing.T) {
	type fields struct {
		backend *Backend
		webhook string
		from    string
	}
	type args struct {
		from string
		opts smtp.MailOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				backend: tt.fields.backend,
				webhook: tt.fields.webhook,
				from:    tt.fields.from,
			}
			if err := s.Mail(tt.args.from, tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("Session.Mail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSession_Rcpt(t *testing.T) {
	type fields struct {
		backend *Backend
		webhook string
		from    string
	}
	type args struct {
		to string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				backend: tt.fields.backend,
				webhook: tt.fields.webhook,
				from:    tt.fields.from,
			}
			if err := s.Rcpt(tt.args.to); (err != nil) != tt.wantErr {
				t.Errorf("Session.Rcpt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSession_Data(t *testing.T) {
	type fields struct {
		backend *Backend
		webhook string
		from    string
	}
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				backend: tt.fields.backend,
				webhook: tt.fields.webhook,
				from:    tt.fields.from,
			}
			if err := s.Data(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("Session.Data() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSession_Reset(t *testing.T) {
	type fields struct {
		backend *Backend
		webhook string
		from    string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				backend: tt.fields.backend,
				webhook: tt.fields.webhook,
				from:    tt.fields.from,
			}
			s.Reset()
		})
	}
}

func TestSession_Logout(t *testing.T) {
	type fields struct {
		backend *Backend
		webhook string
		from    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				backend: tt.fields.backend,
				webhook: tt.fields.webhook,
				from:    tt.fields.from,
			}
			if err := s.Logout(); (err != nil) != tt.wantErr {
				t.Errorf("Session.Logout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
