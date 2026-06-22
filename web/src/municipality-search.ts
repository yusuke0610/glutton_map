import type { Municipality } from "./municipalities";

// normalize は全半角ゆれ・大文字小文字・前後空白を吸収する。
function normalize(s: string): string {
  return s.normalize("NFKC").trim().toLowerCase();
}

// toHiragana はカタカナをひらがなに変換する（かな検索の表記ゆれ対策）。
function toHiragana(s: string): string {
  return s.replace(/[ァ-ヶ]/g, (ch) =>
    String.fromCharCode(ch.charCodeAt(0) - 0x60),
  );
}

// searchMunicipalities は市区町村をあいまい検索する。
// prefecture を指定するとその都道府県に絞り込む（空なら全国）。
// 名称（漢字）の部分一致、または読みかなの部分一致でヒットする。
export function searchMunicipalities(
  list: readonly Municipality[],
  prefecture: string,
  query: string,
  limit = 20,
): Municipality[] {
  if (limit <= 0) return [];
  const nq = normalize(query);
  if (nq === "") return [];
  const hq = toHiragana(nq);

  const result: Municipality[] = [];
  for (const m of list) {
    if (prefecture !== "" && m.prefecture !== prefecture) continue;
    const name = normalize(m.name);
    const kana = toHiragana(normalize(m.kana));
    if (name.includes(nq) || kana.includes(hq)) {
      result.push(m);
      if (result.length >= limit) break;
    }
  }
  return result;
}
