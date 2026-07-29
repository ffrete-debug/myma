/**
 * Minimal INI parse/merge that PRESERVES everything it does not understand.
 *
 * The visual config editors used to rebuild GameUserSettings.ini and Game.ini
 * from their own curated parameter catalogue. ARK's real files contain hundreds
 * of keys that are not in that catalogue, so touching the visual editor once
 * silently dropped every one of them - the editor replaced the server's real
 * configuration with its own subset instead of reflecting it.
 *
 * These helpers instead treat the real file as the source of truth and apply the
 * editor's values on top, leaving unknown keys, sections, comments, ordering and
 * duplicate keys untouched.
 */

export type IniValue = string | number | boolean;

interface Line {
  /** Verbatim source line, used when nothing about it changes. */
  raw: string;
  /** Section this line belongs to ('' before any header). */
  section: string;
  /** Key, when the line is a key=value assignment. */
  key?: string;
}

const SECTION_RE = /^\s*\[([^\]]+)\]\s*$/;
// ARK INIs use key=value; comments start with ; or #. A key may legitimately
// contain [] (e.g. per-level arrays), so only match up to the first '='.
const ENTRY_RE = /^\s*([^;#=\s][^=]*?)\s*=(.*)$/;

/** Sections are compared case-insensitively: ARK writes
 *  [/Script/ShooterGame.ShooterGameMode] while docs and code often use other
 *  casings, and a mismatch would silently create a duplicate section. */
function sameSection(a: string, b: string): boolean {
  return a.toLowerCase() === b.toLowerCase();
}

/**
 * ARK runs under Proton and writes its INI files with CRLF. Splitting on '\n'
 * alone leaves a trailing '\r' on every line, so section headers and keys never
 * match - which duplicated keys instead of updating them and made reads return
 * undefined. Split on either, and remember which ending to write back.
 */
function splitLines(text: string): string[] {
  return text.split(/\r?\n/);
}

/** The line ending the file already uses, so merging does not rewrite the whole
 *  file into a different style. */
function detectEol(text: string): string {
  return text.includes('\r\n') ? '\r\n' : '\n';
}

function parse(text: string): Line[] {
  const lines: Line[] = [];
  let section = '';

  for (const raw of splitLines(text)) {
    const sectionMatch = raw.match(SECTION_RE);
    if (sectionMatch) {
      section = `[${sectionMatch[1].trim()}]`;
      lines.push({ raw, section });
      continue;
    }
    const entryMatch = raw.match(ENTRY_RE);
    lines.push(entryMatch ? { raw, section, key: entryMatch[1].trim() } : { raw, section });
  }

  return lines;
}

function formatValue(value: IniValue): string {
  if (typeof value === 'boolean') return value ? 'True' : 'False';
  return String(value);
}

/**
 * Apply `overrides` to `original`, keeping every other line exactly as it was.
 *
 * A key already present in the target section is updated in place, so ordering
 * and surrounding comments survive. A key that is missing is appended to the end
 * of its section, and a section that does not exist yet is created at the end.
 *
 * Keys are matched case-insensitively, because ARK is inconsistent about casing
 * between what it writes and what the documentation uses.
 */
export function mergeIni(
  original: string,
  overrides: Record<string, IniValue | undefined>,
  section: string,
): string {
  const wanted = new Map<string, { key: string; value: IniValue }>();
  for (const [key, value] of Object.entries(overrides)) {
    if (value === undefined || value === '') continue;
    wanted.set(key.toLowerCase(), { key, value });
  }
  if (wanted.size === 0) return original;

  const eol = detectEol(original);
  const lines = parse(original);
  const applied = new Set<string>();

  // 1. Update in place where the key already exists in the target section.
  for (const line of lines) {
    if (!sameSection(line.section, section) || !line.key) continue;
    const hit = wanted.get(line.key.toLowerCase());
    if (!hit || applied.has(line.key.toLowerCase())) continue;
    line.raw = `${line.key}=${formatValue(hit.value)}`;
    applied.add(line.key.toLowerCase());
  }

  // 2. Append anything still unapplied to the end of its section.
  const missing = [...wanted.entries()].filter(([k]) => !applied.has(k));
  if (missing.length === 0) return lines.map((l) => l.raw).join(eol);

  let lastIndexOfSection = -1;
  for (let i = 0; i < lines.length; i++) {
    if (sameSection(lines[i].section, section)) lastIndexOfSection = i;
  }

  const additions = missing.map(([, v]) => `${v.key}=${formatValue(v.value)}`);

  if (lastIndexOfSection === -1) {
    // Section absent entirely: create it at the end.
    const body = lines.map((l) => l.raw).join(eol).replace(/\s*$/, '');
    return [body, '', section, ...additions, ''].join(eol);
  }

  const out = lines.map((l) => l.raw);
  out.splice(lastIndexOfSection + 1, 0, ...additions);
  return out.join(eol);
}

/**
 * Read the values of `keys` from a section, for populating the visual editor
 * from the real file.
 */
export function readIniSection(text: string, section: string): Record<string, string> {
  const values: Record<string, string> = {};
  let current = '';

  for (const raw of splitLines(text)) {
    const sectionMatch = raw.match(SECTION_RE);
    if (sectionMatch) {
      current = `[${sectionMatch[1].trim()}]`;
      continue;
    }
    if (!sameSection(current, section)) continue;
    const entryMatch = raw.match(ENTRY_RE);
    if (entryMatch) values[entryMatch[1].trim()] = entryMatch[2].trim();
  }

  return values;
}

/** Every section header present, in order. */
export function iniSections(text: string): string[] {
  const seen: string[] = [];
  for (const raw of splitLines(text)) {
    const m = raw.match(SECTION_RE);
    if (m) {
      const name = `[${m[1].trim()}]`;
      if (!seen.includes(name)) seen.push(name);
    }
  }
  return seen;
}
