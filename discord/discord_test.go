package discord

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestNewClient(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetGuildID(t *testing.T) {
	type fields struct {
		client *discordgo.Session
	}
	type args struct {
		guildName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				client: tt.fields.client,
			}
			got, err := c.GetGuildID(tt.args.guildName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetGuildID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Client.GetGuildID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetChannelID(t *testing.T) {
	type fields struct {
		client *discordgo.Session
	}
	type args struct {
		guildID     string
		channelName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				client: tt.fields.client,
			}
			got, err := c.GetChannelID(tt.args.guildID, tt.args.channelName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetChannelID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Client.GetChannelID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetWebhook(t *testing.T) {
	type fields struct {
		client *discordgo.Session
	}
	type args struct {
		username  string
		channelID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				client: tt.fields.client,
			}
			got, err := c.GetWebhook(tt.args.username, tt.args.channelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Client.GetWebhook() = %v, want %v", got, tt.want)
			}
		})
	}
}
