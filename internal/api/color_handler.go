package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/wyw14/solo-038/internal/color"
	"github.com/wyw14/solo-038/internal/model"
)

type colorBaselineRequest struct {
	Ink       string             `json:"ink"`
	InkBatch  string             `json:"ink_batch"`
	Lab       model.LabColor     `json:"lab"`
	Tolerance float64            `json:"tolerance"`
	Formula   model.ColorFormula `json:"formula"`
}

type colorReadingRequest struct {
	InkBatch   string         `json:"ink_batch"`
	WebMeter   float64        `json:"web_meter"`
	Lab        model.LabColor `json:"lab"`
	MeasuredAt time.Time      `json:"measured_at"`
}

func (handler *Handler) establishColorBaseline(writer http.ResponseWriter, request *http.Request) {
	var input colorBaselineRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	if input.Formula == "" {
		input.Formula = model.ColorFormulaCIE2000
	}
	target := model.ColorTarget{Ink: input.Ink, Lab: input.Lab, Tolerance: input.Tolerance, Formula: input.Formula}
	baseline, err := handler.runtime.EstablishColorBaseline(target, input.InkBatch)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusCreated, baseline)
}

func (handler *Handler) createColorReading(writer http.ResponseWriter, request *http.Request) {
	var input colorReadingRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	deviation, err := handler.runtime.AcceptColorReading(input.InkBatch, input.WebMeter, input.Lab, input.MeasuredAt)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusCreated, deviation)
}

func (handler *Handler) colorStatus(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, handler.runtime.ColorStatus())
}

func (handler *Handler) releaseColor(writer http.ResponseWriter, request *http.Request) {
	record, err := handler.runtime.ReleaseColor()
	if err != nil {
		var exceeded *color.ExceededError
		if errors.As(err, &exceeded) {
			writeJSON(writer, http.StatusConflict, map[string]any{"error": err.Error(), "segments": exceeded.Segments})
			return
		}
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, record)
}

func (handler *Handler) colorReleases(writer http.ResponseWriter, request *http.Request) {
	releases, err := handler.runtime.ColorReleases()
	if err != nil {
		writeError(writer, http.StatusInternalServerError, err)
		return
	}
	writeJSON(writer, http.StatusOK, releases)
}
