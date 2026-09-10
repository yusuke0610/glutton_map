package config

import "testing"

func TestValidate_IPHashSalt未設定はエラー(t *testing.T) {
	err := Validate(Env{IPHashSalt: "", TrustedProxies: "", GinMode: ""})
	if err == nil {
		t.Fatal("IP_HASH_SALT 未設定なのにエラーが返らない")
	}
}

func TestValidate_IPHashSaltがあれば開発モードは通る(t *testing.T) {
	err := Validate(Env{IPHashSalt: "salt", TrustedProxies: "", GinMode: ""})
	if err != nil {
		t.Errorf("Validate: %v", err)
	}
}

func TestValidate_releaseモードでTrustedProxies未設定はエラー(t *testing.T) {
	err := Validate(Env{IPHashSalt: "salt", TrustedProxies: "", GinMode: "release"})
	if err == nil {
		t.Fatal("release モードで TRUSTED_PROXIES 未設定なのにエラーが返らない")
	}
}

func TestValidate_releaseモードでもTrustedProxiesがあれば通る(t *testing.T) {
	err := Validate(Env{IPHashSalt: "salt", TrustedProxies: "10.0.0.0/8", GinMode: "release"})
	if err != nil {
		t.Errorf("Validate: %v", err)
	}
}
