// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { mount, tick, unmount } from "svelte";
// @ts-ignore
import CompactSignalBanner from "./CompactSignalBanner.svelte";

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

describe("CompactSignalBanner", () => {
  it("renders generated preserve/drop copy and copies the focus text", async () => {
    const component = mount(CompactSignalBanner, {
      target: document.body,
      props: {
        signal: {
          should_compact: true,
          confidence: "high",
          reasons: ["High occupancy"],
          score: 79,
          estimated_reclaimable: 18000,
          keep_items: ["auth refactor plan"],
          drop_items: ["deployment warning branch"],
          compact_focus_text:
            "Preserve the auth refactor plan and drop the deployment warning branch.",
          focus_provenance: "model-generated",
          focus_model: "claude-haiku-4-5-20251001",
        },
      },
    });

    await tick();
    expect(document.body.textContent).toContain(
      "Preserve auth refactor plan, drop deployment warning branch",
    );

    const toggle = document.querySelector(
      ".toggle-btn",
    ) as HTMLButtonElement | null;
    toggle?.click();
    await tick();

    const copyButton = document.querySelector(
      '[aria-label="Copy suggested compact focus"]',
    ) as HTMLButtonElement | null;
    copyButton?.click();
    await tick();

    expect(mockCopyToClipboard).toHaveBeenCalledWith(
      "Preserve the auth refactor plan and drop the deployment warning branch.",
    );

    unmount(component);
  });

  it("renders the star hint when summaries are idle", async () => {
    const component = mount(CompactSignalBanner, {
      target: document.body,
      props: {
        signal: {
          should_compact: true,
          confidence: "medium",
          reasons: ["High stale context ratio"],
          score: 48,
          estimated_reclaimable: 9000,
        },
        summaryCoverage: {
          status: "idle",
          total_turns: 18,
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
