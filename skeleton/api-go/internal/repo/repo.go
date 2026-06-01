package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/pii-admin-api/internal/model"
)

// ErrNotImplemented đánh dấu phần cần đội dev hiện thực.
var ErrNotImplemented = errors.New("repo: not implemented")

// Repo bọc pool kết nối DB. Dùng credentials CHỈ-ĐỌC cho bảng audit.
type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// ListLogs — F1: lọc + keyset pagination trên pii_audit.
//
// TODO(API-ADM-03): hiện thực truy vấn, ví dụ:
//
//	SELECT seq, ts, actor, action, subject_ref, field, purpose, result
//	FROM pii_audit
//	WHERE ($1='' OR actor=$1)
//	  AND ($2='' OR result=$2)
//	  AND ($3::timestamptz IS NULL OR ts >= $3)
//	  AND ($4::timestamptz IS NULL OR ts <= $4)
//	  AND ($5::bigint IS NULL OR seq < $5)   -- keyset cursor
//	ORDER BY seq DESC
//	LIMIT $6;
//
// next_cursor = seq của phần tử cuối nếu còn trang.
func (r *Repo) ListLogs(ctx context.Context, f model.LogFilter) (model.LogPage, error) {
	return model.LogPage{}, ErrNotImplemented
}

// GetLog — F1: chi tiết một bản ghi theo seq.
func (r *Repo) GetLog(ctx context.Context, seq int64) (model.LogEntry, error) {
	return model.LogEntry{}, ErrNotImplemented
}

// VerifyChain — F1: kiểm tra hash-chain trong khoảng [from,to].
//
// TODO(API-ADM-05): đọc tuần tự, tính lại row_hash từ prev_hash và so khớp;
// trả broken_at tại dòng đầu tiên sai.
func (r *Repo) VerifyChain(ctx context.Context, fromSeq, toSeq int64) (model.VerifyChainResult, error) {
	return model.VerifyChainResult{}, ErrNotImplemented
}

// AccessSeries — F2: đọc từ materialized view mv_access_hourly.
func (r *Repo) AccessSeries(ctx context.Context, fromISO, toISO, bucket string) ([]model.AccessPoint, error) {
	return nil, ErrNotImplemented
}

// TopActors — F2.
func (r *Repo) TopActors(ctx context.Context, fromISO, toISO string, limit int) ([]model.TopActor, error) {
	return nil, ErrNotImplemented
}

// Summary — F2.
func (r *Repo) Summary(ctx context.Context, fromISO, toISO string) (model.Summary, error) {
	return model.Summary{}, ErrNotImplemented
}

// ListAlerts — F3.
func (r *Repo) ListAlerts(ctx context.Context, status, severity string, cursor *int64, limit int) ([]model.Alert, error) {
	return nil, ErrNotImplemented
}

// UpdateAlert — F3: đổi trạng thái/assignee.
func (r *Repo) UpdateAlert(ctx context.Context, alertID, status, assignee string) error {
	return ErrNotImplemented
}

// SubjectTimeline — F3: dựng lịch sử truy cập một subject.
func (r *Repo) SubjectTimeline(ctx context.Context, subjectRef string) ([]model.TimelineEntry, error) {
	return nil, ErrNotImplemented
}

// ListApprovals — F4.
func (r *Repo) ListApprovals(ctx context.Context, status string) ([]model.ApprovalRequest, error) {
	return nil, ErrNotImplemented
}

// DecideApproval — F4: approve/reject. Bắt buộc approver != requester.
func (r *Repo) DecideApproval(ctx context.Context, requestID, approver, decision, reason string) error {
	return ErrNotImplemented
}
