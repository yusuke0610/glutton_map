// Package config は main.go の起動時バリデーションをテスト可能な純粋関数として切り出す。
package config

import "fmt"

// Env は起動時バリデーションに関係する環境変数の値。
type Env struct {
	// IPHashSalt は ip_hash 生成の salt（IP_HASH_SALT）。
	IPHashSalt string
	// TrustedProxies は ClientIP() の信頼境界（TRUSTED_PROXIES）。
	TrustedProxies string
	// GinMode は Gin の実行モード（GIN_MODE）。"release" のとき本番運用とみなす。
	GinMode string
}

// Validate は事故につながる設定漏れを起動時に検出する。
//
//   - IP_HASH_SALT が未設定: salt が既知の固定値のままだと ip_hash から元 IP を
//     復元されうるため、警告ログではなく起動失敗にする（気づいた時点で手遅れになる性質のため）。
//   - GIN_MODE=release（本番運用）で TRUSTED_PROXIES が未設定: リバースプロキシ/LB の
//     背後では ClientIP() が全リクエストで同一値になり、レート制限とファン数集計が
//     静かに壊れる。開発モードでは localhost 直公開が前提のため許容する。
func Validate(env Env) error {
	if env.IPHashSalt == "" {
		return fmt.Errorf("IP_HASH_SALT が未設定です。ip_hash の salt が推測可能な既定値のままになるため、必ず設定してください")
	}
	if env.GinMode == "release" && env.TrustedProxies == "" {
		return fmt.Errorf("TRUSTED_PROXIES が未設定です。GIN_MODE=release でリバースプロキシ/LB の背後に置く場合は、ClientIP() の信頼境界として必ず設定してください（未設定だと全ユーザーが同一IPとみなされ、レート制限とファン数集計が壊れます）")
	}
	return nil
}
