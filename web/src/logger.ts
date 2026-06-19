// フロント用の薄いロガー。console をラップし、環境に応じて出力レベルを絞る。
// 判定ロジック（thresholdFor / shouldLog）は純粋関数として切り出し、テスト可能にしている。

export type LogLevel = "debug" | "info" | "warn" | "error";

// レベルの重み。大きいほど深刻。
const ORDER: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
};

type LogEnv = {
  PROD: boolean;
  VITE_LOG_LEVEL?: string;
};

// thresholdFor は出力する最小レベルを決める。
// VITE_LOG_LEVEL が妥当ならそれを優先し、無ければ本番は warn / 開発は debug。
export function thresholdFor(env: LogEnv): LogLevel {
  const explicit = env.VITE_LOG_LEVEL;
  if (explicit && explicit in ORDER) {
    return explicit as LogLevel;
  }
  return env.PROD ? "warn" : "debug";
}

// shouldLog は level が閾値以上（同じか深刻）なら true。
export function shouldLog(level: LogLevel, threshold: LogLevel): boolean {
  return ORDER[level] >= ORDER[threshold];
}

const threshold = thresholdFor(import.meta.env);

function log(level: LogLevel, ...args: unknown[]): void {
  if (!shouldLog(level, threshold)) {
    return;
  }
  // console のメソッドにそのまま流す（debug→console.debug 等）。
  console[level](...args);
}

export const logger = {
  debug: (...args: unknown[]) => log("debug", ...args),
  info: (...args: unknown[]) => log("info", ...args),
  warn: (...args: unknown[]) => log("warn", ...args),
  error: (...args: unknown[]) => log("error", ...args),
};
