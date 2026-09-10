import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PinForm } from "./PinForm";

describe("PinForm", () => {
  it("都道府県を変更すると市区町村の入力がリセットされる", async () => {
    const user = userEvent.setup();
    render(<PinForm hidden={false} onSubmitted={vi.fn()} initialOpen />);

    const cityInput = screen.getByLabelText("市区町村");
    await user.type(cityInput, "横浜市");
    expect(cityInput).toHaveValue("横浜市");

    const prefectureSelect = screen.getByLabelText("都道府県");
    await user.selectOptions(prefectureSelect, "北海道");

    expect(cityInput).toHaveValue("");
  });
});
