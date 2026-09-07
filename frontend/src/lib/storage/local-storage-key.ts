export type MigratingStorage = Pick<Storage, "getItem" | "setItem"> &
  Partial<Pick<Storage, "removeItem">>;

export function legacyLocalStorageKey(key: string): string {
  if (!key.startsWith("periscope-")) return key;
  return `agentsview-${key.slice("periscope-".length)}`;
}

/** Read a value from `key`, migrating a one-time legacy `agentsview-*` copy. */
export function readMigratedLocalStorageValue(
  storage: MigratingStorage | null | undefined,
  key: string,
): string | null {
  if (!storage) return null;
  try {
    const current = storage.getItem(key);
    if (current !== null) return current;

    const legacyKey = legacyLocalStorageKey(key);
    if (legacyKey === key) return null;

    const legacy = storage.getItem(legacyKey);
    if (legacy === null) return null;

    storage.setItem(key, legacy);
    storage.removeItem?.(legacyKey);
    return legacy;
  } catch {
    return null;
  }
}

/** Remove both the current and legacy keys for a migrated storage entry. */
export function removeMigratedLocalStorageValue(
  storage: MigratingStorage | null | undefined,
  key: string,
): void {
  if (!storage) return;
  try {
    storage.removeItem?.(key);
    const legacyKey = legacyLocalStorageKey(key);
    if (legacyKey !== key) storage.removeItem?.(legacyKey);
  } catch {
    // ignore
  }
}
