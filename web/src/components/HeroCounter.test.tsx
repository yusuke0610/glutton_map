import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { HeroCounter } from "./HeroCounter";
import { PREFECTURES } from "../geo/prefectures";

describe("HeroCounter", () => {
  it("uniqueFansとprefectureCountの両方を表示する", () => {
    render(<HeroCounter uniqueFans={1234} prefectureCount={32} />);

    expect(screen.getByText(/1,234/)).toBeInTheDocument();
    expect(screen.getByText(/32/)).toBeInTheDocument();
    expect(screen.getByText(new RegExp(String(PREFECTURES.length)))).toBeInTheDocument();
  });
});
