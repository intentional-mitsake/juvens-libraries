package handlers

import "net/http"

func (rt *Router) MatchHandler(w http.ResponseWriter, r *http.Request) {
}

func (rt *Router) ShuffleHandler(w http.ResponseWriter, r *http.Request) {
}

func (rt *Router) MembersHandler(w http.ResponseWriter, r *http.Request) {}

func (rt *Router) MessagesHandler(w http.ResponseWriter, r *http.Request) {}

func (rt *Router) LeaveHandler(w http.ResponseWriter, r *http.Request) {}

func (rt *Router) StayHandler(w http.ResponseWriter, r *http.Request) {}
