package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-insiders/site/internal/data"
	"github.com/golang-insiders/site/internal/types"
)

func (app *application) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Fprint(w, "Only accepts get requests")
		return
	}

	app.tmpl.render(w, "index", nil)
}

func (app *application) handleTalkForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Fprint(w, "Only accepts get requests")
		return
	}

	templateData := newTemplateData()
	templateData.TimeZones = app.services.TimeZone.LoadTimeZones("")

	app.tmpl.render(w, "new-talk", templateData)
}

func (app *application) handleTalkPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Fprint(w, "Only accepts post requests")
		return
	}

	ctx := context.Background()

	username := r.PostFormValue("twitter-username")
	title := r.PostFormValue("title")
	summary := r.PostFormValue("summary")
	tz := r.PostFormValue("time-zone")
	talk := types.Talk{
		TwitterUsername: username,
		Title:           title,
		Summary:         summary,
		Timezone:        tz,
	}

	errs := types.ValidateTalk(talk)
	if errs != nil {
		templateData := newTemplateData()
		templateData.TimeZones = app.services.TimeZone.LoadTimeZones("")
		for _, e := range errs {
			templateData.Errors = append(templateData.Errors, e.Error())
		}

		app.tmpl.render(w, "new-talk", templateData)
		return
	}

	err := app.services.Talks.Insert(ctx, &talk)
	if err != nil {
		log.Println("Error inserting data", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error inserting data")
		return
	}

	redirectUrl := fmt.Sprintf("/talk?id=%d", talk.ID)
	http.Redirect(w, r, redirectUrl, http.StatusFound)
}

func (app *application) handleGetTalkByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Fprint(w, "Only accepts get requests")
		return
	}

	ctx := context.Background()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Fprint(w, "ID must be an int")
		return
	}

	t, err := app.services.Talks.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Talk not found")
		default:
			log.Println(err)
			fmt.Fprintf(w, "Error getting talk")
		}
		return
	}

	templateData := newTemplateData()
	templateData.Talk = t

	app.tmpl.render(w, "talk-id", templateData)
}
