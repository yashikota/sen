const stampOpts: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
};

export function formatStamp(iso: string, timeZone: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) {
    return iso;
  }
  try {
    return new Intl.DateTimeFormat('en-CA', { ...stampOpts, timeZone }).format(d);
  } catch {
    return new Intl.DateTimeFormat('en-CA', { ...stampOpts, timeZone: 'UTC' }).format(d);
  }
}
