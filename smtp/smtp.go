package smtp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emersion/go-smtp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/mail"
	"os"
	"time"
)

type Backend struct {
	client   *mongo.Client
	webhook  string
	username string
	password string
}

func NewBackend(db, discordToken, username, password string) (*Backend, error) {

	clientOptions := options.Client().ApplyURI(db)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	return &Backend{
		client:   client,
		webhook:  discordToken,
		username: username,
		password: password,
	}, nil
}

func (b *Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	if username != b.username || password != b.password {
		return nil, errors.New("Invalid username or password")
	}
	return &Session{
		backend: b,
	}, nil
}

func (b *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return nil, smtp.ErrAuthRequired
}

type Session struct {
	backend *Backend
	webhook string
	from    string
}

func (s *Session) Mail(from string, opts smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *Session) Rcpt(to string) error {
	//address, err := email.Parse(to)
	//if err != nil {
	//	return err
	//}
	//
	//guildID, err := s.backend.discordClient.GetGuildID(address.TLD)
	//if err != nil {
	//	return err
	//}
	//
	//channelID, err := s.backend.discordClient.GetChannelID(*guildID, address.Domain)
	//if err != nil {
	//	return err
	//}
	//
	//webhook, err := s.backend.discordClient.GetWebhook(address.User, *channelID)
	//if err != nil {
	//	return err
	//}

	s.webhook = s.backend.webhook

	return nil
}

type mailDto struct {
	Date    string `json:"date"`
	Subject string `json:"subject"`
	Data    string `json:"data"`
	To      string `json:"to"`
	IsRead  int    `json:"isRead"`
	From    string `json:"from"`
	Body    string `json:"body"`
	Cc      string `json:"cc"`
	Bcc     string `json:"bcc"`
}

func (s *Session) Data(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	reqBody, err := json.Marshal(
		map[string]string{
			"content": string(b),
		},
	)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(b)
	msg, err := mail.ReadMessage(reader)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(msg.Body)
	if err != nil {
		log.Fatal(err)
	}

	var dec mime.WordDecoder

	subject, err := dec.DecodeHeader(msg.Header.Get("Subject"))
	if err != nil {
		return err
	}

	to, err := dec.DecodeHeader(msg.Header.Get("To"))
	if err != nil {
		return err
	}

	from, err := dec.DecodeHeader(msg.Header.Get("From"))
	if err != nil {
		return err
	}

	cc, err := dec.DecodeHeader(msg.Header.Get("Cc"))
	if err != nil {
		return err
	}

	bcc, err := dec.DecodeHeader(msg.Header.Get("Bcc"))
	if err != nil {
		return err
	}

	var newMail mailDto
	newMail.Data = string(b)
	newMail.Subject = subject
	newMail.To = to
	newMail.From = from
	newMail.Body = string(body)
	newMail.Cc = cc
	newMail.Bcc = bcc
	newMail.Date = time.Now().Format("01/02/2006 15:04:05")
	newMail.IsRead = 0
	var mailCollection = s.backend.client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
	insert, err := mailCollection.InsertOne(context.TODO(), newMail)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(insert.InsertedID)

	resp, err := http.Post(
		s.webhook,
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}
