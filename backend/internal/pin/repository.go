package pin

import (
	"context"
	"fmt"
	"time"

	"github.com/glebarez/sqlite" // pure Go(modernc ベース)の GORM ドライバ。非cgo。
	"gorm.io/gorm"

	"github.com/kisaragi-ai-map/backend/internal/outbound"
)

// PinRepository は永続層の seam。DB の詳細はこの interface の裏に隠す。
type PinRepository interface {
	GetPins(ctx context.Context) ([]Pin, error)
	// Count は seed 判定（空かどうか）に使う。
	Count(ctx context.Context) (int, error)
	// Insert は seed が1件投入するための最小 API。
	Insert(ctx context.Context, p Pin) error
	// ListForStats は提出用集計に必要な最小データ（prefecture/ip_hash）を全件返す。
	ListForStats(ctx context.Context) ([]PinStat, error)
}

// pinRow は永続化モデル。スキーマ（カラム/制約）を知るのはこのファイルだけ。
// ドメインの Pin には DB 知識（GORM タグ）を持ち込まない（案B: モデル分離）。
type pinRow struct {
	ID         uint    `gorm:"primaryKey;autoIncrement"`
	Prefecture string  `gorm:"not null"`
	Lat        float64 `gorm:"not null"`
	Lng        float64 `gorm:"not null"`
	// ファン投稿の表示用フィールド。seed 由来のピンでは空文字。
	Nickname string `gorm:"not null;default:''"`
	City     string `gorm:"not null;default:''"`
	Comment  string `gorm:"not null;default:''"`
	// IPHash は投稿者の匿名識別子（分析専用）。地図には出さない。既存行は空文字。
	IPHash string `gorm:"not null;default:'';index"`
	// 分析専用フィールド。後から復元できない情報を投稿の瞬間に保存する。地図には出さない。
	// PrefectureCode は JIS 都道府県コード(2桁)。集計高速化の非正規化。既存行は空文字。
	PrefectureCode string `gorm:"not null;default:'';index"`
	// AnonToken は匿名投稿者の識別子（後から claim する余地用）。既存行は空文字。
	AnonToken string `gorm:"not null;default:'';index"`
	// UTM* は流入元の計測値（X 共有リンク等）。既存行は空文字。
	UTMSource   string `gorm:"not null;default:''"`
	UTMMedium   string `gorm:"not null;default:''"`
	UTMCampaign string `gorm:"not null;default:''"`
	// CreatedAt は GORM が作成時に自動設定する。連投・curl の時系列分析用。
	CreatedAt time.Time
}

// TableName は GORM の複数形化（pin_rows）を抑え、テーブル名を pins に固定する。
func (pinRow) TableName() string { return "pins" }

// outboundClickRow は公式 URL への計測付き送客クリックの永続化モデル。
// DB を知るのはこのファイルだけ、という隔離原則を守るためここに同居させる
// （outbound パッケージ自体はドライバを import しない）。
type outboundClickRow struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	Destination string `gorm:"not null;index"` // ホワイトリストのキー（official_menu 等）
	UTMSource   string `gorm:"not null;default:''"`
	UTMMedium   string `gorm:"not null;default:''"`
	UTMCampaign string `gorm:"not null;default:''"`
	// CreatedAt は UTC で保存する（後から復元できないため）。
	CreatedAt time.Time `gorm:"not null"`
}

func (outboundClickRow) TableName() string { return "outbound_clicks" }

func (r pinRow) toDomain() Pin {
	return Pin{
		Prefecture: Prefecture(r.Prefecture), Lat: r.Lat, Lng: r.Lng,
		Nickname: r.Nickname, City: r.City, Comment: r.Comment,
		IPHash:         r.IPHash,
		PrefectureCode: r.PrefectureCode, AnonToken: r.AnonToken,
		UTMSource: r.UTMSource, UTMMedium: r.UTMMedium, UTMCampaign: r.UTMCampaign,
	}
}

func rowFromDomain(p Pin) pinRow {
	return pinRow{
		Prefecture: string(p.Prefecture), Lat: p.Lat, Lng: p.Lng,
		Nickname: p.Nickname, City: p.City, Comment: p.Comment,
		IPHash:         p.IPHash,
		PrefectureCode: p.PrefectureCode, AnonToken: p.AnonToken,
		UTMSource: p.UTMSource, UTMMedium: p.UTMMedium, UTMCampaign: p.UTMCampaign,
	}
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
	if err := db.AutoMigrate(&pinRow{}, &outboundClickRow{}); err != nil {
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

// ListForStats は集計に必要な prefecture/ip_hash だけを全件返す。
// 地図用の GetPins とは別経路で、提出用のユニークファン集計に使う。
func (r *sqliteRepo) ListForStats(ctx context.Context) ([]PinStat, error) {
	var rows []pinRow
	if err := r.db.WithContext(ctx).Select("prefecture", "ip_hash").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("集計用ピンの取得: %w", err)
	}
	stats := make([]PinStat, 0, len(rows))
	for _, row := range rows {
		stats = append(stats, PinStat{Prefecture: Prefecture(row.Prefecture), IPHash: row.IPHash})
	}
	return stats, nil
}

// RecordClick は outbound.ClickRepository の実装。同じ DB 接続にクリックを1件保存する。
// 時刻は UTC で保存する（呼び出し側が UTC で渡す前提だが、防御的にここでも UTC へ正規化する）。
func (r *sqliteRepo) RecordClick(ctx context.Context, c outbound.Click) error {
	row := outboundClickRow{
		Destination: c.Destination,
		UTMSource:   c.UTMSource, UTMMedium: c.UTMMedium, UTMCampaign: c.UTMCampaign,
		CreatedAt: c.CreatedAt.UTC(),
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("クリックの記録: %w", err)
	}
	return nil
}

// listClicks は記録済みクリックを全件返すテスト専用ヘルパ。分析用の読み出し経路は
// スコープ外（分析 UI は未実装）のため、ここでは結合テストの検証用にのみ提供する。
func (r *sqliteRepo) listClicks(ctx context.Context) ([]outbound.Click, error) {
	var rows []outboundClickRow
	if err := r.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("クリック一覧の取得: %w", err)
	}
	clicks := make([]outbound.Click, 0, len(rows))
	for _, row := range rows {
		clicks = append(clicks, outbound.Click{
			Destination: row.Destination,
			UTMSource:   row.UTMSource, UTMMedium: row.UTMMedium, UTMCampaign: row.UTMCampaign,
			CreatedAt: row.CreatedAt,
		})
	}
	return clicks, nil
}
