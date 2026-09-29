export type BusyPeriod = { start: string | Date; end: string | Date };

export function freeWindow(busy: BusyPeriod[], now: Date) {
  const start = now.getTime() + 2 * 60_000;
  let end = start + 90 * 60_000;
  for (const period of busy) {
    const busyStart = new Date(period.start).getTime();
    const busyEnd = new Date(period.end).getTime();
    if (!Number.isFinite(busyStart) || !Number.isFinite(busyEnd) || busyEnd <= start || busyStart >= end) continue;
    if (busyStart <= start) return [];
    end = Math.min(end, busyStart);
  }
  return end - start >= 20 * 60_000 ? [{ start: new Date(start).toISOString(), end: new Date(end).toISOString() }] : [];
}
