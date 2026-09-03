package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/zixuan-come/whaleshop/internal/middleware"
	"github.com/zixuan-come/whaleshop/internal/model"
	"github.com/zixuan-come/whaleshop/internal/store"
)

type Orders struct {
	store *store.Store
}

func NewOrders(s *store.Store) *Orders {
	return &Orders{store: s}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Orders) List(w http.ResponseWriter, r *http.Request) {
	pid := middleware.ProjectID(r.Context())
	writeJSON(w, http.StatusOK, h.store.List(pid))
}

func (h *Orders) Get(w http.ResponseWriter, r *http.Request) {
	pid := middleware.ProjectID(r.Context())
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id 必须是整数")
		return
	}
	o, ok := h.store.Get(pid, id)
	if !ok {
		writeErr(w, http.StatusNotFound, "订单不存在")
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (h *Orders) Create(w http.ResponseWriter, r *http.Request) {
	pid := middleware.ProjectID(r.Context())
	var o model.Order
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}
	if o.Item == "" {
		writeErr(w, http.StatusUnprocessableEntity, "item 不能为空")
		return
	}
	if o.Quantity <= 0 {
		writeErr(w, http.StatusUnprocessableEntity, "quantity 必须 > 0")
		return
	}
	writeJSON(w, http.StatusCreated, h.store.Create(pid, &o))
}

func (h *Orders) Update(w http.ResponseWriter, r *http.Request) {
	pid := middleware.ProjectID(r.Context())
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id 必须是整数")
		return
	}
	var o model.Order
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体不是合法 JSON")
		return
	}
	// 业务规则示例:已取消的订单不能改 -> 409
	if existing, ok := h.store.Get(pid, id); ok && existing.Status == "cancelled" {
		writeErr(w, http.StatusConflict, "已取消的订单不能修改")
		return
	}
	updated, ok := h.store.Update(pid, id, &o)
	if !ok {
		writeErr(w, http.StatusNotFound, "订单不存在")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Orders) Delete(w http.ResponseWriter, r *http.Request) {
	pid := middleware.ProjectID(r.Context())
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id 必须是整数")
		return
	}
	if !h.store.Delete(pid, id) {
		writeErr(w, http.StatusNotFound, "订单不存在")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// Slow 定时 sleep,压测靶子。GET /orders/slow?ms=500
func (h *Orders) Slow(w http.ResponseWriter, r *http.Request) {
	ms := 500
	if raw := r.URL.Query().Get("ms"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 && n <= 30000 {
			ms = n
		}
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	writeJSON(w, http.StatusOK, map[string]int{"slept_ms": ms})
}

// Error 强返指定状态码,断言/失败重试演示。GET /orders/error?code=500
func (h *Orders) Error(w http.ResponseWriter, r *http.Request) {
	code := 500
	if raw := r.URL.Query().Get("code"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 100 && n < 600 {
			code = n
		}
	}
	writeErr(w, code, "触发指定错误码")
}
