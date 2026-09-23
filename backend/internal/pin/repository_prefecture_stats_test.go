package pin

import (
	"context"
	"testing"
)

func TestSQLiteRepository_CountUniqueFansByPrefecture(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	in := []Pin{
		{Prefecture: "高知県", Lat: 33.5, Lng: 133.5, IPHash: "hashA"}, // 同一ファンの連投
		{Prefecture: "高知県", Lat: 33.6, Lng: 133.6, IPHash: "hashA"},
		{Prefecture: "高知県", Lat: 33.7, Lng: 133.7, IPHash: "hashB"},
		{Prefecture: "高知県", Lat: 33.8, Lng: 133.8, IPHash: ""}, // seed由来(ip_hash無し)は個別に数える
		{Prefecture: "高知県", Lat: 33.9, Lng: 133.9, IPHash: ""},
		{Prefecture: "東京都", Lat: 35.6, Lng: 139.7, IPHash: "hashC"},
	}
	for _, p := range in {
		if err := repo.Insert(ctx, p); err != nil {
			t.Fatalf("Insert(%v): %v", p, err)
		}
	}

	got, err := repo.CountUniqueFansByPrefecture(ctx, "高知県")
	if err != nil {
		t.Fatalf("CountUniqueFansByPrefecture: %v", err)
	}
	// hashA(1) + hashB(1) + 空hash2件(個別) = 4
	if want := 4; got != want {
		t.Errorf("CountUniqueFansByPrefecture(高知県) = %d, want %d", got, want)
	}

	got, err = repo.CountUniqueFansByPrefecture(ctx, "東京都")
	if err != nil {
		t.Fatalf("CountUniqueFansByPrefecture: %v", err)
	}
	if want := 1; got != want {
		t.Errorf("CountUniqueFansByPrefecture(東京都) = %d, want %d", got, want)
	}

	got, err = repo.CountUniqueFansByPrefecture(ctx, "北海道")
	if err != nil {
		t.Fatalf("CountUniqueFansByPrefecture: %v", err)
	}
	if want := 0; got != want {
		t.Errorf("CountUniqueFansByPrefecture(北海道) = %d, want %d", got, want)
	}
}
