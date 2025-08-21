package recorder

import (
	"time"

	"github.com/ARK-ASWINRAJ/tat/internal/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ingester struct {
	st *storage.Store
	ch chan any
}

func NewIngester(st *storage.Store) *Ingester {
	i := &Ingester{
		st: st,
		ch: make(chan any, 5000),
	}
	go i.loop()
	return i
}

func (i *Ingester) Emit(ev any) {
	select {
	case i.ch <- ev:
	default:
		// queue full; drop event for now (MVP)
	}
}

func (i *Ingester) loop() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	batch := make([]any, 0, 1000)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		i.flush(batch)
		batch = batch[:0]
	}

	for {
		select {
		case ev := <-i.ch:
			batch = append(batch, ev)
			if len(batch) >= 1000 {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (i *Ingester) flush(batch []any) {
	_ = i.st.DB.Transaction(func(tx *gorm.DB) error {
		for _, ev := range batch {
			switch e := ev.(type) {
			case CommandStart:
				cmd := storage.Command{
					ID:        uuid.NewString(),
					SessionID: e.SessionID,
					LineNo:    0,
					Cmd:       e.Cmd,
					CWD:       e.CWD,
					StartedAt: e.StartedAt,
				}
				if err := tx.Create(&cmd).Error; err != nil {
					return err
				}

			case CommandEnd:
				if err := tx.Model(&storage.Command{}).
					Where("id = ?", e.CommandID).
					Updates(map[string]any{
						"exit_code":   e.ExitCode,
						"ended_at":    e.EndedAt,
						"duration_ms": e.DurationMs,
					}).Error; err != nil {
					return err
				}

			case CommandOutput:
				out := storage.Output{
					ID:        uuid.NewString(),
					CommandID: e.CommandID,
					Stream:    e.Stream,
					Chunk:     e.Chunk,
					At:        e.At,
				}
				if err := tx.Create(&out).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
