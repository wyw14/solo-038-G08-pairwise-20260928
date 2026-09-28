package api

import (
	"net/http"

	"github.com/wyw14/solo-038/internal/color"
	"github.com/wyw14/solo-038/internal/model"
)

type colorBaselineRequest struct {
	Target    model.Lab `json:"target"`
	Tolerance float64   `json:"tolerance"`
	Formula   string    `json:"formula"`
}

func (handler *Handler) createColorBaseline(writer http.ResponseWriter, request *http.Request) {
	var input colorBaselineRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	baseline, err := handler.runtime.EstablishColorBaseline(input.Target, input.Tolerance, input.Formula)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusCreated, baseline)
}

type colorMeasurementsRequest struct {
	Samples []color.Sample `json:"samples"`
}

func (handler *Handler) createColorMeasurements(writer http.ResponseWriter, request *http.Request) {
	var input colorMeasurementsRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	measurements, err := handler.runtime.RecordColorMeasurements(input.Samples)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusAccepted, measurements)
}

func (handler *Handler) colorStatus(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, handler.runtime.ColorStatus())
}

func (handler *Handler) releaseColor(writer http.ResponseWriter, request *http.Request) {
	record, err := handler.runtime.ReleaseColor(request.Context())
	if err != nil {
		if record.BaselineID != "" {
			writeJSON(writer, http.StatusConflict, record)
			return
		}
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, record)
}

type inkBatchRequest struct {
	Batch string `json:"batch"`
}

func (handler *Handler) setInkBatch(writer http.ResponseWriter, request *http.Request) {
	var input inkBatchRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	if err := handler.runtime.ChangeInkBatch(input.Batch); err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"ink_batch": input.Batch})
}
