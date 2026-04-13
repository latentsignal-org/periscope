export interface TreemapInput {
  id: string;
  value: number;
}

export interface TreemapTile {
  id: string;
  value: number;
  x: number;
  y: number;
  width: number;
  height: number;
}

/**
 * Squarified treemap layout (Bruls/Huijsen/van Wijk 2000).
 *
 * Filters to positive values, sorts descending, scales so values
 * sum to total area, then recursively lays out rows along the
 * shorter edge using the worst-aspect-ratio heuristic.
 */
export function squarify(
  input: readonly TreemapInput[],
  width: number,
  height: number,
): TreemapTile[] {
  const items = input
    .filter((d) => d.value > 0)
    .sort((a, b) => b.value - a.value);

  if (items.length === 0 || width <= 0 || height <= 0) return [];

  const totalValue = items.reduce((s, d) => s + d.value, 0);
  const totalArea = width * height;
  const scaled = items.map((d) => ({
    id: d.id,
    area: (d.value / totalValue) * totalArea,
    value: d.value,
  }));

  const tiles: TreemapTile[] = [];
  layoutRect(scaled, 0, 0, width, height, tiles);
  return tiles;
}

interface ScaledItem {
  id: string;
  area: number;
  value: number;
}

function worstRatio(
  row: ScaledItem[],
  sideLength: number,
): number {
  const rowArea = row.reduce((s, d) => s + d.area, 0);
  let worst = 0;
  for (const item of row) {
    const w = rowArea / sideLength;
    const h = item.area / w;
    const ratio = Math.max(w / h, h / w);
    if (ratio > worst) worst = ratio;
  }
  return worst;
}

function layoutRect(
  items: ScaledItem[],
  x: number,
  y: number,
  w: number,
  h: number,
  tiles: TreemapTile[],
): void {
  if (items.length === 0) return;

  if (items.length === 1) {
    const d = items[0]!;
    tiles.push({
      id: d.id,
      value: d.value,
      x,
      y,
      width: w,
      height: h,
    });
    return;
  }

  const shortSide = Math.min(w, h);
  const row: ScaledItem[] = [items[0]!];
  let idx = 1;

  while (idx < items.length) {
    const candidate = [...row, items[idx]!];
    if (worstRatio(candidate, shortSide) <=
        worstRatio(row, shortSide)) {
      row.push(items[idx]!);
      idx++;
    } else {
      break;
    }
  }

  const rowArea = row.reduce((s, d) => s + d.area, 0);
  const horizontal = w >= h;
  const rowSpan = rowArea / (horizontal ? h : w);

  let offset = 0;
  for (const item of row) {
    const itemSpan = item.area / rowSpan;
    if (horizontal) {
      tiles.push({
        id: item.id,
        value: item.value,
        x,
        y: y + offset,
        width: rowSpan,
        height: itemSpan,
      });
    } else {
      tiles.push({
        id: item.id,
        value: item.value,
        x: x + offset,
        y,
        width: itemSpan,
        height: rowSpan,
      });
    }
    offset += itemSpan;
  }

  const remaining = items.slice(idx);
  if (remaining.length > 0) {
    if (horizontal) {
      layoutRect(
        remaining, x + rowSpan, y, w - rowSpan, h, tiles,
      );
    } else {
      layoutRect(
        remaining, x, y + rowSpan, w, h - rowSpan, tiles,
      );
    }
  }
}
