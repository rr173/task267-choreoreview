// 动作单元 HTTP 处理器。
package httpapi

import (
	"net/http"
)

// handleAddMovement POST /api/dances/{id}/movements
func (a *API) handleAddMovement(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		DancerNo   int    `json:"dancer_no"`
		ActionName string `json:"action_name"`
		StartBeat  int    `json:"start_beat"`
		EndBeat    int    `json:"end_beat"`
		Connection string `json:"connection"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	m, err := a.svc.Movement().AddMovement(id, req.DancerNo, req.StartBeat, req.EndBeat, req.ActionName, req.Connection)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

// handleListMovements GET /api/dances/{id}/movements
func (a *API) handleListMovements(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	movements, err := a.svc.Movement().ListMovements(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movements)
}
