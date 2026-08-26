// HTTP API 层：路由与 JSON 编解码。
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"task267-choreoreview/internal/model"
	"task267-choreoreview/internal/service"
)

// API HTTP 处理器。
type API struct {
	svc *service.Service
	mux *http.ServeMux
}

// New 构造 HTTP API。
func New(svc *service.Service) *API {
	a := &API{svc: svc, mux: http.NewServeMux()}
	a.routes()
	return a
}

// Handler 返回 HTTP 处理器。
func (a *API) Handler() http.Handler {
	return logMiddleware(a.mux)
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Printf("%s %s", r.Method, r.URL.Path)
	})
}

// routes 注册全部路由。
func (a *API) routes() {
	// 舞段
	a.mux.HandleFunc("POST /api/dances", a.handleCreateDance)
	a.mux.HandleFunc("GET /api/dances", a.handleListDances)
	a.mux.HandleFunc("GET /api/dances/{id}", a.handleGetDance)
	a.mux.HandleFunc("PATCH /api/dances/{id}/status", a.handleTransitDance)
	a.mux.HandleFunc("POST /api/dances/{id}/seal", a.handleSealDance)
	a.mux.HandleFunc("POST /api/dances/{id}/versions", a.handleCreateVersion)
	a.mux.HandleFunc("GET /api/dances/{id}/versions", a.handleListVersions)

	// 动作单元
	a.mux.HandleFunc("POST /api/dances/{id}/movements", a.handleAddMovement)
	a.mux.HandleFunc("GET /api/dances/{id}/movements", a.handleListMovements)

	// 节拍锚点
	a.mux.HandleFunc("POST /api/dances/{id}/beats", a.handleAddBeatAnchor)
	a.mux.HandleFunc("GET /api/dances/{id}/beats", a.handleListBeatAnchors)

	// 队形关系
	a.mux.HandleFunc("POST /api/dances/{id}/formations", a.handleAddFormation)
	a.mux.HandleFunc("GET /api/dances/{id}/formations", a.handleListFormations)

	// 分析
	a.mux.HandleFunc("POST /api/dances/{id}/analyze", a.handleRunFullAnalysis)
	a.mux.HandleFunc("GET /api/dances/{id}/alignment", a.handleAlignBeats)
	a.mux.HandleFunc("GET /api/dances/{id}/continuity", a.handleCheckContinuity)
	a.mux.HandleFunc("GET /api/dances/{id}/formation-check", a.handleCheckFormation)

	// 异读候选
	a.mux.HandleFunc("GET /api/dances/{id}/variants", a.handleListVariants)
	a.mux.HandleFunc("GET /api/dances/{id}/variants/open", a.handleOpenVariants)
	a.mux.HandleFunc("POST /api/variants/{id}/adjudicate", a.handleAdjudicateVariant)

	// 版本
	a.mux.HandleFunc("POST /api/versions/{id}/freeze", a.handleFreezeVersion)
	a.mux.HandleFunc("POST /api/versions/{id}/supersede", a.handleSupersedeVersion)

	// 统计与健康
	a.mux.HandleFunc("GET /api/stats", a.handleStats)
	a.mux.HandleFunc("GET /api/health", a.handleHealth)
}

// idParam 解析路径中的数字 ID。
func idParam(r *http.Request, name string) (int64, error) {
	raw := r.PathValue(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("%w: invalid %s=%q", model.ErrBadRequest, name, raw)
	}
	return id, nil
}

// writeJSON 写 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeErr 写错误响应。
func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrBadRequest):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrDuplicate), errors.Is(err, model.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, model.ErrInvalidState), errors.Is(err, model.ErrSealed),
		errors.Is(err, model.ErrVersionConflict), errors.Is(err, model.ErrFormationConflict):
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// decodeJSON 解析请求体。
func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: %v", model.ErrBadRequest, err)
	}
	return nil
}

// trimJoin 拼接错误信息。
func trimJoin(parts ...string) string {
	return strings.Join(parts, " ")
}
