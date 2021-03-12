package main

// Landing page (ledger.html) on local server with form fields.
// Reference:
// https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/04.1.html

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"encoding/json"


	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/joho/godotenv/autoload"
	bc "golang.org/x/crypto/bcrypt"
	"msudenver.edu/ledger/db"
	"msudenver.edu/ledger/repos"
    "os"

    "github.com/plaid/plaid-go/plaid"
    //"github.com/gin-gonic/gin"


)
type Item struct {
	AvailableProducts     []string  `json:"available_products"`
	BilledProducts        []string  `json:"billed_products"`
	Error                 error     `json:"error"`
	InstitutionID         string    `json:"institution_id"`
	ItemID                string    `json:"item_id"`
	Webhook               string    `json:"webhook"`
	ConsentExpirationTime time.Time `json:"consent_expiration_time"`
}
var (
	PLAID_CLIENT_ID     = os.Getenv("PLAID_CLIENT_ID")
	PLAID_SECRET        = os.Getenv("PLAID_SECRET")
	PLAID_PRODUCTS      = os.Getenv("PLAID_PRODUCTS")
	PLAID_COUNTRY_CODES = os.Getenv("PLAID_COUNTRY_CODES")
)
var accessToken string
var itemID string
// UserInfo form fields.
type UserInfo struct {
	Email    string
	Name     string `json:"name"`
	Password string `json:"password"`
}
type Access struct {
	Token string `json:"public_token"`
}

var repo *repos.Repo

// Temp env var expires on session close
var jwtEnv = os.Getenv("jwt")
var signKey = []byte("")

// User pw byte slice
var bcryptPW = []byte("")

var clientOptions = plaid.ClientOptions{
    os.Getenv("PLAID_CLIENT_ID"),
    os.Getenv("PLAID_SECRET"),
    plaid.Sandbox, // Available environments are Sandbox, Development, and Production
    &http.Client{}, // This parameter is optional
}

var client, err = plaid.NewClient(clientOptions)

func main() {
	//r := gin.Default()
	//r.POST("/create_link_token", create_link_token)
	//r.GET("/get_access_token", get_access_token)
	//r.Run()

	database := db.Init()
	repo = repos.CreateRepo(database)
	if err := repo.CreateSchema(database); err != nil {
		log.Fatal("Unable to create schemas", err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// *** Stuck on displaying info in current page ***
	http.HandleFunc("/welcome", registerBtn)
	http.HandleFunc("/create_link_token", create_link_token)
	http.HandleFunc("/get_access_token", getAccessToken)


	log.Println("Listening on port :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	http.ListenAndServe(":8080", nil)
}

func registerBtn(w http.ResponseWriter, r *http.Request) {

	// create_link_token()

	details := UserInfo{
		Email:    r.FormValue("email"),
		Name:     r.FormValue("name"),
		Password: r.FormValue("password"),
	}

	// Hash user password. 
	encryptPassword(details.Password)

	// *** TODO: Hash pw added to CreateUser/db ****
	user, err := repo.Users.CreateUser(details.Name, details.Email)
	if err != nil {
		panic(err)
	}

	str := fmt.Sprintf(
		"Welcome to Ledger! user: %s, email: %s, ID: %d, Registered on: %s",
		user.FullName,
		user.Email,
		user.ID,
		user.CreatedAt,
	)
	// Display user entered name & email in browser.
	fmt.Fprintf(w, str)

	tokenString, err := GenerateJWT(user)
	if err != nil {
		fmt.Println("Error creating JWT.")
	}
	// Out jwt
	fmt.Println("\nJWT: " + tokenString)

}

/*
 encryptPassword() & confirmPassword() funcs
 Ref: https://stackoverflow.com/questions/23259586/bcrypt-password-hashing-in-golang-compatible-with-node-js
*/
func encryptPassword(password string) string {
	// Byte slice of User password to use bcrypt.
	bcryptPW = []byte(password)
	// Hashing the password with the default cost of 10
	hashedPassword, err := bc.GenerateFromPassword(bcryptPW, bc.DefaultCost)
	if err != nil {
		panic(err)
	}

	// Out hashed pw
	fmt.Println("HASH'D PW: " + string(hashedPassword) + "\n")

	// Test hash validation, nil == match
	confirmPassword(hashedPassword)

	return string(hashedPassword)
}

// Compare password to db pw hash record (login pw validation).
func confirmPassword(hash []byte) {
	err := bc.CompareHashAndPassword(hash, bcryptPW)
	fmt.Print("Confirm PW (nil if match): ")
	fmt.Println(err) 
}

// GenerateJWT ...
func GenerateJWT(usr *repos.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = usr.ID
	claims["user"] = usr.FullName
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Minute * 60).Unix()

	// Nav to: https://jwt.io/  paste tokenString in text field.
	var signKey = []byte(jwtEnv)

	tokenString, err := token.SignedString(signKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func create_link_token(w http.ResponseWriter, r *http.Request) {

	// Create a link_token for the given user
	linkTokenResp, err := client.CreateLinkToken(plaid.LinkTokenConfigs{
		User: &plaid.LinkTokenUser{
		  ClientUserID:             "123-test-user-id",
		},
		ClientName:            "My App",
		Products:              []string{"auth", "transactions"},
		CountryCodes:          []string{"US"},
		Language:              "en",
		Webhook:               "https://webhook-uri.com",
		LinkCustomizationName: "default",
		AccountFilters: &map[string]map[string][]string{
		  "depository": map[string][]string{
			"account_subtypes": []string{"checking", "savings"},
		  },
		},
	  })
	linkToken := linkTokenResp.LinkToken
	if err != nil {
		panic(err)
	}

	println(linkToken)
	fmt.Fprint(w, linkToken)

	// Send the data to the client
	//c.JSON(http.StatusOK, gin.H{
	//  "link_token": linkToken,
	//})
}

func getAccessToken(w http.ResponseWriter, r *http.Request) {

	var req Access

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	response, err := client.ExchangePublicToken(req.Token)
	accessToken = response.AccessToken
	itemID = response.ItemID

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	fmt.Println("public token: " + req.Token)
	fmt.Println("access token: " + accessToken)
	fmt.Println("item ID: " + itemID)

	fmt.Fprint(w, accessToken)
	fmt.Fprint(w, itemID)

	// Check if this item already exists
	// GetItem retrieves an item associated with an access token.
	// See https://plaid.com/docs/api/items/#itemget.
	// itemResp, errrrr := client.GetItem(accessToken)
	// item := itemResp.Item
	// status := itemResp.Status
	// if errrrr != nil {
	// 	http.Error(w, errrrr.Error(), 400)
	// 	return
	// }

	//fmt.Println("Item: %s" + item.Products)
	//fmt.Println("Status: %s" + status["transactions"])

	// Endpoint: /accounts/get
	// GetAccounts retrieves accounts associated with an Item.
	// See https://plaid.com/docs/api/accounts/.
	responsee, err := client.GetAccounts(accessToken)
	fmt.Println(responsee.Accounts)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
}