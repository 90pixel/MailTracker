package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/encoding/unicode"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

type mailDto struct {
	Id          string `json:"id"`
	Date        string `json:"date"`
	Subject     string `json:"subject"`
	Data        string `json:"data"`
	To          string `json:"to"`
	IsRead      int    `json:"isRead"`
	From        string `json:"from"`
	Body        string `json:"body"`
	Cc          string `json:"cc"`
	Bcc         string `json:"bcc"`
	Rcpt        string `json:"rcpt"`
	MimeVersion string `json:"mimeVersion"`
	ContentType string `json:"contentType"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
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

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/assets", "./assets")
	router.GET("/", func(c *gin.Context) {

		curl := "curl  --url 'smtp://" + os.Getenv("HOST") + ":" + os.Getenv("SMTP_PORT") + "' --user '" + os.Getenv("SMTP_USERNAME") + ":" + os.Getenv("SMTP_PASSWORD") + "' --mail-from " + os.Getenv("SMTP_USERNAME") + " --mail-rcpt " + os.Getenv("SMTP_USERNAME") + " --upload-file - <<EOF\n" +
			"From: My Inbox <" + os.Getenv("SMTP_USERNAME") + ">\n" +
			"To: Your Inbox <" + os.Getenv("SMTP_USERNAME") + ">\n" +
			"Subject: Test Mail\n" +
			"Content-Type: multipart/alternative; boundary=\"boundary-string\"\n" +
			"\n" +
			"--boundary-string\n" +
			"--boundary-string\n" +
			"Content-Type: text/plain; charset=\"utf-8\"\n" +
			"Content-Transfer-Encoding: quoted-printable\n" +
			"Content-Disposition: inline\n" +
			"\n" +
			"Test Mail\n" +
			"\n" +
			"Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book\n" +
			"\n" +
			"--boundary-string\n" +
			"Content-Type: text/html; charset=\"utf-8\"\n" +
			"Content-Transfer-Encoding: quoted-printable\n" +
			"Content-Disposition: inline\n" +
			"\n" +
			"<!doctype html>\n" +
			"<html>\n" +
			"<head>\n" +
			"<meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\">\n" +
			"</head>\n" +
			"<body style=\"font-family: sans-serif;\">\n" +
			"<div style=\"display: block; margin: auto; max-width: 600px;\" class=\"main\">\n" +
			"<h1 style=\"font-size: 18px; font-weight: bold; margin-top: 20px\">Test Mail</h1>\n" +
			"<p>Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book</p>\n" +
			"</div>\n" +
			"<style>\n" +
			".main { background-color: #EEE; }\n" +
			"a:hover { border-left-width: 1em; min-height: 2em; }\n" +
			"</style>\n" +
			"</body>\n" +
			"</html>\n" +
			"\n" +
			"--boundary-string--\n" +
			"EOF"

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":    "MailTracker",
			"curl":     curl,
			"host":     os.Getenv("HOST"),
			"port":     os.Getenv("SMTP_PORT"),
			"username": os.Getenv("SMTP_USERNAME"),
			"password": os.Getenv("SMTP_USERNAME"),
		})
	})
	router.GET("/api/mails", func(c *gin.Context) {
		var mails []mailDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		// order by date desc
		opts := options.Find()
		opts.SetSort(bson.D{{"_id", -1}})
		cur, err := collection.Find(context.TODO(), bson.D{}, opts)
		if err != nil {
			log.Fatal(err)
		}
		// has empty result empty array
		if cur.RemainingBatchLength() == 0 {
			c.JSON(http.StatusOK, gin.H{
				"data": mails,
			})
			return
		}

		for cur.Next(context.TODO()) {
			var mail mailDto

			err := cur.Decode(&mail)
			if err != nil {
				log.Fatal(err)
			}
			mail.Id = cur.Current.Lookup("_id").ObjectID().Hex()
			mails = append(mails, mail)
		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}
		err = cur.Close(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"data": mails,
		})
	})
	// get mail from iframe
	router.GET("/iframe/mails/:id", func(c *gin.Context) {
		objID, _ := primitive.ObjectIDFromHex(c.Param("id"))
		var mail mailDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		err := collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&mail)
		if err != nil {
			log.Fatal(err)
		}
		html := ""
		// only html tag in body

		var dec = unicode.UTF8.NewDecoder()
		body, _ := dec.String("=?UTF-8?" + mail.Body)

		data := regexp.MustCompile(`(?s)<body.*?>(.*?)</body>`).FindStringSubmatch(string(body))
		if len(data) > 0 {
			html = data[1]
		}

		// replace =\n
		html = regexp.MustCompile(`=\r`).ReplaceAllString(html, "")
		html = regexp.MustCompile(`\n`).ReplaceAllString(html, "")
		html = regexp.MustCompile(`3D\"`).ReplaceAllString(html, "\"")

		// when request query param return json
		if c.Query("json") == "true" {
			c.JSON(http.StatusOK, gin.H{
				"data": html,
			})
			return
		}

		// return template
		c.HTML(http.StatusOK, "iframe.tmpl", gin.H{
			"body": template.HTML(html),
		})
	})
	router.GET("/api/mails/:id", func(c *gin.Context) {
		objID, _ := primitive.ObjectIDFromHex(c.Param("id"))

		var mail mailDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		err := collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&mail)
		if err != nil {
			log.Fatal(err)
		}
		mail.Id = objID.Hex()
		_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.D{
			{"$set", bson.D{
				{"isRead", 1},
			},
			},
		})
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"data": mail,
		})
	})
	router.DELETE("/api/mails/:id", func(c *gin.Context) {
		objID, _ := primitive.ObjectIDFromHex(c.Param("id"))
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Mail deleted",
		})
	})
	// delete all
	router.DELETE("/api/mails", func(c *gin.Context) {
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		_, err := collection.DeleteMany(context.TODO(), bson.M{})
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "All mails deleted",
		})
	})
	// read all
	router.PUT("/api/mails", func(c *gin.Context) {
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		_, err := collection.UpdateMany(context.TODO(), bson.M{}, bson.D{
			{"$set", bson.D{
				{"isRead", 1},
			},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "All mails read",
		})
	})
	router.Run()
}
