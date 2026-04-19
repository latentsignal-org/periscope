import { describe, expect, it, vi } from "vitest";
import {
  applyDevProxyHeaders,
  DEFAULT_DEV_PROXY_TARGET,
  getDevProxyTarget,
  getProxyOrigin,
} from "./dev-proxy";

describe("dev proxy helpers", () => {
  it("uses the default backend target when no env override is set", () => {
    expect(getDevProxyTarget({})).toBe(DEFAULT_DEV_PROXY_TARGET);
  });

  it("uses the env override when provided", () => {
    expect(
      getDevProxyTarget({
        AGENTSVIEW_DEV_PROXY_TARGET: "http://127.0.0.1:8081",
      }),
    ).toBe("http://127.0.0.1:8081");
  });

  it("derives the origin from the proxy target", () => {
    expect(getProxyOrigin("http://127.0.0.1:8081/api")).toBe(
      "http://127.0.0.1:8081",
    );
  });

  it("rewrites the forwarded Origin header to the backend origin", () => {
    const proxyReq = {
      setHeader: vi.fn(),
    };

    applyDevProxyHeaders(proxyReq, "http://127.0.0.1:8081");

    expect(proxyReq.setHeader).toHaveBeenCalledWith(
      "Origin",
      "http://127.0.0.1:8081",
    );
  });
});