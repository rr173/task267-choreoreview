// 分析与异读候选 HTTP 处理器。
package httpapi

import (
	"net/http"
)

// handleRunFullAnalysis POST /api/dances/{id}/analyze
func (a *API) handleRunFullAnalysis(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	res, err := a.svc.Analyze().RunFullAnalysis(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleAlignBeats GET /api/dances/{id}/alignment
func (a *API) handleAlignBeats(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	res, err := a.svc.Analyze().AlignBeats(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleCheckContinuity GET /api/dances/{id}/continuity
func (a *API) handleCheckContinuity(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	res, err := a.svc.Analyze().CheckContinuity(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleCheckFormation GET /api/dances/{id}/formation-check
func (a *API) handleCheckFormation(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	res, err := a.svc.Analyze().CheckFormation(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleListVariants GET /api/dances/{id}/variants
func (a *API) handleListVariants(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	variants, err := a.svc.Variant().ListVariants(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, variants)
}

// handleOpenVariants GET /api/dances/{id}/variants/open
func (a *API) handleOpenVariants(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	variants, err := a.svc.Variant().OpenVariants(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, variants)
}

// handleAdjudicateVariant POST /api/variants/{id}/adjudicate
func (a *API) handleAdjudicateVariant(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		Verdict string `json:"verdict"`
		Reason  string `json:"reason"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	v, err := a.svc.Variant().Adjudicate(id, req.Verdict, req.Reason)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}
