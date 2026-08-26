// 队形关系 HTTP 处理器。
package httpapi

import (
	"net/http"
)

// handleAddFormation POST /api/dances/{id}/formations
func (a *API) handleAddFormation(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		FromDancer int    `json:"from_dancer"`
		ToDancer   int    `json:"to_dancer"`
		Relation   string `json:"relation"`
		BeatStart  int    `json:"beat_start"`
		BeatEnd    int    `json:"beat_end"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	f, err := a.svc.Formation().AddFormation(id, req.FromDancer, req.ToDancer, req.BeatStart, req.BeatEnd, req.Relation)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

// handleListFormations GET /api/dances/{id}/formations
func (a *API) handleListFormations(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	formations, err := a.svc.Formation().ListFormations(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, formations)
}
