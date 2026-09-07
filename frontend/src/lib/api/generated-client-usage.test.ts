import { describe, expect, it } from "vitest";
import source from "./client.ts?raw";
import analyticsStoreSource from "../stores/analytics.svelte.ts?raw";
import messagesStoreSource from "../stores/messages.svelte.ts?raw";
import searchStoreSource from "../stores/search.svelte.ts?raw";
import sessionsStoreSource from "../stores/sessions.svelte.ts?raw";
import trendsStoreSource from "../stores/trends.svelte.ts?raw";
import usageStoreSource from "../stores/usage.svelte.ts?raw";

describe("API client implementation", () => {
  it("keeps generated JSON endpoint calls out of the bespoke client module", () => {
    expect(source).not.toContain("from \"./generated/index\"");
    expect(source).not.toContain("function fetchJSON");
    expect(source).not.toContain("function buildQuery");
    expect(source).not.toContain("export function listSessions");
    expect(source).not.toContain("export function getSession");
    expect(source).not.toContain("export function getMessages");
    expect(source).not.toContain("export function search(");
    expect(source).not.toContain("export function getSettings");
    expect(source).not.toContain("export function listPins");
    expect(source).not.toContain("export function listStarred");
    expect(source).not.toContain("getAnalyticsSummary");
    expect(source).not.toContain("getTrendsTerms");
    expect(source).not.toContain("getUsageSummary");

    expect(sessionsStoreSource).toContain(
      "SessionsService.getApiV1Sessions",
    );
    expect(messagesStoreSource).toContain(
      "SessionsService.getApiV1SessionsIdMessages",
    );
    expect(searchStoreSource).toContain(
      "SearchService.getApiV1Search",
    );
    expect(analyticsStoreSource).toContain(
      "AnalyticsService.getApiV1AnalyticsSummary",
    );
    expect(trendsStoreSource).toContain(
      "TrendsService.getApiV1TrendsTerms",
    );
    expect(usageStoreSource).toContain(
      "UsageService.getApiV1UsageSummary",
    );
  });
});
