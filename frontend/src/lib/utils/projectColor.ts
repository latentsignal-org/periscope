export const PROJECT_PALETTE: readonly string[] = [
  "var(--accent-blue)",
  "var(--accent-purple)",
  "var(--accent-amber)",
  "var(--accent-teal)",
  "var(--accent-rose)",
  "var(--accent-green)",
  "var(--accent-indigo)",
  "var(--accent-orange)",
  "var(--accent-sky)",
  "var(--accent-pink)",
  "var(--accent-coral)",
  "var(--accent-lime)",
] as const;

const FALLBACK = "var(--text-muted)";

function djb2(s: string): number {
  let h = 5381;
  for (let i = 0; i < s.length; i++) {
    h = ((h << 5) + h + s.charCodeAt(i)) | 0;
  }
  return Math.abs(h);
}

export function projectColor(name: string): string {
  if (!name) return FALLBACK;
  return PROJECT_PALETTE[djb2(name) % PROJECT_PALETTE.length]!;
}
