package pin

import (
	"context"
	"fmt"

	"github.com/glebarez/sqlite" // pure Go(modernc ベース)の GORM ドライバ。非cgo。
	"gorm.io/gorm"
)

// PinRepository は永続層の seam。DB の詳細はこの interface の裏に隠す。
type PinRepository interface {
	GetPins(ctx context.Context) ([]Pin, error)
	// Count は seed 判定（空かどうか）に使う。
	Count(ctx context.Context) (int, error)
	// Insert は seed が1件投入するための最小 API。
	Insert(ctx context.Context, p Pin) error
}

// pinRow は永続化モデル。スキーマ（カラム/制約）を知るのはこのファイルだけ。
// ドメインの Pin には DB 知識（GORM タグ）を持ち込まない（案B: モデル分離）。
type pinRow struct {
	ID         uint    `gorm:"primaryKey;autoIncrement"`
	Prefecture string  `gorm:"not null"`
	Lat        float64 `gorm:"not null"`
	Lng        float64 `gorm:"not null"`
}

// TableName は GORM の複数形化（pin_rows）を抑え、テーブル名を pins に固定する。
func (pinRow) TableName() string { return "pins" }

func (r pinRow) toDomain() Pin {
	return Pin{Prefecture: Prefecture(r.Prefecture), Lat: r.Lat, Lng: r.Lng}
}

func rowFromDomain(p Pin) pinRow {
	return pinRow{Prefecture: string(p.Prefecture), Lat: p.Lat, Lng: p.Lng}
}

// sqliteRepo は glebarez/sqlite を使う唯一の実装。
// DB(GORM/driver) を知るのはこのファイルだけ。
type sqliteRepo struct{ db *gorm.DB }

// NewSQLiteRepository は DSN（例: file:./data/pins.db）で接続し、
// 起動時 AutoMigrate で pins テーブルを pinRow に追従させる。
func NewSQLiteRepository(dsn string) (PinRepository, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("DB 接続: %w", err)
	}
	if err := db.AutoMigrate(&pinRow{}); err != nil {
		return nil, fmt.Errorf("マイグレーション: %w", err)
	}
	return &sqliteRepo{db: db}, nil
}

func (r *sqliteRepo) GetPins(ctx context.Context) ([]Pin, error) {
	var rows []pinRow
	if err := r.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("ピン一覧の取得: %w", err)
	}
	pins := make([]Pin, 0, len(rows))
	for _, row := range rows {
		pins = append(pins, row.toDomain())
	}
	return pins, nil
}

func (r *sqliteRepo) Count(ctx context.Context) (int, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&pinRow{}).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("ピン件数の取得: %w", err)
	}
	return int(n), nil
}

func (r *sqliteRepo) Insert(ctx context.Context, p Pin) error {
	row := rowFromDomain(p)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("ピンの挿入: %w", err)
	}
	return nil
}
