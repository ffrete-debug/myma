export interface ServerPreset {
  name: string;
  label: string;
  type: 'PVE' | 'PVP';
  multiplier: string;
  description: string;
  changes: string[];
  iniSections: Record<string, Record<string, string>>;
}

const presets: ServerPreset[] = [
  // ── PVE ──
  {
    name: 'pve-5x',
    label: 'PVE 5x',
    type: 'PVE',
    multiplier: '5x',
    description: 'Balanced PVE. Taming/Harvest/XP at 5x. Ideal for small tribes.',
    changes: ['TamingSpeedMultiplier=5.0', 'HarvestAmountMultiplier=5.0', 'XPMultiplier=5.0', 'PlayerCharacterFoodDrainMultiplier=0.5'],
    iniSections: {
      '[/Script/ShooterGame.ShooterGameMode]': {
        TamingSpeedMultiplier: '5.0',
        HarvestAmountMultiplier: '5.0',
        XPMultiplier: '5.0',
        PlayerCharacterFoodDrainMultiplier: '0.5',
        DinoCharacterFoodDrainMultiplier: '0.5',
      },
    },
  },
  {
    name: 'pve-10x',
    label: 'PVE 10x',
    type: 'PVE',
    multiplier: '10x',
    description: 'Fast PVE. 10x rates for quick progression. Popular for casual servers.',
    changes: ['TamingSpeedMultiplier=10.0', 'HarvestAmountMultiplier=10.0', 'XPMultiplier=10.0', 'PlayerCharacterFoodDrainMultiplier=0.3'],
    iniSections: {
      '[/Script/ShooterGame.ShooterGameMode]': {
        TamingSpeedMultiplier: '10.0',
        HarvestAmountMultiplier: '10.0',
        XPMultiplier: '10.0',
        PlayerCharacterFoodDrainMultiplier: '0.3',
        DinoCharacterFoodDrainMultiplier: '0.3',
        MatingSpeedMultiplier: '3.0',
        BabyMatureSpeedMultiplier: '3.0',
        EggHatchSpeedMultiplier: '3.0',
      },
    },
  },
  {
    name: 'pve-20x',
    label: 'PVE 20x',
    type: 'PVE',
    multiplier: '20x',
    description: 'Very fast PVE. 20x rates for large tribes. Fast breeding and maturation.',
    changes: ['TamingSpeedMultiplier=20.0', 'HarvestAmountMultiplier=20.0', 'XPMultiplier=20.0', 'MatingSpeedMultiplier=5.0', 'BabyMatureSpeedMultiplier=5.0'],
    iniSections: {
      '[/Script/ShooterGame.ShooterGameMode]': {
        TamingSpeedMultiplier: '20.0',
        HarvestAmountMultiplier: '20.0',
        XPMultiplier: '20.0',
        PlayerCharacterFoodDrainMultiplier: '0.2',
        DinoCharacterFoodDrainMultiplier: '0.2',
        MatingSpeedMultiplier: '5.0',
        BabyMatureSpeedMultiplier: '5.0',
        EggHatchSpeedMultiplier: '5.0',
        BabyCuddleIntervalMultiplier: '0.1',
      },
    },
  },
  {
    name: 'pve-1000x',
    label: 'PVE 1000x',
    type: 'PVE',
    multiplier: '1000x',
    description: 'Instant PVE. Everything maxed. For testing or creative building.',
    changes: ['TamingSpeedMultiplier=1000.0', 'HarvestAmountMultiplier=1000.0', 'XPMultiplier=1000.0', 'BabyMatureSpeedMultiplier=100.0'],
    iniSections: {
      '[/Script/ShooterGame.ShooterGameMode]': {
        TamingSpeedMultiplier: '1000.0',
        HarvestAmountMultiplier: '1000.0',
        XPMultiplier: '1000.0',
        PlayerCharacterFoodDrainMultiplier: '0.01',
        DinoCharacterFoodDrainMultiplier: '0.01',
        MatingSpeedMultiplier: '50.0',
        BabyMatureSpeedMultiplier: '100.0',
        EggHatchSpeedMultiplier: '100.0',
        BabyCuddleIntervalMultiplier: '0.01',
      },
    },
  },
  // ── PVP ──
  {
    name: 'pvp-5x',
    label: 'PVP 5x',
    type: 'PVP',
    multiplier: '5x',
    description: 'Balanced PVP. 5x rates with PVP settings enabled. Tribe wars enabled.',
    changes: ['TamingSpeedMultiplier=5.0', 'HarvestAmountMultiplier=5.0', 'XPMultiplier=5.0', 'bEnablePvPGamma=True', 'bPvEDisableFriendlyFire=False'],
    iniSections: {
      '[/Script/ShooterGame.ShooterGameMode]': {
        TamingSpeedMultiplier: '5.0',
        HarvestAmountMultiplier: '5.0',
        XPMultiplier: '5.0',
      },
      '[ServerSettings]': {
        bEnablePvPGamma: 'True',
        bPvEDisableFriendlyFire: 'False',
        bDisableFriendlyFire: 'False',
        bAllowFlyerCarryPvE: 'False',
      },
    },
  },
  {
    name: 'pvp-10x',
    label: 'PVP 10x',
    type: 'PVP',
    multiplier: '10x',
    description: 'Fast PVP. 10x rates with boosted structure resistance.',
    changes: ['TamingSpeedMultiplier=10.0', 'HarvestAmountMultiplier=10.0', 'XPMultiplier=10.0', 'StructureResistanceMultiplier=3.0'],
    iniSections: {
      '[/Script/ShooterGame.ShooterGameMode]': {
        TamingSpeedMultiplier: '10.0',
        HarvestAmountMultiplier: '10.0',
        XPMultiplier: '10.0',
        StructureResistanceMultiplier: '3.0',
        StructureDamageMultiplier: '1.5',
      },
      '[ServerSettings]': {
        bEnablePvPGamma: 'True',
        bPvEDisableFriendlyFire: 'False',
        bDisableFriendlyFire: 'False',
      },
    },
  },
  {
    name: 'pvp-20x',
    label: 'PVP 20x',
    type: 'PVP',
    multiplier: '20x',
    description: 'Very fast PVP. 20x rates with reduced structure resistance for raiding.',
    changes: ['TamingSpeedMultiplier=20.0', 'HarvestAmountMultiplier=20.0', 'XPMultiplier=20.0', 'HarvestHealthMultiplier=2.0'],
    iniSections: {
      '[/Script/ShooterGame.ShooterGameMode]': {
        TamingSpeedMultiplier: '20.0',
        HarvestAmountMultiplier: '20.0',
        XPMultiplier: '20.0',
        HarvestHealthMultiplier: '2.0',
        StructureResistanceMultiplier: '2.0',
        StructureDamageMultiplier: '2.0',
        MatingSpeedMultiplier: '3.0',
        BabyMatureSpeedMultiplier: '3.0',
      },
      '[ServerSettings]': {
        bEnablePvPGamma: 'True',
        bPvEDisableFriendlyFire: 'False',
        bDisableFriendlyFire: 'False',
      },
    },
  },
  {
    name: 'pvp-1000x',
    label: 'PVP 1000x',
    type: 'PVP',
    multiplier: '1000x',
    description: 'Instant PVP. Everything maxed. Pure chaos.',
    changes: ['TamingSpeedMultiplier=1000.0', 'HarvestAmountMultiplier=1000.0', 'XPMultiplier=1000.0', 'StructureDamageMultiplier=10.0'],
    iniSections: {
      '[/Script/ShooterGame.ShooterGameMode]': {
        TamingSpeedMultiplier: '1000.0',
        HarvestAmountMultiplier: '1000.0',
        XPMultiplier: '1000.0',
        StructureDamageMultiplier: '10.0',
        MatingSpeedMultiplier: '50.0',
        BabyMatureSpeedMultiplier: '100.0',
        EggHatchSpeedMultiplier: '100.0',
      },
      '[ServerSettings]': {
        bEnablePvPGamma: 'True',
        bPvEDisableFriendlyFire: 'False',
        bDisableFriendlyFire: 'False',
      },
    },
  },
];

export function getPresetsByType(type: 'PVE' | 'PVP'): ServerPreset[] {
  return presets.filter(p => p.type === type);
}

export function getAllPresets(): ServerPreset[] {
  return presets;
}

export function getPresetByName(name: string): ServerPreset | undefined {
  return presets.find(p => p.name === name);
}

export function presetToIni(preset: ServerPreset): string {
  return Object.entries(preset.iniSections)
    .map(([section, kv]) => `${section}\n${Object.entries(kv).map(([k, v]) => `${k}=${v}`).join('\n')}`)
    .join('\n\n');
}
