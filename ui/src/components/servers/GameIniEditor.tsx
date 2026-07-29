"use client";

import { useState, useEffect, useCallback, useRef } from 'react';
import { useTranslations } from 'next-intl';
import { mergeIni } from '@/lib/ini';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

import { Info, Eye, EyeOff } from 'lucide-react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

// Game.ini configuration parameters
interface GameIniParam {
  type: 'boolean' | 'number' | 'text';
  default: boolean | number | string;
  min?: number;
  max?: number;
  step?: number;
  description?: string;
}

type GameIniCategoryKey =
  | 'gameBasic'
  | 'experienceSettings'
  | 'breedingSettings'
  | 'itemSettings'
  | 'dinoSettings'
  | 'tribeSettings'
  | 'pvpSettings'
  | 'structureSettings'
  | 'advancedSettings'
  | 'customSettings';

const gameIniParams: Record<GameIniCategoryKey, Record<string, GameIniParam>> = {
  gameBasic: {
    bUseSingleplayerSettings: { type: 'boolean', default: false },
    bDisableStructurePlacementCollision: { type: 'boolean', default: false, description: 'Permite que estruturas sejam colocadas sobrepostas, desativando a verificação de colisão de posicionamento.' },
    bAllowFlyerCarryPvE: { type: 'boolean', default: true, description: 'Permite que criaturas voadoras carreguem dinos selvagens no modo PvE.' },
    bDisableStructureDecayPvE: { type: 'boolean', default: false, description: 'Desativa os temporizadores automáticos de deterioração de estruturas no PvE.' },
    bAllowUnlimitedRespecs: { type: 'boolean', default: true, description: 'Permite redefinições ilimitadas de atributos e engramas usando o Mindwipe.' },
    bAllowPlatformSaddleMultiFloors: { type: 'boolean', default: true, description: 'Permite construir múltiplos andares em selas plataforma como Quetzal e Bronto.' },
    bPassiveDefensesDamageRiderlessDinos: { type: 'boolean', default: true, description: 'Faz com que defesas passivas causem dano a dinos sem montaria.' },
    bPvEDisableFriendlyFire: { type: 'boolean', default: false, description: 'Desativa completamente o dano amigável no modo PvE.' },
    bDisableFriendlyFire: { type: 'boolean', default: false, description: 'Desativa o dano amigável em todos os modos de jogo.' },
    bEnablePvPGamma: { type: 'boolean', default: false, description: 'Permite ajustar o brilho e gama durante o modo PvP.' },
    DifficultyOffset: { type: 'number', default: 1.0, min: 0.0, max: 1.0, step: 0.1, description: 'Controla a distribuição de níveis dos dinos selvagens. 1.0 = nível máximo 150.' },
    OverrideOfficialDifficulty: { type: 'number', default: 5.0, min: 1.0, max: 10.0, step: 0.1, description: 'Substitui a dificuldade oficial, alterando o nível máximo dos dinos selvagens.' }
  },
  experienceSettings: {
    XPMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador de toda experiência ganha. 2.0 dobra o XP recebido.' },
    PlayerCharacterWaterDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da taxa de consumo de água do jogador. Menor = drena mais devagar.' },
    PlayerCharacterFoodDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da taxa de consumo de comida do jogador. Menor = drena mais devagar.' },
    PlayerCharacterStaminaDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da taxa de uso de estamina do jogador. Menor = cansa mais devagar.' },
    PlayerCharacterHealthRecoveryMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da velocidade de regeneração passiva de vida do jogador.' }
  },
  breedingSettings: {
    MatingIntervalMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador do intervalo de acasalamento entre reproduções. Menor = intervalos mais curtos.' },
    EggHatchSpeedMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 50.0, step: 0.1, description: 'Multiplicador da velocidade de incubação e eclosão de ovos. Maior = eclosão mais rápida.' },
    BabyMatureSpeedMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 50.0, step: 0.1, description: 'Multiplicador da velocidade de maturação de filhotes. Maior = crescimento mais rápido.' },
    BabyFoodConsumptionSpeedMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da taxa de consumo de comida de filhotes. Menor = precisam de menos comida.' },
    BabyCuddleIntervalMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador do intervalo de carinho para imprinting. Menor = imprinting mais frequente.' },
    BabyCuddleGracePeriodMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador do período de carência antes de perder qualidade do imprint dos filhotes.' },
    BabyCuddleLoseImprintQualitySpeedMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da velocidade de perda de qualidade de imprint quando filhotes não recebem carinho.' }
  },
  itemSettings: {
    HarvestAmountMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da quantidade de recursos colhidos. Maior = mais recursos por colheita.' },
    HarvestHealthMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da vida dos recursos minerais e plantas ao serem colhidos.' },
    ResourcesRespawnPeriodMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador do período de reaparecimento de recursos naturais como pedras e madeira.' },
    ItemStackSizeMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador do tamanho máximo da pilha de cada tipo de item no inventário.' },
    CropGrowthSpeedMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da velocidade de crescimento de plantações cultivadas.' },
    GlobalItemDecompositionTimeMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador do tempo de decomposição de itens largados no chão.' },
    GlobalCorpseDecompositionTimeMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador do tempo de decomposição de cadáveres de jogadores.' }
  },
  dinoSettings: {
    TamingSpeedMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 50.0, step: 0.1, description: 'Multiplicador da velocidade de domação. Maior = domação mais rápida.' },
    DinoCharacterFoodDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da taxa de consumo de comida de todos os dinos. Menor = comem menos.' },
    DinoCharacterStaminaDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da taxa de uso de estamina de todos os dinos.' },
    DinoCharacterHealthRecoveryMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da regeneração passiva de vida dos dinos.' },
    DinoCountMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da quantidade de dinos selvagens que spawnam. Maior = mais dinos no mapa.' },
    WildDinoCharacterFoodDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da taxa de consumo de comida de dinos selvagens.' },
    WildDinoTorporDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da taxa de perda de torpor de dinos selvagens. Menor = ficam inconscientes por mais tempo.' }
  },
  tribeSettings: {
    MaxNumberOfPlayersInTribe: { type: 'number', default: 0, min: 0, max: 100, step: 1, description: 'Número máximo de jogadores permitidos em uma única tribo. 0 = ilimitado.' },
    TribeNameChangeCooldown: { type: 'number', default: 15, min: 0, max: 1440, step: 1, description: 'Tempo de espera em dias entre alterações de nome de tribo.' },
    bPvEAllowTribeWar: { type: 'boolean', default: false, description: 'Permite que tribos declarem guerra a outras tribos no modo PvE.' },
    bPvEAllowTribeWarCancel: { type: 'boolean', default: false, description: 'Permite que tribos cancelem uma declaração de guerra ativa.' }
  },
  pvpSettings: {
    bIncreasePvPRespawnInterval: { type: 'boolean', default: false, description: 'Aumenta o intervalo de reaparecimento de jogadores no PvP após serem mortos.' },
    IncreasePvPRespawnIntervalCheckPeriod: { type: 'number', default: 300, min: 60, max: 3600, step: 60, description: 'Período em segundos para verificar a elegibilidade do intervalo de reaparecimento PvP.' },
    IncreasePvPRespawnIntervalMultiplier: { type: 'number', default: 1.0, min: 1.0, max: 5.0, step: 0.1, description: 'Multiplicador aplicado ao temporizador de reaparecimento após mortes em PvP.' },
    IncreasePvPRespawnIntervalBaseAmount: { type: 'number', default: 60, min: 30, max: 600, step: 30, description: 'Valor base em segundos do atraso de reaparecimento antes da aplicação do multiplicador PvP.' }
  },
  structureSettings: {
    StructureDamageMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador do dano causado a estruturas. Maior = estruturas recebem mais dano.' },
    StructureResistanceMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da resistência de estruturas a dano. Maior = estruturas recebem menos dano.' },
    StructureDamageRepairCooldown: { type: 'number', default: 180, min: 0, max: 3600, step: 60, description: 'Tempo de espera em segundos antes que uma estrutura danificada possa ser reparada.' },
    PvEStructureDecayPeriodMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Multiplicador da duração do temporizador de deterioração de estruturas no PvE.' },
    MaxStructuresInRange: { type: 'number', default: 1300, min: 100, max: 10000, step: 100, description: 'Número máximo de estruturas permitidas em um raio de proximidade para otimizar o desempenho.' }
  },
  advancedSettings: {
    bAutoPvETimer: { type: 'boolean', default: false, description: 'Ativa a alternância automática entre modos PvP e PvE baseada em temporizador.' },
    bAutoPvEUseSystemTime: { type: 'boolean', default: false, description: 'Usa o horário do sistema para a alternância automática do modo PvE.' },
    AutoPvEStartTimeSeconds: { type: 'number', default: 0, min: 0, max: 86400, step: 3600, description: 'Horário em segundos (0 a 86400) quando a alternância para PvE é ativada.' },
    AutoPvEStopTimeSeconds: { type: 'number', default: 0, min: 0, max: 86400, step: 3600, description: 'Horário em segundos (0 a 86400) quando a alternância para PvE é desativada.' },
    bOnlyAllowSpecifiedEngrams: { type: 'boolean', default: false, description: 'Permite apenas os engramas explicitamente habilitados na lista de engramas permitidos.' },
    bAutoUnlockAllEngrams: { type: 'boolean', default: false, description: 'Desbloqueia automaticamente todos os engramas para todos os jogadores ao entrar no servidor.' },
    bShowCreativeMode: { type: 'boolean', default: false, description: 'Exibe a categoria de engramas do modo criativo no menu de criação.' },
    bUseCorpseLocator: { type: 'boolean', default: false, description: 'Ativa um marcador na bússola apontando para a localização da última morte do jogador.' },
    bDisableLootCrates: { type: 'boolean', default: false, description: 'Desativa o surgimento de baús de saque no mapa do servidor.' },
    bDisableDinoRiding: { type: 'boolean', default: false, description: 'Impede que jogadores montem em dinos no servidor.' },
    bDisableDinoTaming: { type: 'boolean', default: false, description: 'Desativa completamente a domação de dinos no servidor.' },
    bAllowCustomRecipes: { type: 'boolean', default: true, description: 'Permite que jogadores criem receitas de criação personalizadas.' }
  },
  customSettings: {
    DayCycleSpeedScale: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Velocidade do ciclo completo de dia e noite. 2.0 = ciclo duas vezes mais rápido.' },
    NightTimeSpeedScale: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Velocidade da fase noturna especificamente. 2.0 = noites mais curtas.' },
    DayTimeSpeedScale: { type: 'number', default: 1.0, min: 0.1, max: 10.0, step: 0.1, description: 'Velocidade da fase diurna especificamente. 2.0 = dias mais curtos.' }
  }
};

function getGameIniParamsByCategory(category: GameIniCategoryKey): Record<string, GameIniParam> {
  return gameIniParams[category] || {};
}



interface GameIniEditorProps {
  value?: string;
  onChange?: (value: string) => void;
}

export function GameIniEditor({ value, onChange }: GameIniEditorProps) {
  const t = useTranslations('servers.editor');
  const tCategories = useTranslations('servers.gameIniCategories');
  const tParams = useTranslations('servers.gameIniParams');
  const tParamDescriptions = useTranslations('servers.gameIniParamDescriptions');
  const [editMode, setEditMode] = useState<'visual' | 'text'>('visual');
  const [textContent, setTextContent] = useState('');
  const [visualConfig, setVisualConfig] = useState<Record<string, string | number | boolean>>({});
  const [activeTab, setActiveTab] = useState<GameIniCategoryKey>('gameBasic');
  const [isUserEditing, setIsUserEditing] = useState(false);
  // Raw text being typed into a number field, so partial input ("", "1.") survives.
  const [numberDrafts, setNumberDrafts] = useState<Record<string, string>>({});

  // Keep the latest onChange in a ref so the callbacks below stay stable even
  // when the parent passes a new inline arrow function on every render.
  const onChangeRef = useRef(onChange);
  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  // Content we emitted ourselves, so the echoed `value` does not get reparsed.
  const lastEmittedRef = useRef<string | null>(null);
  // Only propagate upwards after the user actually touched the editor, so that
  // merely opening a server does not rewrite its stored Game.ini.
  const hasUserEditedRef = useRef(false);
  const editingTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const parseTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => () => {
    if (editingTimerRef.current) clearTimeout(editingTimerRef.current);
    if (parseTimerRef.current) clearTimeout(parseTimerRef.current);
  }, []);

  // Get parameter display name from translations
  const getParamDisplayName = (paramKey: string): string =>
    tParams.has(paramKey) ? tParams(paramKey) : paramKey;

  const getParamDescription = (paramKey: string, param: GameIniParam): string | undefined =>
    tParamDescriptions.has(paramKey) ? tParamDescriptions(paramKey) : param.description;

  const parseTextToVisual = useCallback((text: string) => {
    try {
      const config: Record<string, string | number | boolean> = {};
      const lines = text.split('\n');

      for (const line of lines) {
        const trimmedLine = line.trim();
        if (trimmedLine && !trimmedLine.startsWith(';') && !trimmedLine.startsWith('#') && !trimmedLine.startsWith('[')) {
          const [key, ...valueParts] = trimmedLine.split('=');
          if (key && valueParts.length > 0) {
            const value = valueParts.join('=').trim();

            // Find parameter definition
            let paramDef: GameIniParam | undefined;
            for (const categoryKey of Object.keys(gameIniParams)) {
              const category = categoryKey as GameIniCategoryKey;
              const params = getGameIniParamsByCategory(category);
              if (params[key.trim()]) {
                paramDef = params[key.trim()];
                break;
              }
            }

            if (paramDef) {
              if (paramDef.type === 'boolean') {
                config[key.trim()] = value.toLowerCase() === 'true';
              } else if (paramDef.type === 'number') {
                const numValue = parseFloat(value);
                config[key.trim()] = isNaN(numValue) ? paramDef.default : numValue;
              } else {
                config[key.trim()] = value;
              }
            }
          }
        }
      }

      setVisualConfig(config);
    } catch (error) {
      console.error(t('parseGameIniError') + ':', error);
    }
  }, [t]);

  // Initialize with default values, or hydrate from the incoming value
  useEffect(() => {
    if (!value) {
      const defaultConfig: Record<string, string | number | boolean> = {};
      Object.keys(gameIniParams).forEach(categoryKey => {
        const category = categoryKey as GameIniCategoryKey;
        const params = getGameIniParamsByCategory(category);
        Object.keys(params).forEach(paramKey => {
          defaultConfig[paramKey] = params[paramKey].default;
        });
      });
      setVisualConfig(defaultConfig);
      return;
    }

    // Ignore the echo of the content this editor just emitted.
    if (value === lastEmittedRef.current) return;

    setTextContent(value);
    parseTextToVisual(value);
  }, [value, parseTextToVisual]);

  // Keep the newest text in a ref so the merge applies onto the CURRENT file,
  // not a value captured when the callback was created.
  const textContentRef = useRef(textContent);
  useEffect(() => { textContentRef.current = textContent; }, [textContent]);

  const syncVisualToText = useCallback(() => {
    try {
      // MERGE onto the real Game.ini instead of rebuilding it.
      //
      // Rebuilding from the editor's own parameter catalogue silently deleted
      // every key ARK had written that the catalogue did not know about,
      // including whole sections such as [ModInstaller]. Unknown content is now
      // preserved untouched.
      const base = textContentRef.current || '';
      const merged = mergeIni(base, visualConfig, '[/Script/ShooterGame.ShooterGameMode]');

      setTextContent(merged);
      onChange?.(merged);
    } catch (error) {
      console.error('Failed to apply visual settings to Game.ini:', error);
    }
  }, [visualConfig, onChange]);

  // Sync visual to text when visual config changes
  useEffect(() => {
    if (editMode === 'visual' && !isUserEditing) {
      syncVisualToText();
    }
  }, [editMode, isUserEditing, syncVisualToText]);

  const handleVisualChange = (key: string, value: string | number | boolean) => {
    hasUserEditedRef.current = true;
    setIsUserEditing(true);
    setVisualConfig(prev => ({ ...prev, [key]: value }));

    // Debounce the visual -> text sync until the user stops typing.
    if (editingTimerRef.current) clearTimeout(editingTimerRef.current);
    editingTimerRef.current = setTimeout(() => {
      editingTimerRef.current = null;
      setIsUserEditing(false);
    }, 1000);
  };

  const handleTextChange = (newText: string) => {
    hasUserEditedRef.current = true;
    setTextContent(newText);
    lastEmittedRef.current = newText;
    onChangeRef.current?.(newText);

    // Parse text to visual with debounce
    if (parseTimerRef.current) clearTimeout(parseTimerRef.current);
    parseTimerRef.current = setTimeout(() => {
      parseTimerRef.current = null;
      parseTextToVisual(newText);
    }, 500);
  };

  const handleModeSwitch = (mode: 'visual' | 'text') => {
    if (mode === 'text' && editMode === 'visual') {
      syncVisualToText();
    } else if (mode === 'visual' && editMode === 'text') {
      parseTextToVisual(textContent);
    }
    setEditMode(mode);
  };



  const getCategoryDisplayName = (categoryKey: GameIniCategoryKey): string =>
    tCategories.has(categoryKey) ? tCategories(categoryKey) : categoryKey;

  const renderParamControl = (paramKey: string, param: GameIniParam) => {
    const currentValue = visualConfig[paramKey] ?? param.default;

    switch (param.type) {
      case 'boolean':
        return (
          <div className="flex items-center space-x-2">
            <Switch
              checked={Boolean(currentValue)}
              onCheckedChange={(checked) => handleVisualChange(paramKey, checked)}
            />
            <span className="text-sm text-muted-foreground">
              {Boolean(currentValue) ? t('enabled') : t('disabled')}
            </span>
          </div>
        );

      case 'number': {
        const commitNumberDraft = (raw: string) => {
          const parsed = Number.parseFloat(raw);
          if (raw.trim() === '' || !Number.isFinite(parsed)) return;
          handleVisualChange(paramKey, parsed);
        };

        return (
          <div className="space-y-1">
            <Input
              type="number"
              value={numberDrafts[paramKey] ?? String(currentValue)}
              onChange={(e) => {
                const raw = e.target.value;
                setNumberDrafts(prev => ({ ...prev, [paramKey]: raw }));
                commitNumberDraft(raw);
              }}
              onBlur={(e) => {
                const raw = e.target.value;
                // Drop the draft so the field falls back to the committed value.
                setNumberDrafts(prev => {
                  const next = { ...prev };
                  delete next[paramKey];
                  return next;
                });
                commitNumberDraft(raw);
              }}
              min={param.min}
              max={param.max}
              step={param.step}
              className="w-full"
            />
            {(param.min !== undefined || param.max !== undefined) && (
              <div className="text-xs text-muted-foreground">
                {t('range')}: {param.min ?? '∞'} - {param.max ?? '∞'}
              </div>
            )}
          </div>
        );
      }

      case 'text':
        return (
          <Input
            type="text"
            value={String(currentValue)}
            onChange={(e) => handleVisualChange(paramKey, e.target.value)}
            className="w-full"
          />
        );

      default:
        return null;
    }
  };

  return (
    <div className="space-y-4">
      {/* Mode Switch */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <Button
            variant={editMode === 'visual' ? 'default' : 'outline'}
            size="sm"
            onClick={() => handleModeSwitch('visual')}
          >
            <Eye className="w-4 h-4 mr-1" />
            {t('visualEditMode')}
          </Button>
          <Button
            variant={editMode === 'text' ? 'default' : 'outline'}
            size="sm"
            onClick={() => handleModeSwitch('text')}
          >
            <EyeOff className="w-4 h-4 mr-1" />
            {t('textEditMode')}
          </Button>
        </div>


      </div>

      {editMode === 'visual' ? (
        <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as GameIniCategoryKey)}>
          <TabsList className="grid w-full grid-cols-3 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-9 gap-2 h-auto p-2 bg-muted/50 border rounded-lg">
            {Object.keys(gameIniParams).map((categoryKey) => (
              <TabsTrigger 
                key={categoryKey} 
                value={categoryKey} 
                className="text-xs px-3 py-2 flex-shrink-0 border-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-primary data-[state=active]:text-primary-foreground data-[state=active]:shadow-md hover:border-muted-foreground/50 hover:bg-muted transition-all duration-200 font-medium rounded-md"
              >
                {getCategoryDisplayName(categoryKey as GameIniCategoryKey)}
              </TabsTrigger>
            ))}
          </TabsList>

          {Object.keys(gameIniParams).map((categoryKey) => {
            const category = categoryKey as GameIniCategoryKey;
            const params = getGameIniParamsByCategory(category);

            return (
              <TabsContent key={categoryKey} value={categoryKey} className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      {getCategoryDisplayName(category)}
                    </CardTitle>
                    <CardDescription>
                      {Object.keys(params).length}{t('parametersCount')}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid gap-4">
                      {Object.keys(params).map((paramKey) => {
                        const param = params[paramKey];
                        const paramDescription = getParamDescription(paramKey, param);
                        return (
                          <div key={paramKey} className="space-y-2">
                            <Label className="text-sm font-medium">
                              {getParamDisplayName(paramKey)}
                              <TooltipProvider>
                                <Tooltip>
                                  <TooltipTrigger asChild>
                                    <Info className="w-4 h-4 ml-1 inline-block cursor-help" />
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    <div className="max-w-xs">
                                      <p className="font-medium">{paramKey}</p>
                                      <p className="text-sm mt-1">
                                        {paramDescription ?? getParamDisplayName(paramKey)}
                                      </p>
                                      <p className="text-xs mt-1 text-muted-foreground">
                                        {t('defaultValue')}: {String(param.default)}
                                      </p>
                                    </div>
                                  </TooltipContent>
                                </Tooltip>
                              </TooltipProvider>
                            </Label>
                            {renderParamControl(paramKey, param)}
                          </div>
                        );
                      })}
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>
            );
          })}
        </Tabs>
      ) : (
        <div className="space-y-2">
          <Label htmlFor="game-ini-text">{t('textEditMode')}</Label>
          <Textarea
            id="game-ini-text"
            value={textContent}
            onChange={(e) => handleTextChange(e.target.value)}
            className="min-h-[400px] font-mono text-sm"
            placeholder={t('placeholders.gameIni')}
          />
        </div>
      )}
    </div>
  );
}