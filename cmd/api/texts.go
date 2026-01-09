package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/epaitoo/ephermalbridge/internal/data"
)

func (app *application) getAllTextsHandler(w http.ResponseWriter, r *http.Request) {
	texts, err := app.models.Texts.GetAllTexts()

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"texts": texts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showTextsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	text, err := app.models.Texts.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"text": text}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createTextHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		TextDetails string `json:"text_details"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r)
		return
	}

	text := &data.Texts{
		Name:        input.Name,
		TextDetails: input.TextDetails,
	}

	// Insert New Exercise category
	err = app.models.Texts.Insert(text)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/texts/%d", text.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"text": text}, headers)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateTextHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	text, err := app.models.Texts.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name        *string `json:"name"`
		TextDetails *string `json:"text_details"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r)
		return
	}

	if input.Name != nil {
		text.Name = *input.Name
	}

	if input.TextDetails != nil {
		text.TextDetails = *input.TextDetails
	}

	err = app.models.Texts.Update(text)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"text": text}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteTextHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Texts.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Text Successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
