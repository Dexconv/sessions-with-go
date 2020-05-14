package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"time"
)

var tpl *template.Template
type user struct {
	Email string
	Password []byte
	Fname string
	Lname string
	Role string
}

type session struct {
	un string
	lastActivity time.Time
}

var dbUsers = map[string]user{}
var dbSession = map[string]session{}
var dbSessionCleaned time.Time

const sessionlength  = 30

func init(){
	tpl = template.Must(template.ParseGlob("templates/*"))
	dbSessionCleaned = time.Now()
}

func main(){
	http.HandleFunc("/", index)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/bar", bar)
	http.HandleFunc("/logout", logout)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)
}
func index(w http.ResponseWriter, r *http.Request){
	c := getCookie(w,r)
	fmt.Println(c)
	usr := dbUsers[dbSession[c.Value].un]
	tpl.ExecuteTemplate(w,"index.gohtml",usr)
}

func signup(w http.ResponseWriter, r *http.Request) {
	if alreadyLoggedIn(r){
		http.Redirect(w,r,"/",http.StatusSeeOther)
	}
	c := getCookie(w,r)
	if r.Method == http.MethodPost{
		var usr user
		usr.Email = r.FormValue("username")
		p := r.FormValue("password")
		usr.Fname = r.FormValue("firstname")
		usr.Lname = r.FormValue("lastname")
		usr.Role = r.FormValue("role")

		if _, ok := dbUsers[usr.Email]; ok{
			http.Redirect(w,r,"signup",http.StatusSeeOther)
			return
		}
		var err error
		usr.Password, err = bcrypt.GenerateFromPassword([]byte(p),bcrypt.MinCost)
		if err != nil{
			http.Error(w,http.StatusText(500),http.StatusInternalServerError)
			return
		}

		dbUsers[usr.Email] = usr
		dbSession[c.Value] = session{usr.Email,time.Now()}
		http.Redirect(w,r,"/",http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w,"signup.gohtml",nil)
}
func bar(w http.ResponseWriter, r *http.Request){
	if !alreadyLoggedIn(r){
		http.Redirect(w,r,"/",http.StatusSeeOther)
		return
	}
	c := getCookie(w,r)
	usr , ok := dbSession[c.Value]
	if !ok{
		http.Redirect(w , r, "/", http.StatusSeeOther)
		return
	}
	u := dbUsers[usr.un]
	if dbUsers[usr.un].Role == "007"{
		tpl.ExecuteTemplate(w, "bar.gohtml", u)
		return
	}
	http.Error(w,"you're not 007",http.StatusSeeOther)
}
func login(w http.ResponseWriter, r *http.Request){
	if alreadyLoggedIn(r){
		http.Redirect(w,r,"/",http.StatusSeeOther)
	}
	c := getCookie(w,r)
	if r.Method == http.MethodPost{
		un := r.FormValue("username")
		p := r.FormValue("password")

		if _,ok := dbUsers[un];!ok{
			http.Error(w,"user doesnt exist", http.StatusForbidden)
			return
		}
		err := bcrypt.CompareHashAndPassword(dbUsers[un].Password,[]byte(p))
		if err != nil{
			http.Error(w,"password does not match", http.StatusForbidden)
			return
		}
		dbSession[c.Value] = session{un,time.Now()}
	}
	tpl.ExecuteTemplate(w,"login.gohtml",nil)
}

func logout(w http.ResponseWriter, r *http.Request){
	if !alreadyLoggedIn(r){
		http.Redirect(w,r,"/login",http.StatusSeeOther)
		return
	}
	c, _:= r.Cookie("session")
	delete(dbSession,c.Value)
	c = &http.Cookie{
		Name:       "session",
		Value:		"",
		MaxAge:     -1,
	}
	http.SetCookie(w,c)

	if time.Now().Sub(dbSessionCleaned) > (time.Second * 30){
		go cleanSessions()
	}

	http.Redirect(w,r,"/login",http.StatusSeeOther)
}