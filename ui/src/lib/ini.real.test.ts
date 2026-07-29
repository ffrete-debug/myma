import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { join } from 'path';
import { mergeIni, readIniSection } from './ini';

// The genuine GameUserSettings.ini written by a real ARK server, captured from
// the test VM. 219 keys, almost none of which the visual editor knows about.
// Rebuilding the file from the editor's catalogue destroyed all of them; this
// pins that the merge does not.
const REAL = readFileSync(join(__dirname, '__fixtures__', 'real-gameusersettings.ini'), 'utf8');

const keyCount = (text: string) => text.split('\n').filter((l) => /^[^;#\[\s][^=]*=/.test(l)).length;

describe('merging into a real ARK GameUserSettings.ini', () => {
  it('starts from a file with the full key set', () => {
    expect(keyCount(REAL)).toBeGreaterThan(200);
  });

  it('preserves every key when the visual editor changes an existing value', () => {
    const before = keyCount(REAL);
    // ServerPVE genuinely exists in [ServerSettings] in the real file.
    const out = mergeIni(REAL, { ServerPVE: false }, '[ServerSettings]');

    expect(keyCount(out)).toBe(before);
    expect(out).toContain('ServerPVE=False');
  });

  // ARK keeps MaxPlayers in [/Script/Engine.GameSession], not [ServerSettings],
  // so it must be written there or it has no effect on the server.
  it('updates MaxPlayers in the GameSession section without duplicating it', () => {
    const before = keyCount(REAL);
    const out = mergeIni(REAL, { MaxPlayers: 55 }, '[/Script/Engine.GameSession]');

    expect(keyCount(out)).toBe(before);
    expect(out).toContain('MaxPlayers=55');
    expect(out.split(/\r?\n/).filter((l) => /^MaxPlayers=/.test(l))).toHaveLength(1);
  });

  it('preserves the specific ARK-written keys the editor does not model', () => {
    const out = mergeIni(REAL, { MaxPlayers: 55, ServerPVE: false }, '[ServerSettings]');

    for (const key of [
      'TheMaxStructuresInRange',
      'OxygenSwimSpeedStatMultiplier',
      'StructurePreventResourceRadiusMultiplier',
    ]) {
      expect(out).toContain(key);
    }
    expect(out).toContain('ServerPVE=False');
  });

  it('keeps every section that was in the real file', () => {
    const sectionsOf = (t: string) => (t.match(/^\[.+\]$/gm) ?? []).sort();
    const out = mergeIni(REAL, { MaxPlayers: 55 }, '[ServerSettings]');
    expect(sectionsOf(out)).toEqual(sectionsOf(REAL));
  });

  it('reads real values back for the visual editor to display', () => {
    const values = readIniSection(REAL, '[ServerSettings]');
    expect(values.SessionName).toBe('Real Boot Test');
    expect(values.ServerPVE).toBe('True');
  });
});
