package main

import _ "discord-smtp-server/tzinit"

import (
	"discord-smtp-server/smtp"
	gosmtp "github.com/emersion/go-smtp"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	backend, err := smtp.NewBackend(
		os.Getenv("MONGO_URI"),
		os.Getenv("DISCORD_WEBHOOK"),
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
	)
	if err != nil {
		log.Fatal(err)
	}

	server := gosmtp.NewServer(backend)

	port := ":1025"
	if os.Getenv("SMTP_PORT") != "" {
		port = ":" + os.Getenv("SMTP_PORT")
	}

	host := "localhost"
	if os.Getenv("HOST") != "" {
		host = os.Getenv("HOST")
	}

	server.Addr = port
	server.Domain = host
	server.ReadTimeout = 10 * time.Second
	server.WriteTimeout = 10 * time.Second
	server.MaxMessageBytes = 1024 * 1024
	server.MaxRecipients = 50
	server.AllowInsecureAuth = true

	log.Println("Starting server at", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
