import { describe, expect, it } from "vitest";
import {
  legacyLocalStorageKey,
  readMigratedLocalStorageValue,
  removeMigratedLocalStorageValue,
} from "./local-storage-key.js";

function memoryStorage(initial: Record<string, string> = {}) {
  const values = new Map(Object.entries(initial));
  return {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => {
      values.set(key, value);
    },
    removeItem: (key: string) => {
      values.delete(key);
    },
    values,
  };
}

describe("local storage key migration", () => {
  it("maps periscope keys to their agentsview legacy names", () => {
    expect(legacyLocalStorageKey("periscope-search-mode")).toBe(
      "agentsview-search-mode",
    );
    expect(legacyLocalStorageKey("agentsview-search-mode")).toBe(
      "agentsview-search-mode",
    );
  });

  it("migrates a legacy value into the periscope key once", () => {
    const storage = memoryStorage({
      "agentsview-search-mode": "semantic",
    });

    expect(
      readMigratedLocalStorageValue(storage, "periscope-search-mode"),
    ).toBe("semantic");
    expect(storage.values.get("periscope-search-mode")).toBe("semantic");
    expect(storage.values.has("agentsview-search-mode")).toBe(false);
  });

  it("prefers the periscope key when both keys exist", () => {
    const storage = memoryStorage({
      "periscope-search-mode": "hybrid",
      "agentsview-search-mode": "semantic",
    });

    expect(
      readMigratedLocalStorageValue(storage, "periscope-search-mode"),
    ).toBe("hybrid");
    expect(storage.values.get("agentsview-search-mode")).toBe("semantic");
  });

  it("removes both current and legacy keys", () => {
    const storage = memoryStorage({
      "periscope-starred-sessions": '["a"]',
      "agentsview-starred-sessions": '["b"]',
    });

    removeMigratedLocalStorageValue(storage, "periscope-starred-sessions");

    expect(storage.values.size).toBe(0);
  });
});
