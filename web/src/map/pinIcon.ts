// ピン上部に載せるアイコンを生成する（マーカー型ピン＋上部に顔アイコン）。
// 顔の部分は今はプレースホルダ。公式の食いしんboy アイコンの許可が下りたら、
// 下の「公式アイコンに差し替える箇所」で画像を drawImage するだけで本番化できる。
export function createPinIcon(): {
  image: ImageData;
  options: { pixelRatio: number };
} {
  const w = 40;
  const h = 52;
  const ratio = 2;
  const canvas = document.createElement("canvas");
  canvas.width = w * ratio;
  canvas.height = h * ratio;
  const ctx = canvas.getContext("2d")!;
  ctx.scale(ratio, ratio);

  const color = "#d97b3a"; // ヒートマップと揃えた暖色
  // 下に伸びる尖り（先端が座標を指す）。
  ctx.beginPath();
  ctx.moveTo(9, 27);
  ctx.lineTo(20, 50);
  ctx.lineTo(31, 27);
  ctx.closePath();
  ctx.fillStyle = color;
  ctx.fill();
  // 頭（丸）。
  ctx.beginPath();
  ctx.arc(20, 18, 16, 0, Math.PI * 2);
  ctx.fillStyle = color;
  ctx.fill();
  ctx.lineWidth = 2;
  ctx.strokeStyle = "#ffffff";
  ctx.stroke();

  // --- 公式アイコンに差し替える箇所（今は「如」の文字）---
  ctx.fillStyle = "#ffffff";
  ctx.font = "bold 20px sans-serif";
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText("如", 20, 19);
  // --- ここまで ---

  return {
    image: ctx.getImageData(0, 0, canvas.width, canvas.height),
    options: { pixelRatio: ratio },
  };
}
