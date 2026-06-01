package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/your-org/pii-admin-api/internal/audit"
	"github.com/your-org/pii-admin-api/internal/auth"
	"github.com/your-org/pii-admin-api/internal/model"
	"github.com/your-org/pii-admin-api/internal/repo"
)

// Handler gom các phụ thuộc cho HTTP layer.
type Handler struct {
	Repo  *repo.Repo
	Audit *audit.Recorder
}

func New(r *repo.Repo, a *audit.Recorder) *Handler {
	return &Handler{Repo: r, Audit: a}
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func notImplemented(w http.ResponseWriter, ref string) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "not_implemented",
		"todo":  ref,
	})
}

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	return nil
}

// =====================================================================
// F1 — Xem & tìm kiếm log
// =====================================================================

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var cursor *int64
	if c := q.Get("cursor"); c != "" {
		if n, err := strconv.ParseInt(c, 10, 64); err == nil {
			cursor = &n
		}
	}
	f := model.LogFilter{
		Actor:      q.Get("actor"),
		Result:     q.Get("result"),
		Field:      q.Get("field"),
		Purpose:    q.Get("purpose"),
		SubjectRef: q.Get("subject_ref"),
		From:       parseTime(q.Get("from")),
		To:         parseTime(q.Get("to")),
		Cursor:     cursor,
		Limit:      limit,
	}

	// Self-audit việc xem log.
	_ = h.Audit.Record(r.Context(), audit.Event{
		Admin:  actorOf(r),
		Action: "VIEW_LOGS",
		Detail: map[string]any{"filter": f},
	})

	page, err := h.Repo.ListLogs(r.Context(), f)
	if err != nil {
		notImplemented(w, "API-ADM-03 / repo.ListLogs")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *Handler) GetLog(w http.ResponseWriter, r *http.Request) {
	seq, _ := strconv.ParseInt(chi.URLParam(r, "seq"), 10, 64)
	entry, err := h.Repo.GetLog(r.Context(), seq)
	if err != nil {
		notImplemented(w, "API-ADM-04 / repo.GetLog")
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (h *Handler) VerifyChain(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FromSeq int64 `json:"from_seq"`
		ToSeq   int64 `json:"to_seq"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	res, err := h.Repo.VerifyChain(r.Context(), body.FromSeq, body.ToSeq)
	if err != nil {
		notImplemented(w, "API-ADM-05 / repo.VerifyChain")
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// =====================================================================
// F2 — Dashboard thống kê
// =====================================================================

func (h *Handler) StatsAccess(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	bucket := q.Get("bucket")
	if bucket == "" {
		bucket = "hour"
	}
	series, err := h.Repo.AccessSeries(r.Context(), q.Get("from"), q.Get("to"), bucket)
	if err != nil {
		notImplemented(w, "API-ADM-06 / repo.AccessSeries")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"series": series})
}

func (h *Handler) StatsTopActors(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	items, err := h.Repo.TopActors(r.Context(), q.Get("from"), q.Get("to"), limit)
	if err != nil {
		notImplemented(w, "API-ADM-06 / repo.TopActors")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) StatsSummary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	s, err := h.Repo.Summary(r.Context(), q.Get("from"), q.Get("to"))
	if err != nil {
		notImplemented(w, "API-ADM-06 / repo.Summary")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

// =====================================================================
// F3 — Cảnh báo & điều tra sự cố
// =====================================================================

func (h *Handler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	items, err := h.Repo.ListAlerts(r.Context(), q.Get("status"), q.Get("severity"), nil, limit)
	if err != nil {
		notImplemented(w, "API-ADM-07 / repo.ListAlerts")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) UpdateAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Status   string `json:"status"`
		Assignee string `json:"assignee"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	_ = h.Audit.Record(r.Context(), audit.Event{
		Admin: actorOf(r), Action: "CHANGE_ALERT", Target: id,
		Detail: map[string]any{"status": body.Status, "assignee": body.Assignee},
	})

	if err := h.Repo.UpdateAlert(r.Context(), id, body.Status, body.Assignee); err != nil {
		notImplemented(w, "API-ADM-07 / repo.UpdateAlert")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) SubjectTimeline(w http.ResponseWriter, r *http.Request) {
	ref := chi.URLParam(r, "ref")
	items, err := h.Repo.SubjectTimeline(r.Context(), ref)
	if err != nil {
		notImplemented(w, "API-ADM-08 / repo.SubjectTimeline")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) RevokeActor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Thao tác rủi ro cao: phải đi qua phê duyệt bốn mắt (tạo approval_request),
	// KHÔNG thực thi trực tiếp ở đây.
	_ = h.Audit.Record(r.Context(), audit.Event{
		Admin: actorOf(r), Action: "REQUEST_REVOKE", Target: id,
	})
	notImplemented(w, "API-ADM-10 / tạo approval_request rồi gọi M5 sau khi duyệt")
}

// =====================================================================
// F4 — Quyền & phê duyệt bốn mắt
// =====================================================================

func (h *Handler) ListApprovals(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}
	items, err := h.Repo.ListApprovals(r.Context(), status)
	if err != nil {
		notImplemented(w, "API-ADM-11 / repo.ListApprovals")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	h.decide(w, r, "approved")
}

func (h *Handler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	h.decide(w, r, "rejected")
}

func (h *Handler) decide(w http.ResponseWriter, r *http.Request, decision string) {
	id := chi.URLParam(r, "id")
	approver := actorOf(r)
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	// QUY TẮC BỐN MẮT: approver phải khác requester.
	// TODO(API-ADM-11): load request, kiểm tra requester != approver,
	// nếu trùng -> 409 Conflict "four-eyes violation".

	_ = h.Audit.Record(r.Context(), audit.Event{
		Admin: approver, Action: "DECIDE_APPROVAL", Target: id,
		Detail: map[string]any{"decision": decision, "reason": body.Reason},
	})

	if err := h.Repo.DecideApproval(r.Context(), id, approver, decision, body.Reason); err != nil {
		notImplemented(w, "API-ADM-11 / repo.DecideApproval")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "decision": decision})
}

func actorOf(r *http.Request) string {
	if p := auth.FromContext(r.Context()); p != nil {
		return p.Subject
	}
	return "unknown"
}
