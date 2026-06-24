// ディープリンク: X 等から踏んだ人が「トップ」ではなく「刺せる状態のマップ」に
// 直接着地できるよう、エントリ URL のクエリでセンタリング座標・初期ズーム・投稿モード起動・
// 流入元(utm)を受け取る。値はすべて任意。

// UTMParams は流入元の計測値。投稿時に createPin の payload へ載せる。
export interface UTMParams {
  utm_source?: string;
  utm_medium?: string;
  utm_campaign?: string;
}

export interface DeepLink {
  lat?: number;
  lng?: number;
  zoom?: number;
  // openForm は post=1（または post=true）のとき true。着地直後に投稿フォームを開く。
  openForm: boolean;
  // prefecture は共有リンクの pref。着地時の都道府県プリセット用（任意）。
  prefecture?: string;
  utm: UTMParams;
}

// parseDeepLink は location.search 相当のクエリ文字列を解釈する純粋関数。
export function parseDeepLink(search: string): DeepLink {
  const params = new URLSearchParams(search);

  const post = params.get("post");
  const pref = params.get("pref");

  const utm: UTMParams = {};
  for (const key of ["utm_source", "utm_medium", "utm_campaign"] as const) {
    const v = params.get(key);
    if (v) utm[key] = v;
  }

  return {
    lat: finiteNumber(params.get("lat")),
    lng: finiteNumber(params.get("lng")),
    zoom: finiteNumber(params.get("zoom")),
    openForm: post === "1" || post === "true",
    prefecture: pref ?? undefined,
    utm,
  };
}

// finiteNumber は数値として解釈できる文字列のみ number を返す（NaN/空は undefined）。
function finiteNumber(raw: string | null): number | undefined {
  if (raw === null || raw.trim() === "") return undefined;
  const n = Number(raw);
  return Number.isFinite(n) ? n : undefined;
}
