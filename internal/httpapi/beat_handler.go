// 节拍锚点 HTTP 处理器。
package httpapi

import (
	"net/http"
)

// handleAddBeatAnchor POST /api/dances/{id}/beats
func (a *API) handleAddBeatAnchor(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		BeatNo     int     `json:"beat_no"`
		ImageTime  float64 `json:"image_time"`
		Confidence float64 `json:"confidence"`
		Source     string  `json:"source"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	b, err := a.svc.Beat().AddBeatAnchor(id, req.BeatNo, req.ImageTime, req.Confidence, req.Source)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

// handleListBeatAnchors GET /api/dances/{id}/beats
func (a *API) handleListBeatAnchors(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	anchors, err := a.svc.Beat().ListBeatAnchors(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, anchors)
}
