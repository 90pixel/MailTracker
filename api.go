package main

import (
	_ "discord-smtp-server/tzinit"
	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/cors"
)

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	time "time"
)

type mailDto struct {
	Id          string `json:"id"`
	Subject     string `json:"subject"`
	Data        string `json:"data"`
	To          string `json:"to"`
	IsRead      int    `json:"isread"`
	From        string `json:"from"`
	Body        string `json:"body"`
	Cc          string `json:"cc"`
	Bcc         string `json:"bcc"`
	Rcpt        string `json:"rcpt"`
	MimeVersion string `json:"mimeversion"`
	ContentType string `json:"contenttype"`
	CreatedAt   string `json:"createdat"`
}

type mailListDto struct {
	Id        string `json:"id"`
	Subject   string `json:"subject"`
	To        string `json:"to"`
	IsRead    int    `json:"isread"`
	From      string `json:"from"`
	CreatedAt string `json:"createdat"`
}

type userDto = struct {
	Id        string   `json:"id"`
	Salt      string   `json:"salt"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Role      string   `json:"role"`
	Emails    []string `json:"emails"`
	CreatedAt string   `json:"createdat"`
}

type userListDto = struct {
	Id        string   `json:"id"`
	Username  string   `json:"username"`
	Role      string   `json:"role"`
	Emails    []string `json:"emails"`
	CreatedAt string   `json:"createdat"`
}

type supportDto = struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdat"`
	IsRead    int    `json:"isread"`
	Status    string `json:"status"`
}

type supportMessageDto = struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	TicketId      string `json:"ticketid"`
	Message       string `json:"message"`
	IsReadAdmin   int    `json:"isreadadmin"`
	IsReadWatcher int    `json:"isreadwatcher"`
	CreatedAt     string `json:"createdat"`
}

// enum status for support
const (
	SupportStatusOpen       = "open"
	SupportStatusInProgress = "inprogress"
	SupportStatusClosed     = "closed"
	SupportStatusResolved   = "resolved"
)

func connection(startup bool) *mongo.Client {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Fatal(err)
	}
	if startup {
		// Check the connection
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connected to MongoDB!")
	}
	return client
}

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("bad header value given")
	}

	jwtToken := strings.Split(header, " ")
	if len(jwtToken) != 2 {
		return "", errors.New("incorrectly formatted authorization header")
	}

	return jwtToken[1], nil
}

func parseToken(jwtToken string) (*jwt.Token, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, OK := token.Method.(*jwt.SigningMethodHMAC); !OK {
			return nil, errors.New("bad signed method received")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, errors.New("bad jwt token")
	}

	return token, nil
}

func permissionPublic(c *gin.Context) {
	c.Next()
}

func permissionCheckAuth(c *gin.Context) {
	permissionCheck(c, "")
	c.Next()
}

func permissionCheckAdmin(c *gin.Context) {
	permissionCheck(c, "admin")
	c.Next()
}

func permissionCheckWatcher(c *gin.Context) {
	permissionCheck(c, "watcher")
	c.Next()
}

func permissionCheck(c *gin.Context, role string) {
	client := connection(false)
	jwtToken, err := extractBearerToken(c.GetHeader("Authorization"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "bad authorization header",
		})
		return
	}
	token, err := parseToken(jwtToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "bad jwt token",
		})
		return
	}
	salt := token.Claims.(jwt.MapClaims)["sub"]
	var user userListDto
	collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("users")
	err = collection.FindOne(context.TODO(), bson.M{"salt": salt}).Decode(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Kullanıcı bulunamadı",
		})
	}

	if role != "" && user.Role != string(role) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "user not authorized",
		})
	}

	var apiUser userListDto
	apiUser.Id = user.Id
	apiUser.Username = user.Username
	apiUser.Role = user.Role
	apiUser.Emails = user.Emails
	apiUser.CreatedAt = user.CreatedAt

	c.Set("currentUser", apiUser)
	c.Set("currentUserName", apiUser.Username)
	c.Set("currentUserRole", apiUser.Role)

	c.Next()
}

func timeoutResponse(c *gin.Context) {
	c.JSON(http.StatusRequestTimeout, gin.H{
		"message": "Request Timeout",
	})
}

func timeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(500*time.Millisecond),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(timeoutResponse),
	)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}
	log.Default().SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting server at", os.Getenv("HOST")+":"+os.Getenv("PORT"))
	log.Println("Sentry DSN: " + os.Getenv("SENTRY_DSN"))
	err = raven.SetDSN(os.Getenv("SENTRY_DSN"))
	if err != nil {
		log.Fatal("Error setting sentry dsn")
		return
	}

	http.DefaultClient.Timeout = time.Minute * 10
	client := connection(true)
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/assets", "./assets")
	router.Use(cors.Default())
	router.Use(timeoutMiddleware())
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

	router.GET("/iframe/mails/:id", func(c *gin.Context) {
		objID, _ := primitive.ObjectIDFromHex(c.Param("id"))
		var mail mailDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		err := collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&mail)
		if err != nil {
			log.Fatal(err)
			return
		}
		html := ""

		data := regexp.MustCompile(`(?s)<body.*?>(.*?)</body>`).FindStringSubmatch(mail.Body)
		if len(data) > 0 {
			html = data[1]
		}

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

	permissionMailRouter := router.Group("/")
	permissionMailRouter.Use(permissionCheckAuth)
	permissionMailRouter.GET("/api/mails", func(c *gin.Context) {

		var mails []mailListDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		// order by date desc
		opts := options.Find()
		opts.SetSort(bson.D{{"_id", -1}})

		// search by subject
		payload := bson.D{}
		if c.Query("subject") != "" {
			payload = bson.D{{"subject", bson.D{{"$regex", c.Query("subject")}, {"$options", "i"}}}}
		}

		// search by from
		user, userErr := c.Get("currentUser")
		if !userErr {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}
		if user.(userListDto).Role == "watcher" {

			payload = bson.D{
				{"from", bson.D{{"$in", user.(userListDto).Emails}}},
				{"subject", bson.D{{"$regex", c.Query("subject")}, {"$options", "i"}}},
			}
		}

		cur, err := collection.Find(context.TODO(), payload, opts)
		if err != nil {
			log.Fatal(err)
			return
		}
		// has empty result empty array
		if cur.RemainingBatchLength() == 0 {
			c.JSON(http.StatusOK, gin.H{
				"data": mails,
			})
			return
		}

		for cur.Next(context.TODO()) {
			var mail mailListDto

			mail.Id = cur.Current.Lookup("_id").ObjectID().Hex()
			mail.From = cur.Current.Lookup("from").StringValue()
			mail.To = cur.Current.Lookup("to").StringValue()
			mail.IsRead = int(cur.Current.Lookup("isread").Int32())
			mail.Subject = cur.Current.Lookup("subject").StringValue()
			mail.CreatedAt = cur.Current.Lookup("createdat").StringValue()
			mails = append(mails, mail)
		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
			return
		}
		err = cur.Close(context.TODO())
		if err != nil {
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": mails,
		})
	})
	// get mail from iframe

	permissionMailRouter.GET("/api/mails/:id", func(c *gin.Context) {
		objID, _ := primitive.ObjectIDFromHex(c.Param("id"))

		var mail mailDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		err := collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&mail)
		if err != nil {
			log.Fatal(err)
			return
		}
		mail.Id = objID.Hex()
		_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.D{
			{"$set", bson.D{
				{"isread", 1},
			},
			},
		})
		if err != nil {
			log.Fatal(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": mail,
		})
	})
	permissionAdminMailRouter := router.Group("/")
	permissionAdminMailRouter.Use(permissionCheckAdmin)
	permissionAdminMailRouter.DELETE("/api/mails/:id", func(c *gin.Context) {
		objID, _ := primitive.ObjectIDFromHex(c.Param("id"))
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
		if err != nil {
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Mail deleted",
		})
	})
	// delete all
	permissionAdminMailRouter.DELETE("/api/mails", func(c *gin.Context) {
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		_, err := collection.DeleteMany(context.TODO(), bson.M{})
		if err != nil {
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "All mails deleted",
		})
	})
	// read all
	permissionAdminMailRouter.PUT("/api/mails", func(c *gin.Context) {
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("mails")
		_, err := collection.UpdateMany(context.TODO(), bson.M{}, bson.D{
			{"$set", bson.D{
				{"isread", 1},
			},
			},
		})
		if err != nil {
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "All mails read",
		})
	})
	// user and login routes

	router.POST("/api/login", func(c *gin.Context) {
		var user userDto
		c.BindJSON(&user)

		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("users")
		plainPwd := user.Password
		// get username from users
		log.Println(user.Username)
		err := collection.FindOne(context.TODO(), bson.M{"username": user.Username}).Decode(&user)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Kullanıcı adı veya parola hatalı.",
				})
			return
		}

		byteHash := []byte(user.Password)
		err = bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
		if err != nil {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Kullanıcı adı veya parola hatalı. 2",
				})
			return
		}

		token := jwt.New(jwt.SigningMethodHS256)
		claims := make(jwt.MapClaims)
		claims["exp"] = time.Now().Add(time.Hour * 24 * 365).Unix()
		claims["iat"] = time.Now().Unix()
		claims["sub"] = user.Salt
		token.Claims = claims

		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			log.Fatal(err)
			return
		}

		c.JSON(
			http.StatusOK,
			gin.H{
				"data": gin.H{
					"token": tokenString,
					"user":  user,
				},
			},
		)
	})

	permissionUserWatcherRouter := router.Group("/")
	permissionUserWatcherRouter.Use(permissionCheckAuth)
	permissionUserWatcherRouter.GET("/api/users/me", func(c *gin.Context) {

		user, err := c.Get("currentUser")
		if !err {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}

		c.JSON(http.StatusOK, gin.H{
			"data": user,
		})
	})

	permissionUserAdminRouter := router.Group("/")
	permissionUserAdminRouter.Use(permissionCheckAdmin)
	permissionUserAdminRouter.GET("/api/users", func(c *gin.Context) {
		var users []userListDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("users")
		cur, err := collection.Find(context.TODO(), bson.D{})
		if err != nil {
			log.Fatal(err)
			return
		}
		for cur.Next(context.TODO()) {
			var user userListDto
			err := cur.Decode(&user)
			if err != nil {
				log.Fatal(err)
				return
			}
			user.Id = cur.Current.Lookup("_id").ObjectID().Hex()
			user.Username = cur.Current.Lookup("username").StringValue()
			user.Role = cur.Current.Lookup("role").StringValue()

			users = append(users, user)
		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
			return
		}
		err = cur.Close(context.TODO())
		if err != nil {
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": users,
		})
	})
	permissionAdminMailRouter.POST("/api/users", func(c *gin.Context) {
		var user userDto
		c.BindJSON(&user)

		// validate username length
		if len(user.Username) < 3 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Kullanıcı adı en az 3 karakter olmalıdır",
			})
			return
		}
		// Regex to validate password strength

		secure := true
		tests := []string{".{7,}", "[a-z]", "[A-Z]", "[0-9]", "[^\\d\\w]"}
		for _, test := range tests {
			t, _ := regexp.MatchString(test, user.Password)
			if !t {
				secure = false
				break
			}
		}

		if secure == false {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Parola en az 8 karakter uzunluğunda olmalıdır, en az bir büyük harf, bir küçük harf, bir sayı ve bir özel karakter içermelidir",
			})
			return
		}

		rand.Seed(time.Now().UnixNano())
		b := make([]byte, 10+2)
		rand.Read(b)
		salt := fmt.Sprintf("%x", b)[2 : 10+2]
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("users")

		var userExists userDto
		c.ShouldBindJSON(&userExists)
		err = collection.FindOne(context.TODO(), bson.M{"username": user.Username}).Decode(&userExists)
		if err != nil {
			if mongo.ErrNoDocuments != err {
				log.Fatal(err)
			}
		}
		if userExists.Username != "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Kullanıcı adı zaten kullanılıyor",
			})
		}

		hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
		if err != nil {
			log.Fatal(err)
		}

		index := []mongo.IndexModel{
			{
				Keys: bsonx.Doc{{Key: "index", Value: bsonx.String("text")}},
			},
			{
				Keys: bsonx.Doc{{Key: "date", Value: bsonx.Int32(-1)}},
			},
		}

		opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
		_, err = collection.Indexes().CreateMany(context.TODO(), index, opts)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Hata oluştu",
			})
		}

		_, err = collection.InsertOne(context.TODO(), bson.D{
			{"username", user.Username},
			{"password", string(hashPassword)},
			{"emails", user.Emails},
			{"salt", salt},
			{"role", user.Role},
			{"createdat", time.Now().UTC().String()},
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Hata oluştu",
			})
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.JSON(http.StatusOK, gin.H{
			"message": "User created",
		})
	})
	permissionUserAdminRouter.DELETE("/api/users/:id", func(c *gin.Context) {
		objID, _ := primitive.ObjectIDFromHex(c.Param("id"))
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("users")
		_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
		if err != nil {
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "User deleted",
		})
	})
	// update user
	permissionUserAdminRouter.PUT("/api/users/:id", func(c *gin.Context) {
		objID, _ := primitive.ObjectIDFromHex(c.Param("id"))
		var user userDto
		c.BindJSON(&user)
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("users")

		// validate username length
		if len(user.Username) < 3 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Kullanıcı adı en az 3 karakter olmalıdır",
			})
			return
		}

		if user.Password != "" {
			secure := true
			tests := []string{".{7,}", "[a-z]", "[A-Z]", "[0-9]", "[^\\d\\w]"}
			for _, test := range tests {
				t, _ := regexp.MatchString(test, user.Password)
				if !t {
					secure = false
					break
				}
			}
			if secure == false {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": "Parola en az 8 karakter uzunluğunda olmalıdır, en az bir büyük harf, bir küçük harf, bir sayı ve bir özel karakter içermelidir",
				})
				return
			}
		}

		// check if user exists
		var userExists userDto
		err := c.ShouldBindJSON(&userExists)
		if err != nil {
			log.Println("No extras provided")
		}
		err = collection.FindOne(context.TODO(), bson.M{"username": user.Username, "_id": bson.M{"$ne": objID}}).Decode(&userExists)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				log.Fatal(err)
			}
		}
		if userExists.Username != "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "Kullanıcı adı zaten kullanılıyor",
			})
			return
		}

		_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.D{
			{"$set", bson.D{
				{"username", user.Username},
				{"emails", user.Emails},
				{"role", user.Role},
			},
			},
		})
		if err != nil {
			log.Fatal(err)
		}

		if len(user.Password) > 0 {
			h := md5.New()
			hashPassword, err := io.WriteString(h, user.Password)
			if err != nil {
				log.Fatal(err)
				return
			}
			_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.D{
				{"$set", bson.D{
					{"password", hashPassword},
				},
				},
			})
			if err != nil {
				log.Fatal(err)
			}
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.JSON(http.StatusOK, gin.H{
			"message": "User updated",
		})
	})
	permissionUserAdminRouter.GET("/api/users/:id", func(c *gin.Context) {
		objID, _ := primitive.ObjectIDFromHex(c.Param("id"))
		var user userListDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("users")
		err := collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Kullanıcı bulunamadı",
				})
				return
			} else {
				log.Fatal(err)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"data": user,
		})
	})

	// support (ticket system)

	permissionUserAdminRouter.DELETE("/api/tickets/:id", func(c *gin.Context) {
		id := c.Param("id")
		objID, _ := primitive.ObjectIDFromHex(id)
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("supports")
		_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Support deleted",
		})
	})
	permissionUserAdminRouter.PUT("/api/tickets/:id", func(c *gin.Context) {
		id := c.Param("id")
		objID, _ := primitive.ObjectIDFromHex(id)
		var support supportDto
		c.BindJSON(&support)
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("supports")
		_, err := collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.D{
			{"$set", bson.D{
				{"status", support.Status},
			},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Support updated",
		})
	})
	permissionUserWatcherRouter.GET("/api/tickets/:id", func(c *gin.Context) {
		username, error := c.Get("currentUserName")
		if !error {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}
		role, errorRole := c.Get("currentUserRole")
		if !errorRole {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}

		objID, _ := primitive.ObjectIDFromHex(c.Param("id"))
		data := bson.M{"_id": objID}
		if role == "watcher" {
			data = bson.M{"_id": objID, "username": username}
		}

		var support supportDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("supports")
		err := collection.FindOne(context.TODO(), data).Decode(&support)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Support bulunamadı",
				})
				return
			} else {
				log.Fatal(err)
			}
		}

		if role == "admin" {
			payload := bson.D{
				{"isread", 1},
			}

			if support.Status == SupportStatusOpen {
				payload = bson.D{
					{"isread", 1},
					{"status", SupportStatusInProgress},
				}
				support.Status = SupportStatusInProgress
			}

			// isread and status update
			_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.D{
				{"$set", payload},
			})
			if err != nil {
				log.Fatal(err)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": support,
		})
	})
	permissionUserWatcherRouter.POST("/api/tickets", func(c *gin.Context) {
		username, error := c.Get("currentUserName")
		if !error {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}

		var support supportDto
		c.BindJSON(&support)
		support.Username = username.(string)
		support.IsRead = 0
		support.Status = SupportStatusOpen
		support.CreatedAt = time.Now().UTC().String()

		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("supports")
		_, err := collection.InsertOne(context.TODO(), support)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Support created",
		})
	})
	permissionUserWatcherRouter.GET("/api/tickets", func(c *gin.Context) {
		username, error := c.Get("currentUserName")
		if !error {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}
		role, errorRole := c.Get("currentUserRole")
		if !errorRole {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}

		var supports []supportDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("supports")
		payload := bson.M{"username": username.(string)}
		if role == "admin" {
			payload = bson.M{}
		}
		opts := options.Find()
		opts.SetSort(bson.D{{"_id", -1}})
		cur, err := collection.Find(context.TODO(), payload, opts)
		if err != nil {
			log.Fatal(err)
		}

		defer cur.Close(context.TODO())
		for cur.Next(context.TODO()) {
			var elem supportDto
			err := cur.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}
			elem.Id = cur.Current.Lookup("_id").ObjectID().Hex()
			elem.CreatedAt = cur.Current.Lookup("createdat").StringValue()
			supports = append(supports, elem)
		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"data": supports,
		})
	})

	permissionUserWatcherRouter.GET("/api/tickets/:id/messages", func(c *gin.Context) {
		username, error := c.Get("currentUserName")
		if !error {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}
		role, errorRole := c.Get("currentUserRole")
		if !errorRole {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}

		ticketId := c.Param("id")
		objID, _ := primitive.ObjectIDFromHex(ticketId)
		if role != "admin" {
			// check ticket owner
			data := bson.M{"_id": objID, "username": username}
			var support supportDto
			collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("supports")
			err := collection.FindOne(context.TODO(), data).Decode(&support)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusNotFound, gin.H{
						"message": "Support bulunamadı",
					})
					return
				} else {
					log.Fatal(err)
				}
			}
		}

		data := bson.M{"ticketid": ticketId}

		var supportMessages []supportMessageDto
		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("support_messages")
		cur, err := collection.Find(context.TODO(), data)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Support bulunamadı",
				})
				return
			} else {
				log.Fatal(err)
			}
		}

		defer cur.Close(context.TODO())
		for cur.Next(context.TODO()) {
			var elem supportMessageDto
			err := cur.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}
			elem.Id = cur.Current.Lookup("_id").ObjectID().Hex()
			elem.CreatedAt = cur.Current.Lookup("createdat").StringValue()
			supportMessages = append(supportMessages, elem)
		}

		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"data": supportMessages,
		})
	})
	permissionUserWatcherRouter.POST("/api/tickets/:id/messages", func(c *gin.Context) {
		username, error := c.Get("currentUserName")
		if !error {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}
		role, errorRole := c.Get("currentUserRole")
		if !errorRole {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"message": "Oturum açmadınız.",
				})
		}

		ticketId := c.Param("id")

		objID, _ := primitive.ObjectIDFromHex(ticketId)

		if role != "admin" {
			// check ticket owner
			data := bson.M{"_id": objID, "username": username}
			var support supportDto
			collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("supports")
			err := collection.FindOne(context.TODO(), data).Decode(&support)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusNotFound, gin.H{
						"message": "Support bulunamadı",
					})
					return
				} else {
					log.Fatal(err)
				}
			}
		}

		var supportMessage supportMessageDto
		c.BindJSON(&supportMessage)
		supportMessage.Username = username.(string)
		supportMessage.TicketId = ticketId
		supportMessage.IsReadWatcher = 0
		supportMessage.IsReadAdmin = 0
		supportMessage.CreatedAt = time.Now().UTC().String()

		collection := client.Database(os.Getenv("MONGO_TABLE_NAME")).Collection("support_messages")
		_, err := collection.InsertOne(context.TODO(), supportMessage)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Support message created",
		})
	})

	// Start and run the server

	router.Run()
}
