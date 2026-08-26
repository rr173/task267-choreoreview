// 版本、统计与健康检查 HTTP 处理器。
package httpapi

import (
	"net/http"
)

// handleCreateVersion POST /api/dances/{id}/versions
func (a *API) handleCreateVersion(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		Summary  string `json:"summary"`
		Evidence string `json:"evidence"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	v, err := a.svc.Version().CreateVersion(id, req.Summary, req.Evidence)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

// handleListVersions GET /api/dances/{id}/versions
func (a *API) handleListVersions(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	versions, err := a.svc.Version().ListVersions(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, versions)
}

// handleFreezeVersion POST /api/versions/{id}/freeze
func (a *API) handleFreezeVersion(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	v, err := a.svc.Version().FreezeVersion(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// handleSupersedeVersion POST /api/versions/{id}/supersede
func (a *API) handleSupersedeVersion(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	v, err := a.svc.Version().SupersedeVersion(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// handleStats GET /api/stats
func (a *API) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := a.svc.Stats()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// handleHealth GET /api/health
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
