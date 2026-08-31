export function sortOrderForDrop(
  column: { identifier: string; sortOrder: number }[],
  dragId: string,
  beforeId: string | null,
): number {
  const rest = column.filter((c) => c.identifier !== dragId);
  if (!beforeId) {
    const last = rest[rest.length - 1];
    return last ? last.sortOrder + 1 : 1;
  }
  const idx = rest.findIndex((c) => c.identifier === beforeId);
  if (idx <= 0) {
    const first = rest[0];
    return first ? first.sortOrder / 2 : 1;
  }
  const prev = rest[idx - 1];
  const target = rest[idx];
  if (!prev || !target) {
    return 1;
  }
  return (prev.sortOrder + target.sortOrder) / 2;
}
