package main

// Landing page (ledger.html) on local server with form fields.
// Reference:
// https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/04.1.html

import (
	"fmt"
	// "html/template"
	"log"
	"net/http"
)

// User form fields.
type UserInfo struct {
	Email    string
	Name     string
	Password string
}

func main() {

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// *** Stuck on displaying info in current page ***
	http.HandleFunc("/welcome", registerBtn)

	log.Println("Listening on port :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	http.ListenAndServe(":8080", nil)
}

func registerBtn(w http.ResponseWriter, r *http.Request) {
	// Struct fields accessible by . operator.
	details := UserInfo{
		Email:    r.FormValue("email"),
		Name:     r.FormValue("name"),
		Password: r.FormValue("password"),
	}

	str := "Welcome to Ledger, user: " + details.Name +
		", email: " + details.Email

	// Display user entered name & email in browser.
	fmt.Fprintf(w, str)

	// fmt.Println("password:", r.Form["password"])
}
