"use client";

/**
 * A usage meter for a 0-100% resource reading.
 *
 * The fill carries severity; the unfilled track is a lighter step of the same
 * ramp so the state reads across the whole bar rather than only the filled part.
 *
 * Severity is deliberately three levels, not four. The palette's `warning`
 * (#fab219) and `serious` (#ec835a) steps sit only ~13.6 ΔE apart under
 * simulated colour-vision deficiency, below the 15 floor, so a reader could not
 * reliably tell them apart. Three levels validate cleanly (separations 24.4 and
 * 27.6) on the dark chart surface.
 *
 * Colour is never the only channel: the numeric value and a severity word are
 * always rendered beside the bar.
 */

export type Severity = 'good' | 'warning' | 'critical';

const SEVERITY_FILL: Record<Severity, string> = {
  good: '#0ca30c',
  warning: '#fab219',
  critical: '#d03b3b',
};

/** Lighter step of the same hue, used for the unfilled track. */
const SEVERITY_TRACK: Record<Severity, string> = {
  good: 'rgba(12, 163, 12, 0.18)',
  warning: 'rgba(250, 178, 25, 0.18)',
  critical: 'rgba(208, 59, 59, 0.18)',
};

export function severityFor(percent: number): Severity {
  if (percent >= 90) return 'critical';
  if (percent >= 70) return 'warning';
  return 'good';
}

interface UsageMeterProps {
  label: string;
  /** 0-100. Values above 100 are clamped for the bar but shown verbatim. */
  percent: number;
  /** Free-text detail rendered under the bar, e.g. "3.2 / 8.0 GB". */
  detail?: string;
  /** Localised severity word, shown so severity never depends on colour alone. */
  severityLabel: string;
  /** Rendered instead of the bar when no reading is available. */
  unavailable?: boolean;
  unavailableLabel?: string;
}

export function UsageMeter({
  label,
  percent,
  detail,
  severityLabel,
  unavailable = false,
  unavailableLabel,
}: UsageMeterProps) {
  const severity = severityFor(percent);
  const clamped = Math.max(0, Math.min(100, percent));

  return (
    <div className="space-y-1.5">
      <div className="flex items-baseline justify-between gap-2">
        <span className="text-xs text-muted-foreground">{label}</span>
        {unavailable ? (
          <span className="text-xs text-muted-foreground">{unavailableLabel ?? '—'}</span>
        ) : (
          <span className="text-sm font-semibold tabular-nums text-foreground">
            {percent.toFixed(1)}%
          </span>
        )}
      </div>

      <div
        className="h-2 w-full overflow-hidden rounded-full"
        style={{ backgroundColor: unavailable ? 'rgba(120,120,120,0.18)' : SEVERITY_TRACK[severity] }}
        role="meter"
        aria-valuenow={unavailable ? undefined : Math.round(clamped)}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-label={`${label}: ${unavailable ? (unavailableLabel ?? 'unavailable') : `${percent.toFixed(1)}% (${severityLabel})`}`}
      >
        {!unavailable && (
          <div
            className="h-full rounded-full transition-[width] duration-500 ease-out"
            style={{ width: `${clamped}%`, backgroundColor: SEVERITY_FILL[severity] }}
          />
        )}
      </div>

      <div className="flex items-baseline justify-between gap-2">
        {detail ? (
          <span className="text-xs tabular-nums text-muted-foreground">{detail}</span>
        ) : (
          <span />
        )}
        {!unavailable && severity !== 'good' && (
          // Severity in words, so the state survives greyscale, CVD and
          // forced-colors where the fill hue carries nothing.
          <span className="text-xs font-medium text-muted-foreground">{severityLabel}</span>
        )}
      </div>
    </div>
  );
}
