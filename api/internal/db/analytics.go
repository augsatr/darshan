package db

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const popularTemples = `SELECT t.id, t.name, t.slug, t.state, COUNT(pv.id) as views
FROM page_views pv JOIN temples t ON t.id = pv.temple_id
GROUP BY t.id ORDER BY views DESC LIMIT $1`

type PopularTemple struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	State string `json:"state"`
	Views int    `json:"views"`
}

type ViewBuffer struct {
	mu       sync.Mutex
	buf      []int64
	maxSize  int
	interval time.Duration
	pool     *pgxpool.Pool
	stop     chan struct{}
	done     chan struct{}
	flushFn  func(context.Context, []int64) error // overrides pool flush in tests
}

func newViewBuffer(pool *pgxpool.Pool, maxSize int, interval time.Duration) *ViewBuffer {
	vb := &ViewBuffer{
		buf:      make([]int64, 0, maxSize),
		maxSize:  maxSize,
		interval: interval,
		pool:     pool,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	go vb.loop()
	return vb
}

func (vb *ViewBuffer) Add(templeID int64) {
	vb.mu.Lock()
	vb.buf = append(vb.buf, templeID)
	shouldFlush := len(vb.buf) >= vb.maxSize
	vb.mu.Unlock()
	if shouldFlush {
		vb.flush()
	}
}

func (vb *ViewBuffer) flush() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := vb.flushCtx(ctx); err != nil {
		slog.Error("view buffer flush failed", "error", err)
	}
}

func (vb *ViewBuffer) flushCtx(ctx context.Context) error {
	vb.mu.Lock()
	if len(vb.buf) == 0 {
		vb.mu.Unlock()
		return nil
	}
	batch := make([]int64, len(vb.buf))
	copy(batch, vb.buf)
	vb.buf = vb.buf[:0]
	vb.mu.Unlock()

	// Entries are removed from the buffer before the DB write, so a failed
	// flush drops the batch on the floor. This is intentional — analytics data
	// is fire-and-forget; blocking or re-queuing risks memory pressure,
	// duplicate inserts, or infinite retry loops under sustained outages.

	if vb.flushFn != nil {
		return vb.flushFn(ctx, batch)
	}

	rows := make([][]any, len(batch))
	for i, id := range batch {
		rows[i] = []any{id}
	}

	tx, err := vb.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.CopyFrom(ctx, pgx.Identifier{"page_views"}, []string{"temple_id"}, pgx.CopyFromRows(rows))
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (vb *ViewBuffer) Stop() {
	close(vb.stop)
	<-vb.done
}

func (vb *ViewBuffer) loop() {
	ticker := time.NewTicker(vb.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			vb.flush()
		case <-vb.stop:
			vb.flush()
			close(vb.done)
			return
		}
	}
}

func (d *DB) RecordPageView(ctx context.Context, templeID int64) error {
	if d.viewBuffer != nil {
		d.viewBuffer.Add(templeID)
		return nil
	}
	_, err := d.Exec(ctx, `INSERT INTO page_views (temple_id) VALUES ($1)`, templeID)
	return err
}

func (d *DB) PopularTemples(ctx context.Context, limit int) ([]PopularTemple, error) {
	rows, err := d.Query(ctx, popularTemples, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PopularTemple
	for rows.Next() {
		var p PopularTemple
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.State, &p.Views); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
