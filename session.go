package main

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"time"
)

func alreadyLoggedIn(r *http.Request) bool{
	c, err := r.Cookie("session")
	if err != nil {
		return false
	}
	s := dbSession[c.Value]
	_,ok := dbUsers[s.un]
	return ok
}
func getCookie(w http.ResponseWriter,r *http.Request) *http.Cookie{
	c ,err := r.Cookie("session")
	if err == http.ErrNoCookie{
		u, _ := uuid.NewV4()
		c = &http.Cookie{Name:"session",Value:u.String(),HttpOnly:true,MaxAge:sessionlength}
		http.SetCookie(w, c)
	}
	c.MaxAge = sessionlength
	http.SetCookie(w,c)
	return c
}
func cleanSessions(){
	fmt.Println("before cleaning")
	showSessions()

	for k,v := range dbSession{
		if time.Now().Sub(v.lastActivity) > (time.Second * 30) {
			delete(dbSession, k)
		}
		dbSessionCleaned = time.Now()
	}

	fmt.Println("after cleaning")
	showSessions()
}
func showSessions(){
	fmt.Println("**********")
	for k,v := range dbSession{
		fmt.Println(k,v.un)
	}
	fmt.Println("")
}