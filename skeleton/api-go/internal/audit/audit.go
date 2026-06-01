package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Event — một thao tác trên trang admin cần ghi self-audit.
type Event struct {
	Admin  string
	Action string // VIEW_LOGS | CHANGE_ALERT | APPROVE | REJECT | REVOKE ...
	Target string
	Detail map[string]any
}

// Recorder ghi self-audit theo hash-chain.
// TODO(API-ADM-13): thay phần lưu in-memory bằng INSERT vào bảng admin_audit,
// đảm bảo append-only (đã thu hồi UPDATE/DELETE ở DB).
type Recorder struct {
	mu       sync.Mutex
	prevHash string
}

func NewRecorder() *Recorder { return &Recorder{} }

// Record tính row_hash móc với prev_hash và (sẽ) ghi xuống DB.
func (rc *Recorder) Record(ctx context.Context, e Event) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	ts := time.Now().UTC().Format(time.RFC3339Nano)
	payload := rc.prevHash + "|" + ts + "|" + e.Admin + "|" + e.Action + "|" + e.Target
	sum := sha256.Sum256([]byte(payload))
	rowHash := hex.EncodeToString(sum[:])

	// TODO: INSERT INTO admin_audit (ts, admin, action, target, detail, prev_hash, row_hash)
	//       VALUES (...)
	rc.prevHash = rowHash
	return nil
}
