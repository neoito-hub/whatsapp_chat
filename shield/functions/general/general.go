package general

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kurrik/oauth1a"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/go-playground/validator.v9"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

var (
	Oauth2Config *oauth2.Config

	Envs map[string]string

	Oauth1aService *oauth1a.Service
)

func init() {

	var err error

	Envs, err = godotenv.Read(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	Oauth2Config = &oauth2.Config{
		ClientID:     Envs["SHIELD_GOOGLE_CLIENT_ID"],
		ClientSecret: Envs["SHIELD_GOOGLE_CLIENT_SECRET"],
		RedirectURL:  Envs["SHIELD_GOOGLE_REDIRECT"],
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	twittercreds := &TwitterCreds{}
	flag.StringVar(&twittercreds.Key, "key", Envs["SHIELD_TWITTER_API_KEY"], "Consumer key of twitter app")
	flag.StringVar(&twittercreds.Sec, "secret", Envs["SHIELD_TWITTER_API_SECRET"], "Consumer secret of twitter app")
	flag.StringVar(&twittercreds.Callback, "callback", Envs["SHIELD_TWITTER_CALLBACK"], "Callback url of twitter app")
	flag.Parse()

	Oauth1aService = &oauth1a.Service{
		RequestURL:   "https://api.twitter.com/oauth/request_token",
		AuthorizeURL: "https://api.twitter.com/oauth/authenticate",
		AccessURL:    "https://api.twitter.com/oauth/access_token",
		ClientConfig: &oauth1a.ClientConfig{
			ConsumerKey:    twittercreds.Key,
			ConsumerSecret: twittercreds.Sec,
			CallbackURL:    twittercreds.Callback,
		},
		Signer: new(oauth1a.HmacSha1Signer),
	}

}

func OutputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err)
		return
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err)
		return
	}
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}

func OutputHTMLTemplate(w http.ResponseWriter, filename string, message string) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// error := DisplayError{ErrorMessage: message}

	// if err := tmpl.Execute(w, error); err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }

	type DisplayError struct {
		ErrorMessage string
	}

	error := DisplayError{message}

	if err := tmpl.Execute(w, error); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// swagger:response RespondHandler
func RespondHandler(w http.ResponseWriter, success bool, status int, data interface{}) {

	message := "Error"
	if success {
		message = "Success"
	}

	response, err := json.Marshal(ResponseTemplate{
		Success: success,
		Message: message,
		Status:  status,
		Data:    data,
	})

	if err != nil {
		log.Println("error in response json marshal")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)

	log.Println("response:", string(response))
}

// func RespondHTMLHandler(w http.ResponseWriter, filename string) {
// 	message := "Error"
// 	if success {
// 		message = "Success"
// 	}

// 	// response, _ := json.Marshal(ResponseHTMLTemplate{
// 	// 	Success: success,
// 	// 	Message: message,
// 	// 	Status:  status,
// 	// 	Html:    data,
// 	// })

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(status)
// 	w.Write(response)
// }

// ValidateStruct is validating api request
func ValidateStruct(NewStruct interface{}) error {
	validate = validator.New()
	err := validate.Struct(NewStruct)
	if err != nil {

		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return err
		}

		for _, err := range err.(validator.ValidationErrors) {
			fmt.Println(err.StructField())
			fmt.Println(err.ActualTag())
			fmt.Println(err.Kind())
			fmt.Println(err.Type())
			fmt.Println(err.Value())
			fmt.Println(err.Param())
			fmt.Println()
		}

		return err
	}

	return nil
}

func GetRandomCode(maxlen int) string {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, maxlen)
	n, err := io.ReadAtLeast(rand.Reader, b, maxlen)
	if n != maxlen {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

// RandToken generates a random @l length token.
func RandToken(l int) (string, string, error) {
	b := make([]byte, l)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	return string(b), base64.StdEncoding.EncodeToString(b), nil
}

func RandPassword(l int) string {
	mathrand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	specials := "~=+%^*/()[]{}/!@#$?|"
	caps := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	all := "abcdefghijklmnopqrstuvwxyz" + digits + specials + caps

	buf := make([]byte, l)
	buf[0] = digits[mathrand.Intn(len(digits))]
	buf[1] = specials[mathrand.Intn(len(specials))]
	buf[2] = caps[mathrand.Intn(len(caps))]
	for i := 3; i < l; i++ {
		buf[i] = all[mathrand.Intn(len(all))]
	}
	mathrand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	str := string(buf)
	log.Println(str)

	return str
}

// HashPassword salt and hash the password using the bcrypt algorithm with default cost.
func HashPassword(pwd string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(bytes), err
}

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func IsValidRedirectUrl(str string) bool {
	regex := `((http|https)://)([A-Za-z0-9-._~!$&'()*+,;=:@//?]{2,256})`
	match, err := regexp.MatchString(regex, str)

	return err == nil && match

}

func FetchErrorFormCode(error_code string) string {
	switch error_code {
	case "":
		return ""
	case "50001":
		return "Something went wrong"
	case "50002":
		return "Google authorization was failed. Please try again or contact support for help."
	case "40001":
		return "User already exists, please try to login."
	case "40002":
		return "User does not exists, please try to signup."
	case "40003":
		return "Not a verified google account."
	case "40004":
		return "Not a valid client id."
	case "50003":
		return "Failed to fetch app details from db."
	case "40005":
		return "Given redirect uri not registered for this app."
	case "50004":
		return "Failed to generate token."
	case "50005":
		return "Twitter authorization was failed. Please try again or contact support for help."
	case "50006":
		return "Failed to fetch email from Twitter."
	case "50007":
		return "Linkedin authorization was failed. Please try again or contact support for help."
	case "50008":
		return "Failed to fetch email from Linkedin."
	default:
		return ""
	}
}
