import { describe, it, expect } from 'vitest';
import { mergeIni, readIniSection, iniSections } from './ini';

// A realistic fragment: ARK writes hundreds of keys the visual editor knows
// nothing about, plus sections it does not model at all.
const REAL_FILE = `[ServerSettings]
SessionName=Real Boot Test
ServerPassword=
MaxPlayers=10
TheMaxStructuresInRange=10500.000000
OxygenSwimSpeedStatMultiplier=1.000000
StructurePreventResourceRadiusMultiplier=1.000000
; a comment ARK left behind
ServerPVE=False

[SessionSettings]
SessionName=Real Boot Test

[/Script/Engine.GameSession]
MaxPlayers=10

[MessageOfTheDay]
Message=Welcome
`;

describe('mergeIni', () => {
  // The bug this exists to prevent: rebuilding the file from a curated list of
  // known keys silently deleted every key ARK had written.
  it('preserves keys it does not know about', () => {
    const out = mergeIni(REAL_FILE, { MaxPlayers: 42 }, '[ServerSettings]');

    for (const preserved of [
      'TheMaxStructuresInRange=10500.000000',
      'OxygenSwimSpeedStatMultiplier=1.000000',
      'StructurePreventResourceRadiusMultiplier=1.000000',
      '; a comment ARK left behind',
    ]) {
      expect(out).toContain(preserved);
    }
  });

  it('preserves sections it does not model', () => {
    const out = mergeIni(REAL_FILE, { MaxPlayers: 42 }, '[ServerSettings]');
    expect(out).toContain('[/Script/Engine.GameSession]');
    expect(out).toContain('[MessageOfTheDay]');
    expect(out).toContain('Message=Welcome');
  });

  it('updates an existing key in place', () => {
    const out = mergeIni(REAL_FILE, { MaxPlayers: 42 }, '[ServerSettings]');
    expect(out).toContain('MaxPlayers=42');
    // the same key in another section must not be touched
    expect(out).toMatch(/\[\/Script\/Engine\.GameSession\]\nMaxPlayers=10/);
  });

  it('appends a key that is missing from the section', () => {
    const out = mergeIni(REAL_FILE, { ServerHardcore: true }, '[ServerSettings]');
    expect(out).toContain('ServerHardcore=True');
    expect(out).toContain('SessionName=Real Boot Test');
  });

  it('writes booleans in the True/False form ARK expects', () => {
    const out = mergeIni(REAL_FILE, { ServerPVE: true, ServerHardcore: false }, '[ServerSettings]');
    expect(out).toContain('ServerPVE=True');
    expect(out).toContain('ServerHardcore=False');
    expect(out).not.toContain('ServerPVE=true');
  });

  // ARK is inconsistent about casing between what it writes and what the docs
  // use, so a case mismatch must not produce a duplicate key.
  it('matches keys case-insensitively without duplicating them', () => {
    const out = mergeIni(REAL_FILE, { serverpve: true }, '[ServerSettings]');
    const occurrences = out.split('\n').filter((l) => /^serverpve=/i.test(l));
    expect(occurrences).toHaveLength(1);
    expect(out).toContain('ServerPVE=True');
  });

  it('creates the section when it is absent', () => {
    const out = mergeIni('[Other]\nX=1\n', { ServerPVE: true }, '[ServerSettings]');
    expect(out).toContain('[Other]');
    expect(out).toContain('X=1');
    expect(out).toContain('[ServerSettings]');
    expect(out).toContain('ServerPVE=True');
  });

  it('is a no-op when there is nothing to apply', () => {
    expect(mergeIni(REAL_FILE, {}, '[ServerSettings]')).toBe(REAL_FILE);
    expect(mergeIni(REAL_FILE, { MaxPlayers: undefined }, '[ServerSettings]')).toBe(REAL_FILE);
  });

  it('does not corrupt values containing = or brackets', () => {
    const src = '[ServerSettings]\nOverrideNamedEngramEntries=(EngramClassName="X",Remove=true)\n';
    const out = mergeIni(src, { MaxPlayers: 5 }, '[ServerSettings]');
    expect(out).toContain('OverrideNamedEngramEntries=(EngramClassName="X",Remove=true)');
  });
});

describe('readIniSection', () => {
  it('reads only the requested section', () => {
    const values = readIniSection(REAL_FILE, '[ServerSettings]');
    expect(values.MaxPlayers).toBe('10');
    expect(values.ServerPVE).toBe('False');
    expect(values.Message).toBeUndefined();
  });

  it('returns the real values so the editor reflects the file', () => {
    const values = readIniSection(REAL_FILE, '[ServerSettings]');
    expect(values.TheMaxStructuresInRange).toBe('10500.000000');
  });
});

describe('iniSections', () => {
  it('lists every section in order', () => {
    expect(iniSections(REAL_FILE)).toEqual([
      '[ServerSettings]',
      '[SessionSettings]',
      '[/Script/Engine.GameSession]',
      '[MessageOfTheDay]',
    ]);
  });
});

describe('section casing', () => {
  // ARK writes [/Script/ShooterGame.ShooterGameMode]; other casings appear in
  // docs and code. A mismatch must update the existing section, not silently
  // append a duplicate one.
  it('matches a section regardless of case', () => {
    const src = '[/Script/ShooterGame.ShooterGameMode]\nXPMultiplier=1.0\nUnknownKey=keepme\n';
    const out = mergeIni(src, { XPMultiplier: 2.5 }, '[/script/shootergame.shootergamemode]');

    expect(out).toContain('XPMultiplier=2.5');
    expect(out).toContain('UnknownKey=keepme');
    expect(out.match(/\[\/Script\/ShooterGame\.ShooterGameMode\]/gi)).toHaveLength(1);
  });
});
