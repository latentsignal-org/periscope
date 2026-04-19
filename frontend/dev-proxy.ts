export const DEFAULT_DEV_PROXY_TARGET = "http://127.0.0.1:8080";

export function getDevProxyTarget(env: Record<string, string | undefined>): string {
  return env.AGENTSVIEW_DEV_PROXY_TARGET || DEFAULT_DEV_PROXY_TARGET;
}

export function getProxyOrigin(target: string): string {
  return new URL(target).origin;
}

export function applyDevProxyHeaders(
  proxyReq: { setHeader(name: string, value: string): void },
  target: string,
) {
  proxyReq.setHeader("Origin", getProxyOrigin(target));
}