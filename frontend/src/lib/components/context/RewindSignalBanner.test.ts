// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { mount, tick, unmount } from "svelte";
// @ts-ignore
import RewindSignalBanner from "./RewindSignalBanner.svelte";

const { mockCopyToClipboard } = vi.hoisted(() => ({
  mockCopyToClipboard: vi.fn().mockResolvedValue(true),
}));

vi.mock("../../utils/clipboard.js", () => ({
  copyToClipboard: mockCopyToClipboard,
}));

afterEach(() => {
  document.body.innerHTML = "";
  mockCopyToClipboard.mockClear();
});

describe("RewindSignalBanner", () => {
  it("renders the suggested prompt block and copies generated text", async () => {
    const component = mount(RewindSignalBanner, {
      target: document.body,
      props: {
        signal: {
          should_rewind: true,
          confidence: "high",
          reasons: ["Low-value tangent"],
          tokens_recoverable: 3200,
          score: 82,
          rewind_to_turn: 12,
          bad_stretch_from: 13,
          bad_stretch_to: 15,
          tangent_label: "deployment warning branch",
          rewind_reprompt_text:
            "Ignore the deployment warning branch and resume the auth refactor.",
          reprompt_provenance: "model-generated",
          reprompt_model: "claude-haiku-4-5-20251001",
        },
      },
    });

    await tick();
    const toggle = document.querySelector(
      ".toggle-btn",
    ) as HTMLButtonElement | null;
    expect(toggle?.textContent).toContain("Suggested rewind prompt");
    toggle?.click();
    await tick();

    expect(document.querySelector(".suggestion-body")?.textContent).toContain(
      "resume the auth refactor",
    );

    const copyButton = document.querySelector(
      '[aria-label="Copy suggested rewind prompt"]',
    ) as HTMLButtonElement | null;
    copyButton?.click();
    await tick();

    expect(mockCopyToClipboard).toHaveBeenCalledWith(
      "Ignore the deployment warning branch and resume the auth refactor.",
    );

    unmount(component);
  });

  it("renders the star hint when no generated text is available yet", async () => {
    const component = mount(RewindSignalBanner, {
      target: document.body,
      props: {
        signal: {
          should_rewind: true,
          confidence: "medium",
          reasons: ["Last turn looks low-value"],
          tokens_recoverable: 1200,
          score: 44,
          rewind_to_turn: 8,
        },
        summaryCoverage: {
          status: "idle",
          total_turns: 12,
          summarised_turns: 0,
          starred: false,
        },
      },
    });

    await tick();
    expect(document.body.textContent).toContain("Star to enable guidance text");
    expect(document.querySelector(".suggestion-body")).toBeNull();

    unmount(component);
  });
});
