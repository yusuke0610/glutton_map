import { useMemo, useState } from "react";
import { createPin, type Pin } from "../api/api";
import { logger } from "../lib/logger";
import { messages } from "../lib/messages";
import { getOrCreateAnonToken } from "../lib/anonToken";
import type { UTMParams } from "../lib/deeplink";
import { PREFECTURES, type Prefecture } from "../geo/prefectures";
import { MUNICIPALITIES } from "../geo/municipalities";
import { searchMunicipalities } from "../geo/municipality-search";
import { canSubmitPin } from "../pin/pin-form";

// 投稿フォームの共通スタイル。
const panelButtonStyle: React.CSSProperties = {
  background: "#d97b3a",
  color: "#fff",
  border: "none",
  borderRadius: 8,
  padding: "10px 14px",
  fontWeight: "bold",
  cursor: "pointer",
};
// 無効時（市区町村を候補から選んでいない等）は白くぼかして押せないことを伝える。
const panelButtonDisabledStyle: React.CSSProperties = {
  background: "#e8e2dc",
  color: "#fff",
  cursor: "not-allowed",
  opacity: 0.6,
  filter: "blur(0.4px)",
  boxShadow: "none",
};
const labelStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: 4,
  fontSize: 13,
  color: "#333",
};
const inputStyle: React.CSSProperties = {
  padding: "6px 8px",
  borderRadius: 6,
  border: "1px solid #ccc",
  font: "inherit",
};
// 市区町村のあいまい検索候補リスト（入力欄の下に重ねて表示）。
const suggestionListStyle: React.CSSProperties = {
  position: "absolute",
  top: "100%",
  left: 0,
  right: 0,
  margin: "2px 0 0",
  padding: 0,
  listStyle: "none",
  background: "#fff",
  border: "1px solid #ccc",
  borderRadius: 6,
  boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
  maxHeight: 180,
  overflowY: "auto",
  zIndex: 3,
};
const suggestionItemStyle: React.CSSProperties = {
  display: "block",
  width: "100%",
  textAlign: "left",
  background: "transparent",
  border: "none",
  padding: "8px 10px",
  font: "inherit",
  cursor: "pointer",
};
// 罰（×）の閉じるボタン。テキストではなくアイコン表示にする。
const closeButtonStyle: React.CSSProperties = {
  background: "transparent",
  border: "none",
  color: "#666",
  fontSize: 20,
  lineHeight: 1,
  cursor: "pointer",
  padding: 0,
  width: 24,
  height: 24,
};

// 投稿フォーム（右上のパネル）。閉じている間はトグルボタンのみ。
// 自身でフォーム状態を保持し、送信成功時に入力をリセットして onSubmitted(created) を呼ぶ。
// hidden=true（読み込みエラー時）は全幅エラーバナーと被るため描画しない。
export function PinForm({
  hidden,
  onSubmitted,
  initialOpen = false,
  utm = {},
}: {
  hidden: boolean;
  onSubmitted: (pin: Pin) => void;
  // initialOpen=true（ディープリンク post=1）のとき、着地直後にフォームを開いた状態にする。
  initialOpen?: boolean;
  // utm はディープリンクで受け取った流入元。投稿時に payload へ載せて計測する。
  utm?: UTMParams;
}) {
  const [formOpen, setFormOpen] = useState(initialOpen);
  const [nickname, setNickname] = useState("");
  const [prefecture, setPrefecture] = useState<Prefecture | "">("");
  const [city, setCity] = useState("");
  // 選択された市区町村の全国地方公共団体コード。空 = 未選択（自由入力のフォールバック）。
  const [municipalityCode, setMunicipalityCode] = useState("");
  // 市区町村入力のフォーカス状態（候補リストの表示制御）。
  const [cityFocused, setCityFocused] = useState(false);
  const [comment, setComment] = useState("");
  const [submitting, setSubmitting] = useState(false);
  // 投稿結果の通知（成功 or 失敗）。
  const [formNotice, setFormNotice] = useState<
    { kind: "success" | "error"; text: string } | null
  >(null);

  // 市区町村のあいまい検索候補。コード選択済み（municipalityCode!=""）のときは出さない。
  const citySuggestions = useMemo(
    () =>
      municipalityCode === ""
        ? searchMunicipalities(MUNICIPALITIES, prefecture, city, 8)
        : [],
    [prefecture, city, municipalityCode],
  );

  // 投稿: createPin で送信し、成功したら入力をリセット → onSubmitted で打ち込み演出へ。
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    // 市区町村は候補から選択（municipalityCode あり）されていないと投稿できない。
    // canSubmitPin は prefecture !== "" も確認するが、tsc に prefecture を絞り込ませる
    // ため明示の早期 return も置く（これがないと createPin に "" が渡りうると判定される）。
    if (!canSubmitPin({ prefecture, municipalityCode, submitting }) || prefecture === "")
      return;
    setSubmitting(true);
    setFormNotice(null);
    try {
      const created = await createPin({
        nickname,
        prefecture,
        city,
        municipality_code: municipalityCode,
        comment: comment || undefined,
        // 匿名トークンと流入元(utm)は投稿の瞬間にしか記録できないため必ず載せる。
        anon_token: getOrCreateAnonToken(),
        ...utm,
      });
      setFormNotice({ kind: "success", text: messages.form.success });
      // 入力をリセットし、打ち込み演出とともに投稿を地図へ反映する。
      setNickname("");
      setPrefecture("");
      setCity("");
      setMunicipalityCode("");
      setComment("");
      onSubmitted(created);
    } catch (err) {
      logger.error("ピンの投稿に失敗", err);
      setFormNotice({ kind: "error", text: messages.error.createPin });
    } finally {
      setSubmitting(false);
    }
  };

  if (hidden) return null;

  return (
    <div
      style={{
        position: "absolute",
        top: 16,
        right: 16,
        zIndex: 2,
        width: formOpen ? 280 : "auto",
      }}
    >
      {!formOpen ? (
        <button
          type="button"
          onClick={() => setFormOpen(true)}
          style={panelButtonStyle}
        >
          {messages.form.open}
        </button>
      ) : (
        <form
          onSubmit={handleSubmit}
          aria-label={messages.form.title}
          style={{
            background: "rgba(255,255,255,0.97)",
            borderRadius: 10,
            padding: 16,
            boxShadow: "0 2px 12px rgba(0,0,0,0.2)",
            display: "flex",
            flexDirection: "column",
            gap: 10,
          }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <strong>{messages.form.title}</strong>
            <button
              type="button"
              aria-label={messages.form.close}
              onClick={() => setFormOpen(false)}
              style={closeButtonStyle}
            >
              ×
            </button>
          </div>

          <label style={labelStyle}>
            {messages.form.nickname}
            <input
              type="text"
              required
              maxLength={30}
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              style={inputStyle}
            />
          </label>

          <label style={labelStyle}>
            {messages.form.prefecture}
            <select
              required
              value={prefecture}
              onChange={(e) => {
                setPrefecture(e.target.value as Prefecture);
                // 都道府県を変えたら市区町村の選択をリセットする。
                setCity("");
                setMunicipalityCode("");
              }}
              style={inputStyle}
            >
              <option value="" disabled>
                ―
              </option>
              {PREFECTURES.map((p) => (
                <option key={p} value={p}>
                  {p}
                </option>
              ))}
            </select>
          </label>

          <label style={labelStyle}>
            {messages.form.city}
            {/* 入力欄を他フォームと同じ幅に揃えるため flex で input を伸ばす。 */}
            <div
              style={{
                position: "relative",
                display: "flex",
                flexDirection: "column",
              }}
            >
              <input
                type="text"
                required
                maxLength={50}
                value={city}
                role="combobox"
                aria-expanded={cityFocused && citySuggestions.length > 0}
                aria-autocomplete="list"
                autoComplete="off"
                onChange={(e) => {
                  // 手入力したらコード選択を解除（候補から選び直す or 自由入力フォールバック）。
                  setCity(e.target.value);
                  setMunicipalityCode("");
                }}
                onFocus={() => setCityFocused(true)}
                // クリック確定（onMouseDown）を取りこぼさないよう遅延して閉じる。
                onBlur={() => setTimeout(() => setCityFocused(false), 120)}
                style={inputStyle}
              />
              {cityFocused && citySuggestions.length > 0 && (
                <ul role="listbox" style={suggestionListStyle}>
                  {citySuggestions.map((m) => (
                    <li key={m.code} role="option" aria-selected={false}>
                      <button
                        type="button"
                        // onMouseDown は input の onBlur より先に発火するので選択が確実に通る。
                        onMouseDown={(e) => {
                          e.preventDefault();
                          setCity(m.name);
                          setMunicipalityCode(m.code);
                          setCityFocused(false);
                        }}
                        // キーボード操作(Enter/Space)は click で発火するため、code 未設定で
                        // 都道府県フォールバックに落ちないよう同じ選択処理を割り当てる。
                        onClick={() => {
                          setCity(m.name);
                          setMunicipalityCode(m.code);
                          setCityFocused(false);
                        }}
                        style={suggestionItemStyle}
                      >
                        {m.name}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
            {city !== "" && municipalityCode === "" && (
              <span role="alert" style={{ color: "#a6471a", fontSize: 12 }}>
                {messages.form.cityHint}
              </span>
            )}
          </label>

          <label style={labelStyle}>
            {messages.form.comment}
            <textarea
              maxLength={200}
              rows={3}
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              // 縦のみリサイズ可とし、パディング込みでもパネル幅を超えないようにする。
              style={{
                ...inputStyle,
                resize: "vertical",
                boxSizing: "border-box",
                width: "100%",
                maxWidth: "100%",
              }}
            />
          </label>

          <button
            type="submit"
            disabled={!canSubmitPin({ prefecture, municipalityCode, submitting })}
            style={
              canSubmitPin({ prefecture, municipalityCode, submitting })
                ? panelButtonStyle
                : { ...panelButtonStyle, ...panelButtonDisabledStyle }
            }
          >
            {submitting ? messages.form.submitting : messages.form.submit}
          </button>

          {formNotice && (
            <div
              role={formNotice.kind === "error" ? "alert" : "status"}
              style={{
                color: formNotice.kind === "error" ? "#a6471a" : "#1a7a3a",
                fontSize: 13,
              }}
            >
              {formNotice.text}
            </div>
          )}
        </form>
      )}
    </div>
  );
}
