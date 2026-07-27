"use client";

import { useState, useEffect, useCallback, useRef } from 'react';
import { useTranslations } from 'next-intl';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Info, Eye, EyeOff } from 'lucide-react';
import { Switch } from '@/components/ui/switch';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';


// GameUserSettings configuration parameters
interface GameUserSettingsParam {
  type: 'boolean' | 'number' | 'text' | 'password';
  default: boolean | number | string;
  min?: number;
  max?: number;
  step?: number;
  description?: string;
}

type GameUserSettingsCategoryKey =
  | 'serverBasic'
  | 'gameMode'
  | 'communication'
  | 'gameMultipliers'
  | 'characterSettings'
  | 'dinoSettings'
  | 'environmentSettings'
  | 'structureSettings'
  | 'tribeSettings'
  | 'breedingSettings'
  | 'itemSettings'
  | 'performanceSettings'
  | 'diseaseSettings'
  | 'offlineRaidSettings'
  | 'crossArkSettings'
  | 'flyerSettings'
  | 'advancedSettings';

const gameUserSettingsParams: Record<GameUserSettingsCategoryKey, Record<string, GameUserSettingsParam>> = {
  serverBasic: {
    ServerPassword: { type: 'password', default: '' , description: 'Password required for players to join the server. Leave empty for public.',
      },
    SpectatorPassword: { type: 'password', default: '' , description: 'Senha exigida para entrar no servidor como espectador (modo spectator).',
      },
    AdminLogging: { type: 'boolean', default: true, description: 'Loga no console todas as ações executadas por admins no servidor.' }
  },
  gameMode: {
    serverPVE: { type: 'boolean', default: false , description: 'Habilita modo PvE. Desativa PvP entre jogadores.',
      },
    serverHardcore: { type: 'boolean', default: false , description: 'Habilita modo hardcore — morte permanente deleta todo o progresso.',
      },
    ShowMapPlayerLocation: { type: 'boolean', default: false , description: 'Exibe a localização de todos os jogadores no mapa do jogo.',
      },
    allowThirdPersonPlayer: { type: 'boolean', default: false , description: 'Permite que jogadores usem visão em terceira pessoa.',
      },
    ServerCrosshair: { type: 'boolean', default: true , description: 'Exibe mira reticulada na tela para todos os jogadores.',
      },
    EnablePvPGamma: { type: 'boolean', default: false , description: 'Habilita ajuste de gamma (brilho noturno) em servidores PvP.',
      },
    DisablePvEGamma: { type: 'boolean', default: false , description: 'Desabilita ajuste de brilho gamma em servidores PvE.',
      },
    serverForceNoHud: { type: 'boolean', default: false , description: 'Força remoção total do HUD da tela para todos os jogadores.',
      },
    ShowFloatingDamageText: { type: 'boolean', default: false , description: 'Mostra números flutuantes de dano na tela ao atacar.',
      },
    AllowHitMarkers: { type: 'boolean', default: true, description: 'Exibe marcadores de acerto ao atacar.' }
  },
  communication: {
    globalVoiceChat: { type: 'boolean', default: false , description: 'Habilita chat de voz global que alcança todos no servidor.',
      },
    proximityChat: { type: 'boolean', default: false , description: 'Habilita chat de voz por proximidade (distância física).',
      },
    alwaysNotifyPlayerJoined: { type: 'boolean', default: false , description: 'Notifica todos os jogadores quando alguém conecta no servidor.',
      },
    alwaysNotifyPlayerLeft: { type: 'boolean', default: false , description: 'Notifica todos quando um jogador desconecta do servidor.',
      },
    DontAlwaysNotifyPlayerJoined: { type: 'boolean', default: false, description: 'Desabilita notificação automática de entrada de jogadores.' }
  },
  gameMultipliers: {
    XPMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 100, step: 0.1 , description: 'Multiplicador de XP ganho. 2.0 = dobro de XP.',
      },
    TamingSpeedMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 100, step: 0.1 , description: 'How fast dinos tame. 2.0 = twice as fast taming speed.',
      },
    HarvestAmountMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 100, step: 0.1 , description: 'Multiplier for resource harvest yield from nodes and dinos.',
      },
    HarvestHealthMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da vida dos recursos naturais ao serem coletados.'},
    ResourcesRespawnPeriodMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador do tempo de respawn de recursos no mapa.',
      },
    ItemStackSizeMultiplier: { type: 'number', default: 1.0, min: 1, max: 100, step: 1 , description: 'Multiplicador da quantidade maxima de itens por pilha no inventario.'}
  },
  characterSettings: {
    PlayerCharacterHealthRecoveryMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da recuperação de saúde do jogador.',
      },
    PlayerCharacterFoodDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da velocidade de perda de fome do jogador.'},
    PlayerCharacterWaterDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador do consumo de água do jogador.',
      },
    PlayerCharacterStaminaDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da velocidade de perda de stamina do jogador.'},
    PlayerDamageMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador de dano causado pelos jogadores.',
      },
    PlayerResistanceMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplier for damage resistance of players.',
      },
    OxygenSwimSpeedStatMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da velocidade de natação e estoque de oxigenio subaquatico.'},
    ImplantSuicideCD: { type: 'number', default: 28800, min: 0, max: 86400 , description: 'Tempo de recarga em segundos para o implante suicida apos uso.'}
  },
  dinoSettings: {
    DinoCountMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da quantidade de dinossauros selvagens no mapa.',
      },
    DinoCharacterHealthRecoveryMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da velocidade de recuperacao de vida das criaturas.'},
    DinoCharacterFoodDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador do consumo de comida dos dinossauros.',
      },
    DinoCharacterStaminaDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da velocidade de perda de stamina dos dinossauros.'},
    DinoDamageMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador de dano de ataque dos dinossauros.',
      },
    TamedDinoDamageMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador do dano de ataque causado por dinossauros domesticados.'},
    DinoResistanceMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador de resistência dos dinossauros contra dano.',
      },
    TamedDinoResistanceMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da resistencia dos dinossauros domesticados contra dano.'},
    MaxTamedDinos: { type: 'number', default: 4000, min: 0, max: 50000 , description: 'Numero maximo de dinossauros domesticados permitidos no servidor.'},
    MaxPersonalTamedDinos: { type: 'number', default: 0, min: 0, max: 5000 , description: 'Numero maximo de dinossauros domesticados por jogador individual.'},
    DisableDinoDecayPvE: { type: 'boolean', default: false , description: 'Desabilita decay de dinossauros selvagens no PvE.',
      },
    AutoDestroyDecayedDinos: { type: 'boolean', default: false , description: 'Destroi automaticamente criaturas selvagens que ja deterioraram completamente.'},
    PvEDinoDecayPeriodMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador do periodo de deterioracao de criaturas selvagens em PvE.'},
    PvPDinoDecay: { type: 'boolean', default: false , description: 'Ativa a deterioracao de criaturas selvagens em servidores PvP.'},
    AllowRaidDinoFeeding: { type: 'boolean', default: false , description: 'Permite alimentar dinossauros de ataque (raid) pelos jogadores.',
      },
    RaidDinoCharacterFoodDrainMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador do consumo de comida de criaturas de ataque durante raids.'},
    AllowFlyerCarryPvE: { type: 'boolean', default: false , description: 'Permite criaturas voadoras carregarem outras em PvE.',
      },
    bForceCanRideFliers: { type: 'boolean', default: false , description: 'Forca a permissao de montar criaturas voadoras independentemente de outras configuracoes.'}
  },
  environmentSettings: {
    DayCycleSpeedScale: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Velocidade do ciclo dia/noite. 2.0 = duas vezes mais rápido.',
      },
    DayTimeSpeedScale: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da velocidade da fase diurna do ciclo dia/noite.'},
    NightTimeSpeedScale: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador de velocidade da fase noturna.',
      },
    DisableWeatherFog: { type: 'boolean', default: false , description: 'Desabilita a neblina e efeitos de clima visual no mapa.'},
    DifficultyOffset: { type: 'number', default: 0.2, min: 0, max: 1, step: 0.1 , description: 'Difficulty offset for wild dino level scaling. Must match Game.ini value.',
      },
    OverrideOfficialDifficulty: { type: 'number', default: 5.0, min: 1, max: 10, step: 0.1 , description: 'Sobrescreve o nível máximo de criatura selvagem.',
      },
    RandomSupplyCratePoints: { type: 'boolean', default: false , description: 'Ativa pontos de spawn aleatorios para baus de suprimentos no mapa.'}
  },
  structureSettings: {
    StructureDamageMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador de dano causado às estruturas.',
      },
    StructureResistanceMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplier for damage resistance of structures.',
      },
    TheMaxStructuresInRange: { type: 'number', default: 10500, min: 1000, max: 50000 , description: 'Quantidade maxima de estruturas permitidas dentro do alcance de outro jogador.'},
    NewMaxStructuresInRange: { type: 'number', default: 6000, min: 1000, max: 50000 , description: 'Novo limite maximo de estruturas no alcance de construcao de um jogador.'},
    MaxStructuresInRange: { type: 'number', default: 1300, min: 100, max: 10000 , description: 'Maximum number of structures allowed within a certain proximity range. Helps performance.',
      },
    DisableStructureDecayPvE: { type: 'boolean', default: false , description: 'Desabilita a deterioracao de estruturas em servidores PvE.'},
    PvEStructureDecayPeriodMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador do período de decay de estruturas em PvE.',
      },
    PvPStructureDecay: { type: 'boolean', default: false , description: 'Ativa a deterioracao de estruturas em servidores PvP.'},
    StructurePickupTimeAfterPlacement: { type: 'number', default: 30, min: 0, max: 3600 , description: 'Tempo em segundos apos a colocacao em que uma estrutura pode ser recolhida sem penalidade.'},
    StructurePickupHoldDuration: { type: 'number', default: 0.5, min: 0, max: 10, step: 0.1 , description: 'Duracao em segundos que o jogador segura a estrutura ao recolhe-la.'},
    AlwaysAllowStructurePickup: { type: 'boolean', default: false , description: 'Sempre permite recolher estruturas sem condições.',
      },
    OnlyAutoDestroyCoreStructures: { type: 'boolean', default: false , description: 'Limita a destruicao automatica apenas as estruturas principais do nucleo.'},
    OnlyDecayUnsnappedCoreStructures: { type: 'boolean', default: false , description: 'Limita a deterioracao automatica apenas a estruturas principais nao conectadas.'},
    FastDecayUnsnappedCoreStructures: { type: 'boolean', default: false , description: 'Acelera a deterioracao de estruturas principais nao conectadas a base.'},
    DestroyUnconnectedWaterPipes: { type: 'boolean', default: false , description: 'Destrói tubulações de água desconectadas da fonte.',
      },
    StructurePreventResourceRadiusMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador do raio que impede a colocacao de recursos ao redor de estruturas.'},
    MaxPlatformSaddleStructureLimit: { type: 'number', default: 25, min: 1, max: 1000 , description: 'Limite maximo de estruturas permitidas em plataformas com sela de plataforma.'},
    PerPlatformMaxStructuresMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador do limite de estruturas por plataforma de sela.'},
    PlatformSaddleBuildAreaBoundsMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador dos limites da area de construcao em selas de plataforma.'},
    OverrideStructurePlatformPrevention: { type: 'boolean', default: false , description: 'Sobrescreve a prevencao padrao de colocar estruturas em plataformas.'},
    EnableExtraStructurePreventionVolumes: { type: 'boolean', default: false , description: 'Ativa volumes extras que impedem a construcao de estruturas em certas areas.'},
    AllowCaveBuildingPvE: { type: 'boolean', default: false , description: 'Permite construção em cavernas em modo PvE.',
      },
    AllowCaveBuildingPvP: { type: 'boolean', default: false , description: 'Permite a construcao de estruturas dentro de cavernas em servidores PvP.'},
    PvEAllowStructuresAtSupplyDrops: { type: 'boolean', default: false , description: 'Permite colocar estruturas nos pontos de spawn de baus de suprimentos em PvE.'},
    AllowCrateSpawnsOnTopOfStructures: { type: 'boolean', default: false , description: 'Permite spawn de baús sobre estruturas.',
      },
    bAllowPlatformSaddleMultiFloors: { type: 'boolean', default: false , description: 'Permite construir multiplos andares de plataforma em cima de selas de plataforma.'},
    MaxGateFrameOnSaddles: { type: 'number', default: 0, min: 0, max: 100 , description: 'Numero maximo de portoes de estrutura que podem ser colocados em selas.'}
  },
  tribeSettings: {
    MaxNumberOfPlayersInTribe: { type: 'number', default: 0, min: 0, max: 500 , description: 'Máximo de jogadores por tribo. 0 = ilimitado.',
      },
    TribeNameChangeCooldown: { type: 'number', default: 15, min: 0, max: 1440 , description: 'Tempo de recarga em minutos entre mudancas de nome de tribo.'},
    PreventTribeAlliances: { type: 'boolean', default: false , description: 'Impede que tribos formem aliancas entre si no servidor.'},
    MaxAlliancesPerTribe: { type: 'number', default: 10, min: 0, max: 100 , description: 'Numero maximo de aliancas que uma tribo pode ter simultaneamente.'},
    MaxTribesPerAlliance: { type: 'number', default: 10, min: 0, max: 100 , description: 'Numero maximo de tribos permitidas em uma unica alianca.'}
  },
  breedingSettings: {
    AllowAnyoneBabyImprintCuddle: { type: 'boolean', default: false , description: 'Permite que qualquer jogador abrace/besar bebês para imprinting.',
      },
    DisableImprintDinoBuff: { type: 'boolean', default: false , description: 'Desabilita o bonus de stats dado pelo imprinting em bebes dinossauros.'},
    BabyImprintingStatScaleMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1, description: 'Multiplicador do bônus de stats ao imprintar em bebês.' }
  },
  itemSettings: {
    ClampItemSpoilingTimes: { type: 'boolean', default: false , description: 'Limita o tempo de deterioração de itens food.',
      },
    ClampResourceHarvestDamage: { type: 'boolean', default: false , description: 'Limita o dano causado aos recursos ao serem coletados.'},
    UseOptimizedHarvestingHealth: { type: 'boolean', default: true , description: 'Usa o calculo otimizado de vida dos recursos ao coletar.'},
    BanListURL: { type: 'text', default: 'http://arkdedicated.com/banlist.txt', description: 'URL da lista de banimento externa (ban list).' }
  },
  performanceSettings: {
    AutoSavePeriodMinutes: { type: 'number', default: 15.0, min: 1, max: 60, step: 0.1 , description: 'Intervalo em minutos para salvamento automático do servidor.',
      },
    KickIdlePlayersPeriod: { type: 'number', default: 3600, min: 0, max: 86400 , description: 'Tempo em segundos antes de desconectar jogadores ociosos sem atividade.'},
    ListenServerTetherDistanceMultiplier: { type: 'number', default: 1.0, min: 0.1, max: 10, step: 0.1 , description: 'Multiplicador da distancia de tethering do servidor ouvinte para jogadores.'},
    RCONServerGameLogBuffer: { type: 'number', default: 600, min: 100, max: 10000 , description: 'Tamanho do buffer do log de jogo do servidor RCON em linhas.'},
    NPCNetworkStasisRangeScalePlayerCountStart: { type: 'number', default: 0, min: 0, max: 1000 , description: 'Numero de jogadores a partir do qual a estase de NPC comeca a ser aplicada.'},
    NPCNetworkStasisRangeScalePlayerCountEnd: { type: 'number', default: 0, min: 0, max: 1000 , description: 'Numero de jogadores no qual a estase de NPC atinge seu alcance maximo.'},
    NPCNetworkStasisRangeScalePercentEnd: { type: 'number', default: 0, min: 0, max: 100, step: 0.1 , description: 'Percentual do alcance maximo de estase de NPC no pico de jogadores.'}
  },
  diseaseSettings: {
    PreventDiseases: { type: 'boolean', default: false , description: 'Desabilita a propagacao de doenças entre jogadores e criaturas no servidor.'},
    NonPermanentDiseases: { type: 'boolean', default: false , description: 'Torna as doenças temporarias, curando-se automaticamente apos um tempo.'},
    PreventSpawnAnimations: { type: 'boolean', default: false, description: 'Previne animação de spawn ao conectar no servidor.' }
  },
  offlineRaidSettings: {
    PreventOfflinePvP: { type: 'boolean', default: false , description: 'Protege jogadores offline contra ataques PvP de outros jogadores.'},
    PreventOfflinePvPInterval: { type: 'number', default: 900, min: 0, max: 3600 , description: 'Periodo em segundos durante o qual um jogador offline esta protegido contra ataques PvP.'}
  },
  crossArkSettings: {
    NoTributeDownloads: { type: 'boolean', default: false , description: 'Impede o download de tributos enviados de outros servidores ARK cruzados.'},
    PreventDownloadSurvivors: { type: 'boolean', default: false , description: 'Bloqueia o download de sobreviventes de outros servidores ARK cruzados.'},
    PreventDownloadItems: { type: 'boolean', default: false , description: 'Bloqueia o download de itens de outros servidores ARK cruzados.'},
    PreventDownloadDinos: { type: 'boolean', default: false , description: 'Bloqueia o download de dinossauros de outros servidores ARK cruzados.'},
    PreventUploadSurvivors: { type: 'boolean', default: false , description: 'Bloqueia o upload de sobreviventes para outros servidores ARK cruzados.'},
    PreventUploadItems: { type: 'boolean', default: false , description: 'Bloqueia o upload de itens para outros servidores ARK cruzados.'},
    PreventUploadDinos: { type: 'boolean', default: false , description: 'Bloqueia o upload de dinossauros para outros servidores ARK cruzados.'},
    CrossARKAllowForeignDinoDownloads: { type: 'boolean', default: false , description: 'Permite download de dinossauros de outros servidores ARK cruzados.',
      },
    MaxTributeDinos: { type: 'number', default: 20, min: 1, max: 500 , description: 'Numero maximo de dinossauros permitidos em tributos entre servidores ARK cruzados.'},
    MaxTributeItems: { type: 'number', default: 50, min: 1, max: 500 , description: 'Numero maximo de itens permitidos em tributos entre servidores ARK cruzados.'},
    MinimumDinoReuploadInterval: { type: 'number', default: 0, min: 0, max: 86400 , description: 'Intervalo minimo em segundos entre reenvios do mesmo dinossauro via tributo.'},
    TributeItemExpirationSeconds: { type: 'number', default: 86400, min: 0, max: 604800 , description: 'Tempo em segundos antes que itens de tributo expirem e sejam removidos.'},
    TributeDinoExpirationSeconds: { type: 'number', default: 86400, min: 0, max: 604800 , description: 'Tempo em segundos antes que dinossauros de tributo expirem e sejam removidos.'},
    TributeCharacterExpirationSeconds: { type: 'number', default: 86400, min: 0, max: 604800 , description: 'Tempo em segundos antes que personagens de tributo expirem e sejam removidos.'}
  },
  flyerSettings: {
    AllowFlyingStaminaRecovery: { type: 'boolean', default: true , description: 'Recuperação de stamina automática para criaturas voadoras.',
      },
    ForceFlyerExplosives: { type: 'boolean', default: false , description: 'Forca que todas as criaturas voadoras carreguem explosivos automaticamente.'}
  },
  advancedSettings: {
    AllowMultipleAttachedC4: { type: 'boolean', default: false , description: 'Permite múltiplos C4s anexados ao mesmo structure.',
      },
    AllowIntegratedSPlusStructures: { type: 'boolean', default: true , description: 'Permite a construcao de estruturas integradas do mod S+ no servidor.'},
    AllowHideDamageSourceFromLogs: { type: 'boolean', default: false , description: 'Oculta a fonte de dano nos logs.',
      },
    AllowSharedConnections: { type: 'boolean', default: false , description: 'Permite conexoes compartilhadas entre jogadores para transferencia de dados.'},
    bFilterTribeNames: { type: 'boolean', default: false , description: 'Ativa o filtro de nomes de tribo para conteudo inadequado.'},
    bFilterCharacterNames: { type: 'boolean', default: false , description: 'Ativa o filtro de nomes de personagens para conteudo inadequado.'},
    bFilterChat: { type: 'boolean', default: false , description: 'Ativa o filtro de chat para bloquear conteudo inadequado.'},
    EnableCryoSicknessPVE: { type: 'boolean', default: true , description: 'Ativa o efeito de doença de criopoda em servidores PvE ao destrancar criaturas.'},
    EnableCryopodNerf: { type: 'boolean', default: false , description: 'Ativa a reducao de poder das criaturas armazenadas em criopodas.'},
    CryopodNerfDuration: { type: 'number', default: 10, min: 0, max: 3600 , description: 'Duração do nerf de dano para criaturas em criopodas.',
      },
    CryopodNerfDamageMult: { type: 'number', default: 0.01, min: 0, max: 1, step: 0.01 , description: 'Multiplicador do dano de criaturas armazenadas em criopodas.'},
    CryopodNerfIncomingDamageMultPercent: { type: 'number', default: 0.25, min: 0, max: 1, step: 0.01 , description: 'Percentual de redução do dano recebido para criaturas em criopoda.',
      },
    DisableCryopodEnemyCheck: { type: 'boolean', default: false , description: 'Desabilita a verificacao de inimigos ao liberar criaturas de criopodas.'},
    DisableCryopodFridgeRequirement: { type: 'boolean', default: false , description: 'Desabilita requisito de criopoda tipo fridge para armazenar comida.',
      },
    AllowCryoFridgeOnSaddle: { type: 'boolean', default: false , description: 'Permite colocar criopodas do tipo geladeira em selas para armazenar comida em viagem.'},
    MaxTrainCars: { type: 'number', default: 8, min: 1, max: 20 , description: 'Numero maximo de vagoes de trem que podem ser acoplados a uma locomotiva.'},
    MaxHexagonsPerCharacter: { type: 'number', default: 2000000000, min: 0, max: 2000000000 , description: 'Quantidade maxima de hexagonos permitidos por personagem no mapa.'},
    AllowTekSuitPowersInGenesis: { type: 'boolean', default: false , description: 'Permite poderes do Traje Tek em mapas Genesis.',
      },
    CustomDynamicConfigUrl: { type: 'text', default: '' , description: 'URL personalizada para o script de configuracao dinamica do servidor.'}
  }
};

// Helper functions
function getGameUserSettingsParamsByCategory(categoryKey: GameUserSettingsCategoryKey): Record<string, GameUserSettingsParam> {
  return gameUserSettingsParams[categoryKey] || {};
}

function getAllGameUserSettingsCategories(): GameUserSettingsCategoryKey[] {
  return Object.keys(gameUserSettingsParams) as GameUserSettingsCategoryKey[];
}

// Helper functions moved inside component to access translations

interface GameUserSettingsEditorProps {
  value?: string;
  onChange?: (value: string) => void;
}

export function GameUserSettingsEditor({ value, onChange }: GameUserSettingsEditorProps) {
  const t = useTranslations('servers.editor');
  const tDefaultValues = useTranslations('servers.defaultValues');
  const tCategories = useTranslations('servers.gameUserSettingsCategories');
  const tParams = useTranslations('servers.gameUserSettingsParams');
  const [editMode, setEditMode] = useState<'visual' | 'text'>('visual');
  const [textContent, setTextContent] = useState('');
  const [visualConfig, setVisualConfig] = useState<Record<string, string | number | boolean>>({});
  const [showPasswords, setShowPasswords] = useState<Record<string, boolean>>({});
  const [activeTab, setActiveTab] = useState<GameUserSettingsCategoryKey>('serverBasic');
  // Raw text held while a number field is being typed, so that intermediate states
  // that are not valid numbers ('', '-', '1.') survive instead of being swallowed.
  const [numberDrafts, setNumberDrafts] = useState<Record<string, string>>({});

  // Latest config, readable synchronously from event handlers without putting a
  // side effect inside a state updater (updaters must stay pure - React may call
  // them twice under StrictMode / concurrent replay).
  const visualConfigRef = useRef<Record<string, string | number | boolean>>({});
  const syncTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Helper functions with i18n support
  const getCategoryName = (categoryKey: GameUserSettingsCategoryKey): string => {
    try {
      return tCategories(categoryKey);
    } catch {
      return categoryKey;
    }
  };

  const getParamDisplayName = (paramKey: string): string => {
    try {
      return tParams(paramKey);
    } catch {
      // Fallback to converting camelCase to readable format
      return paramKey
        .replace(/([A-Z])/g, ' $1')
        .replace(/^./, str => str.toUpperCase())
        .trim();
    }
  };



  const parseTextToVisual = useCallback((content: string) => {
    // Initialize
    const defaultConfig: Record<string, string | number | boolean> = {};
    getAllGameUserSettingsCategories().forEach(categoryKey => {
      const section = gameUserSettingsParams[categoryKey];
      if (section) {
        Object.keys(section).forEach(paramKey => {
          const param = section[paramKey];
          if (param && param.default !== undefined) {
            defaultConfig[paramKey] = param.default;
          }
        });
      }
    });

    // Add SessionSettings and MessageOfTheDay defaults
    defaultConfig.SessionName = tDefaultValues('sessionName');
    defaultConfig.Message = tDefaultValues('message');
    defaultConfig.Duration = 30;

    if (!content) {
      // If，
      setVisualConfig(defaultConfig);
      return;
    }

    try {
      const values = extractConfigValues(content);
      // 
      const mergedConfig = { ...defaultConfig, ...values };
      setVisualConfig(mergedConfig);
    } catch (error) {
      console.error('Failed to parse GameUserSettings.ini, falling back to defaults:', error);
      // Parse
      setVisualConfig(defaultConfig);
    }
  }, [tDefaultValues]);

  // Keep the mirror in step with state updates that did not come from handleVisualChange.
  useEffect(() => {
    visualConfigRef.current = visualConfig;
  }, [visualConfig]);

  // Drop any pending sync when the editor unmounts.
  useEffect(() => {
    return () => {
      if (syncTimerRef.current !== null) {
        clearTimeout(syncTimerRef.current);
        syncTimerRef.current = null;
      }
    };
  }, []);

  // InitializeSettings
  useEffect(() => {
    parseTextToVisual('');
  }, [parseTextToVisual]);

  useEffect(() => {
    setTextContent(value || '');
    // Parse
    if (value) {
      parseTextToVisual(value);
    }
  }, [value, parseTextToVisual]);

  const extractConfigValues = (content: string): Record<string, string | number | boolean> => {
    const values: Record<string, string | number | boolean> = {};
    const lines = content.split('\n');

    lines.forEach(line => {
      line = line.trim();

      // Handlesection
      if (line.startsWith('[') && line.endsWith(']')) {
        return;
      }

      // 
      if (!line || line.startsWith(';') || !line.includes('=')) {
        return;
      }

      const [key, ...valueParts] = line.split('=');
      const value = valueParts.join('=').trim();
      const keyName = key.trim();

      if (keyName && value !== undefined) {
        // 
        if (value.toLowerCase() === 'true') {
          values[keyName] = true;
        } else if (value.toLowerCase() === 'false') {
          values[keyName] = false;
        } else if (!isNaN(Number(value)) && value !== '') {
          values[keyName] = Number(value);
        } else {
          values[keyName] = value;
        }
      }
    });

    return values;
  };

  const syncVisualToTextWithConfig = useCallback((config: Record<string, string | number | boolean>) => {
    try {
      // Build INI content matching GameUserSettings.ini format
      let iniContent = '';

      // ServerSettings section
      iniContent += '[ServerSettings]\n';
      const serverBasicParams = getGameUserSettingsParamsByCategory('serverBasic');
      Object.keys(serverBasicParams).forEach(key => {
        const value = config[key];
        if (value !== undefined && value !== '') {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Game mode settings
      const gameModeParams = getGameUserSettingsParamsByCategory('gameMode');
      Object.keys(gameModeParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Communication settings
      const communicationParams = getGameUserSettingsParamsByCategory('communication');
      Object.keys(communicationParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Game multipliers
      const gameMultipliersParams = getGameUserSettingsParamsByCategory('gameMultipliers');
      Object.keys(gameMultipliersParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Character settings
      const characterSettingsParams = getGameUserSettingsParamsByCategory('characterSettings');
      Object.keys(characterSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Dino settings
      const dinoSettingsParams = getGameUserSettingsParamsByCategory('dinoSettings');
      Object.keys(dinoSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Environment settings
      const environmentSettingsParams = getGameUserSettingsParamsByCategory('environmentSettings');
      Object.keys(environmentSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Structure settings
      const structureSettingsParams = getGameUserSettingsParamsByCategory('structureSettings');
      Object.keys(structureSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Tribe settings
      const tribeSettingsParams = getGameUserSettingsParamsByCategory('tribeSettings');
      Object.keys(tribeSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Breeding settings
      const breedingSettingsParams = getGameUserSettingsParamsByCategory('breedingSettings');
      Object.keys(breedingSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Item settings
      const itemSettingsParams = getGameUserSettingsParamsByCategory('itemSettings');
      Object.keys(itemSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Performance settings
      const performanceSettingsParams = getGameUserSettingsParamsByCategory('performanceSettings');
      Object.keys(performanceSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Disease settings
      const diseaseSettingsParams = getGameUserSettingsParamsByCategory('diseaseSettings');
      Object.keys(diseaseSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Offline raid settings
      const offlineRaidSettingsParams = getGameUserSettingsParamsByCategory('offlineRaidSettings');
      Object.keys(offlineRaidSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Cross ARK settings
      const crossArkSettingsParams = getGameUserSettingsParamsByCategory('crossArkSettings');
      Object.keys(crossArkSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Flyer settings
      const flyerSettingsParams = getGameUserSettingsParamsByCategory('flyerSettings');
      Object.keys(flyerSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Advanced settings
      const advancedSettingsParams = getGameUserSettingsParamsByCategory('advancedSettings');
      Object.keys(advancedSettingsParams).forEach(key => {
        const value = config[key];
        if (value !== undefined) {
          iniContent += `${key}=${value}\n`;
        }
      });

      // Add other required sections
      iniContent += '\n[/Script/Engine.GameSession]\n';
      iniContent += 'MaxPlayers=70\n\n';

      iniContent += '[SessionSettings]\n';
      iniContent += `SessionName=${config.SessionName || tDefaultValues('sessionName')}\n\n`;

      iniContent += '[MessageOfTheDay]\n';
      iniContent += `Message=${config.Message || tDefaultValues('message')}\n`;
      iniContent += `Duration=${config.Duration || 30}\n`;

      setTextContent(iniContent);
      onChange?.(iniContent);
    } catch (error) {
      console.error('Failed to generate GameUserSettings.ini from the visual editor:', error);
    }
  }, [onChange, tDefaultValues]);

  const syncVisualToText = useCallback(() => {
    syncVisualToTextWithConfig(visualConfig);
  }, [visualConfig, syncVisualToTextWithConfig]);



  const handleTextChange = (newContent: string) => {
    setTextContent(newContent);
    onChange?.(newContent);
  };

  const handleModeSwitch = async (mode: 'visual' | 'text') => {
    if (mode === editMode) return;

    try {
      if (mode === 'text' && editMode === 'visual') {
        // From ：
        syncVisualToText();
      } else if (mode === 'visual' && editMode === 'text') {
        // From ：
        parseTextToVisual(textContent);
      }

      setEditMode(mode);
    } catch (error) {
      console.error('Failed to switch GameUserSettings edit mode:', error);
    }
  };

  const handleVisualChange = (paramKey: string, value: string | number | boolean) => {
    const newConfig = { ...visualConfigRef.current, [paramKey]: value };
    visualConfigRef.current = newConfig;
    setVisualConfig(newConfig);

    // The sync is a side effect (it calls onChange), so it is scheduled here rather
    // than from inside the state updater. Coalesce bursts of keystrokes.
    if (syncTimerRef.current !== null) {
      clearTimeout(syncTimerRef.current);
    }
    syncTimerRef.current = setTimeout(() => {
      syncTimerRef.current = null;
      syncVisualToTextWithConfig(visualConfigRef.current);
    }, 0);
  };

  const togglePasswordVisibility = (paramKey: string) => {
    setShowPasswords(prev => ({ ...prev, [paramKey]: !prev[paramKey] }));
  };

  const renderParamControl = (paramKey: string, param: GameUserSettingsParam) => {
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

      case 'number':
        return (
          <div className="space-y-1">
            <Input
              type="number"
              value={numberDrafts[paramKey] ?? String(currentValue)}
              onChange={(e) => {
                const raw = e.target.value;
                // Keep the raw text so the field can be cleared and a decimal point
                // (or leading '-') can be typed without the value snapping back.
                setNumberDrafts(prev => ({ ...prev, [paramKey]: raw }));
                const parsed = Number.parseFloat(raw);
                if (!Number.isNaN(parsed)) {
                  handleVisualChange(paramKey, parsed);
                }
              }}
              onBlur={() => {
                // Hand control back to the committed value once editing stops.
                setNumberDrafts(prev => {
                  if (!(paramKey in prev)) return prev;
                  const next = { ...prev };
                  delete next[paramKey];
                  return next;
                });
              }}
              min={param.min}
              max={param.max}
              step={param.step || 1}
              className="w-full"
            />
            {(param.min !== undefined || param.max !== undefined) && (
              <div className="text-xs text-muted-foreground">
                {t('range')}: {param.min ?? '∞'} - {param.max ?? '∞'}
              </div>
            )}
          </div>
        );

      case 'text':
        return (
          <Input
            type="text"
            value={String(currentValue)}
            onChange={(e) => handleVisualChange(paramKey, e.target.value)}
            className="w-full"
          />
        );

      case 'password':
        return (
          <div className="relative">
            <Input
              type={!showPasswords[paramKey] ? 'password' : 'text'}
              value={String(currentValue)}
              onChange={(e) => handleVisualChange(paramKey, e.target.value)}
              className="w-full pr-10"
            />
            <button
              type="button"
              onClick={() => togglePasswordVisibility(paramKey)}
              className="absolute inset-y-0 right-0 pr-3 flex items-center"
            >
              {showPasswords[paramKey] ? (
                <EyeOff className="h-4 w-4 text-gray-500" />
              ) : (
                <Eye className="h-4 w-4 text-gray-500" />
              )}
            </button>
          </div>
        );

      default:
        return null;
    }
  };

  const getAllCategories = () => getAllGameUserSettingsCategories();

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

      {/* Visual Edit Mode */}
      {editMode === 'visual' && (
        <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as GameUserSettingsCategoryKey)}>
          <TabsList className="grid w-full grid-cols-3 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-9 gap-2 h-auto p-2 bg-muted/50 border rounded-lg">
            {getAllCategories().map((categoryKey) => (
              <TabsTrigger 
                key={categoryKey} 
                value={categoryKey} 
                className="text-xs px-3 py-2 flex-shrink-0 border-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-primary data-[state=active]:text-primary-foreground data-[state=active]:shadow-md hover:border-muted-foreground/50 hover:bg-muted transition-all duration-200 font-medium rounded-md"
              >
                {getCategoryName(categoryKey)}
              </TabsTrigger>
            ))}
          </TabsList>

          {getAllCategories().map((categoryKey) => {
            const params = getGameUserSettingsParamsByCategory(categoryKey);

            return (
              <TabsContent key={categoryKey} value={categoryKey} className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      {getCategoryName(categoryKey)}
                    </CardTitle>
                    <CardDescription>
                      {Object.keys(params).length}{t('parametersCount')}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid gap-4">
                      {Object.entries(params).map(([paramKey, param]) => (
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
                                    <p className="text-sm mt-1">{getParamDisplayName(paramKey)}</p>
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
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>
            );
          })}
        </Tabs>
      )}

      {/* Text Edit Mode */}
      {editMode === 'text' && (
        <div className="space-y-2">
          <Label htmlFor="gameusersettings-text">{t('textEditMode')}</Label>
          <Textarea
            id="gameusersettings-text"
            value={textContent}
            onChange={(e) => handleTextChange(e.target.value)}
            className="min-h-[400px] font-mono text-sm"
            placeholder={t('placeholders.gameUserSettings')}
          />
        </div>
      )}
    </div>
  );
}