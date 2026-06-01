package model

import "time"

// LogEntry — một bản ghi audit hiển thị trên trang admin (KHÔNG chứa PII thật).
type LogEntry struct {
	Seq        int64     `json:"seq"`
	TS         time.Time `json:"ts"`
	Actor      string    `json:"actor"`
	Action     string    `json:"action"`
	SubjectRef string    `json:"subject_ref"`
	Field      string    `json:"field"`
	Purpose    string    `json:"purpose"`
	Result     string    `json:"result"`
}

// LogPage — kết quả phân trang keyset.
type LogPage struct {
	Items      []LogEntry `json:"items"`
	NextCursor *int64     `json:"next_cursor"`
}

// LogFilter — điều kiện lọc log.
type LogFilter struct {
	Actor      string
	Result     string
	Field      string
	Purpose    string
	SubjectRef string
	From       *time.Time
	To         *time.Time
	Cursor     *int64
	Limit      int
}

// VerifyChainResult — kết quả xác minh chuỗi hash audit.
type VerifyChainResult struct {
	OK       bool   `json:"ok"`
	BrokenAt *int64 `json:"broken_at,omitempty"`
}

// AccessPoint — một điểm dữ liệu cho biểu đồ dashboard.
type AccessPoint struct {
	Bucket          time.Time `json:"bucket"`
	AllowCnt        int64     `json:"allow_cnt"`
	DenyCnt         int64     `json:"deny_cnt"`
	DistinctActors  int64     `json:"actors"`
}

// TopActor — dòng trong bảng top actor.
type TopActor struct {
	Actor    string  `json:"actor"`
	Count    int64   `json:"count"`
	DenyRate float64 `json:"deny_rate"`
}

// Summary — thẻ tổng quan.
type Summary struct {
	Total            int64 `json:"total"`
	Allow            int64 `json:"allow"`
	Deny             int64 `json:"deny"`
	DistinctActors   int64 `json:"distinct_actors"`
	DistinctSubjects int64 `json:"distinct_subjects"`
}

// Alert — cảnh báo do M7 sinh.
type Alert struct {
	AlertID    string    `json:"alert_id"`
	CreatedAt  time.Time `json:"created_at"`
	Rule       string    `json:"rule"`
	Severity   string    `json:"severity"`
	Actor      string    `json:"actor"`
	Status     string    `json:"status"`
	Assignee   string    `json:"assignee,omitempty"`
	IncidentID string    `json:"incident_id,omitempty"`
}

// TimelineEntry — một mốc trong timeline của subject.
type TimelineEntry struct {
	TS      time.Time `json:"ts"`
	Actor   string    `json:"actor"`
	Action  string    `json:"action"`
	Purpose string    `json:"purpose"`
	Result  string    `json:"result"`
}

// ApprovalRequest — yêu cầu phê duyệt bốn mắt.
type ApprovalRequest struct {
	RequestID string    `json:"request_id"`
	Requester string    `json:"requester"`
	Action    string    `json:"action"`
	Purpose   string    `json:"purpose"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
