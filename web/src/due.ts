export function isOverdue(due: string | null | undefined, today: string): boolean {
  if (!due) {
    return false;
  }
  return due.slice(0, 10) < today.slice(0, 10);
}

export function localToday(now = new Date()): string {
  const y = now.getFullYear();
  const m = String(now.getMonth() + 1).padStart(2, '0');
  const d = String(now.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}
