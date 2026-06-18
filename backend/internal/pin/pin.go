package pin

// Prefecture は都道府県名（OpenAPI の Prefecture enum と同一語彙）。
type Prefecture string

// Pin は地図上の1点。最小スライスでは緯度経度はただのカラム。
type Pin struct {
	Prefecture Prefecture
	Lat        float64
	Lng        float64
}
