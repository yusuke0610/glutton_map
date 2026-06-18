package pin

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite" // pure Go ドライバ（driver 名は "sqlite"）
)

// PinRepository は永続層の seam。DB の詳細はこの interface の裏に隠す。
type PinRepository interface {
	GetPins(ctx context.Context) ([]Pin, error)
	// Count は seed 判定（空かどうか）に使う。
	Count(ctx context.Context) (int, error)
	// Insert は seed が1件投入するための最小 API。
	Insert(ctx context.Context, p Pin) error
}

// sqliteRepo は modernc.org/sqlite を使う唯一の実装。
// DB(database/sql / driver) を知るのはこのファイルだけ。
type sqliteRepo struct{ db *sql.DB }

// NewSQLiteRepository は DSN（例: file:./data/pins.db）で接続しテーブルを用意する。
func NewSQLiteRepository(dsn string) (PinRepository, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS pins (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			prefecture TEXT    NOT NULL,
			lat        REAL    NOT NULL,
			lng        REAL    NOT NULL
		)`); err != nil {
		return nil, err
	}
	return &sqliteRepo{db: db}, nil
}

func (r *sqliteRepo) GetPins(ctx context.Context) ([]Pin, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT prefecture, lat, lng FROM pins`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pins []Pin
	for rows.Next() {
		var p Pin
		if err := rows.Scan(&p.Prefecture, &p.Lat, &p.Lng); err != nil {
			return nil, err
		}
		pins = append(pins, p)
	}
	return pins, rows.Err()
}

func (r *sqliteRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pins`).Scan(&n)
	return n, err
}

func (r *sqliteRepo) Insert(ctx context.Context, p Pin) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO pins (prefecture, lat, lng) VALUES (?, ?, ?)`,
		string(p.Prefecture), p.Lat, p.Lng)
	return err
}
