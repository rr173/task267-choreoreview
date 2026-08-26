// 舞段项目 HTTP 处理器。
package httpapi

import (
	"net/http"
)

// handleCreateDance POST /api/dances
func (a *API) handleCreateDance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Region  string `json:"region"`
		Dancers int    `json:"dancers"`
		Notes   string `json:"notes"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	dance, err := a.svc.Dance().CreateDance(req.Name, req.Region, req.Dancers, req.Notes)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dance)
}

// handleListDances GET /api/dances
func (a *API) handleListDances(w http.ResponseWriter, r *http.Request) {
	dances, err := a.svc.Dance().ListDances()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dances)
}

// handleGetDance GET /api/dances/{id}
func (a *API) handleGetDance(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	dance, err := a.svc.Dance().GetDance(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dance)
}

// handleTransitDance PATCH /api/dances/{id}/status
func (a *API) handleTransitDance(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	dance, err := a.svc.Dance().TransitDance(id, req.Status)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dance)
}

// handleSealDance POST /api/dances/{id}/seal
func (a *API) handleSealDance(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	dance, err := a.svc.Dance().SealDance(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dance)
}
