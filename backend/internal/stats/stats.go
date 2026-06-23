// Package stats は提出用のユニークファン集計を担う。
// 地図表示（pin.Summarize / GetApiPins）とは独立した経路で、
// 同一 ip_hash の連投や悪意ある curl を重複排除した「ファン数」を算出する。
package stats

import "github.com/kisaragi-ai-map/backend/internal/pin"

// Report はくいしんぼ如月へ提出するための集計結果。
type Report struct {
	// TotalPosts は投稿数の合計（重複込み・地図の total 相当）。
	TotalPosts int `json:"total_posts"`
	// UniqueFans は ip_hash で重複排除したファン数。事業判断のヒーロー指標。
	UniqueFans int `json:"unique_fans"`
	// PrefectureCount はファンが存在する都道府県数。
	PrefectureCount int `json:"prefecture_count"`
	// ByPrefecture は都道府県ごとのユニークファン数（地域内訳）。
	// 注意: 1人が複数県から投稿すると各県でカウントされ得るため、
	// 合計は UniqueFans と必ずしも一致しない。
	ByPrefecture map[string]int `json:"by_prefecture"`
}

// Build は集計用の行から Report を計算する純粋関数。
// ユニーク化のルール:
//   - ip_hash が非空なら、その値で同一ファンとみなす（連投・curl を1人に畳む）。
//   - ip_hash が空（seed/レガシー由来）の行は、畳まずに各行を1ファンとして数える。
func Build(rows []pin.PinStat) Report {
	// 都道府県ごとに「非空ハッシュの集合」と「空ハッシュ行の件数」を持つ。
	type bucket struct {
		hashes map[string]struct{}
		empty  int
	}
	byPref := map[string]*bucket{}
	globalHashes := map[string]struct{}{}
	globalEmpty := 0

	for _, r := range rows {
		pref := string(r.Prefecture)
		b := byPref[pref]
		if b == nil {
			b = &bucket{hashes: map[string]struct{}{}}
			byPref[pref] = b
		}
		if r.IPHash == "" {
			b.empty++
			globalEmpty++
			continue
		}
		b.hashes[r.IPHash] = struct{}{}
		globalHashes[r.IPHash] = struct{}{}
	}

	byPrefecture := make(map[string]int, len(byPref))
	for pref, b := range byPref {
		byPrefecture[pref] = len(b.hashes) + b.empty
	}

	return Report{
		TotalPosts:      len(rows),
		UniqueFans:      len(globalHashes) + globalEmpty,
		PrefectureCount: len(byPref),
		ByPrefecture:    byPrefecture,
	}
}
