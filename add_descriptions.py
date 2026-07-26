#!/usr/bin/env python3
"""Add Portuguese descriptions to gameIniParams missing them in GameIniEditor.tsx.

For each parameter in gameIniParams that lacks a `description:` field, insert a
concise 1-sentence Portuguese description inside the object literal, before the
closing `}` brace. Idempotent: skips params that already have a description.
"""
import re, sys

FILE = "ui/src/components/servers/GameIniEditor.tsx"

DESCS = {
    "bDisableStructurePlacementCollision":
        "Permite que estruturas sejam colocadas sobrepostas, desativando a verificação de colisão de posicionamento.",
    "bAllowFlyerCarryPvE":
        "Permite que criaturas voadoras carreguem dinos selvagens no modo PvE.",
    "bDisableStructureDecayPvE":
        "Desativa os temporizadores automáticos de deterioração de estruturas no PvE.",
    "bAllowUnlimitedRespecs":
        "Permite redefinições ilimitadas de atributos e engramas usando o Mindwipe.",
    "bAllowPlatformSaddleMultiFloors":
        "Permite construir múltiplos andares em selas plataforma como Quetzal e Bronto.",
    "bPassiveDefensesDamageRiderlessDinos":
        "Faz com que defesas passivas causem dano a dinos sem montaria.",
    "bPvEDisableFriendlyFire":
        "Desativa completamente o dano amigável no modo PvE.",
    "bDisableFriendlyFire":
        "Desativa o dano amigável em todos os modos de jogo.",
    "bEnablePvPGamma":
        "Permite ajustar o brilho e gama durante o modo PvP.",
    "DifficultyOffset":
        "Controla a distribuição de níveis dos dinos selvagens. 1.0 = nível máximo 150.",
    "OverrideOfficialDifficulty":
        "Substitui a dificuldade oficial, alterando o nível máximo dos dinos selvagens.",
    "XPMultiplier":
        "Multiplicador de toda experiência ganha. 2.0 dobra o XP recebido.",
    "PlayerCharacterWaterDrainMultiplier":
        "Multiplicador da taxa de consumo de água do jogador. Menor = drena mais devagar.",
    "PlayerCharacterFoodDrainMultiplier":
        "Multiplicador da taxa de consumo de comida do jogador. Menor = drena mais devagar.",
    "PlayerCharacterStaminaDrainMultiplier":
        "Multiplicador da taxa de uso de estamina do jogador. Menor = cansa mais devagar.",
    "PlayerCharacterHealthRecoveryMultiplier":
        "Multiplicador da velocidade de regeneração passiva de vida do jogador.",
    "MatingIntervalMultiplier":
        "Multiplicador do intervalo de acasalamento entre reproduções. Menor = intervalos mais curtos.",
    "EggHatchSpeedMultiplier":
        "Multiplicador da velocidade de incubação e eclosão de ovos. Maior = eclosão mais rápida.",
    "BabyMatureSpeedMultiplier":
        "Multiplicador da velocidade de maturação de filhotes. Maior = crescimento mais rápido.",
    "BabyFoodConsumptionSpeedMultiplier":
        "Multiplicador da taxa de consumo de comida de filhotes. Menor = precisam de menos comida.",
    "BabyCuddleIntervalMultiplier":
        "Multiplicador do intervalo de carinho para imprinting. Menor = imprinting mais frequente.",
    "BabyCuddleGracePeriodMultiplier":
        "Multiplicador do período de carência antes de perder qualidade do imprint dos filhotes.",
    "BabyCuddleLoseImprintQualitySpeedMultiplier":
        "Multiplicador da velocidade de perda de qualidade de imprint quando filhotes não recebem carinho.",
    "HarvestAmountMultiplier":
        "Multiplicador da quantidade de recursos colhidos. Maior = mais recursos por colheita.",
    "HarvestHealthMultiplier":
        "Multiplicador da vida dos recursos minerais e plantas ao serem colhidos.",
    "ResourcesRespawnPeriodMultiplier":
        "Multiplicador do período de reaparecimento de recursos naturais como pedras e madeira.",
    "ItemStackSizeMultiplier":
        "Multiplicador do tamanho máximo da pilha de cada tipo de item no inventário.",
    "CropGrowthSpeedMultiplier":
        "Multiplicador da velocidade de crescimento de plantações cultivadas.",
    "GlobalItemDecompositionTimeMultiplier":
        "Multiplicador do tempo de decomposição de itens largados no chão.",
    "GlobalCorpseDecompositionTimeMultiplier":
        "Multiplicador do tempo de decomposição de cadáveres de jogadores.",
    "TamingSpeedMultiplier":
        "Multiplicador da velocidade de domação. Maior = domação mais rápida.",
    "DinoCharacterFoodDrainMultiplier":
        "Multiplicador da taxa de consumo de comida de todos os dinos. Menor = comem menos.",
    "DinoCharacterStaminaDrainMultiplier":
        "Multiplicador da taxa de uso de estamina de todos os dinos.",
    "DinoCharacterHealthRecoveryMultiplier":
        "Multiplicador da regeneração passiva de vida dos dinos.",
    "DinoCountMultiplier":
        "Multiplicador da quantidade de dinos selvagens que spawnam. Maior = mais dinos no mapa.",
    "WildDinoCharacterFoodDrainMultiplier":
        "Multiplicador da taxa de consumo de comida de dinos selvagens.",
    "WildDinoTorporDrainMultiplier":
        "Multiplicador da taxa de perda de torpor de dinos selvagens. Menor = ficam inconscientes por mais tempo.",
    "MaxNumberOfPlayersInTribe":
        "Número máximo de jogadores permitidos em uma única tribo. 0 = ilimitado.",
    "TribeNameChangeCooldown":
        "Tempo de espera em dias entre alterações de nome de tribo.",
    "bPvEAllowTribeWar":
        "Permite que tribos declarem guerra a outras tribos no modo PvE.",
    "bPvEAllowTribeWarCancel":
        "Permite que tribos cancelem uma declaração de guerra ativa.",
    "bIncreasePvPRespawnInterval":
        "Aumenta o intervalo de reaparecimento de jogadores no PvP após serem mortos.",
    "IncreasePvPRespawnIntervalCheckPeriod":
        "Período em segundos para verificar a elegibilidade do intervalo de reaparecimento PvP.",
    "IncreasePvPRespawnIntervalMultiplier":
        "Multiplicador aplicado ao temporizador de reaparecimento após mortes em PvP.",
    "IncreasePvPRespawnIntervalBaseAmount":
        "Valor base em segundos do atraso de reaparecimento antes da aplicação do multiplicador PvP.",
    "StructureDamageMultiplier":
        "Multiplicador do dano causado a estruturas. Maior = estruturas recebem mais dano.",
    "StructureResistanceMultiplier":
        "Multiplicador da resistência de estruturas a dano. Maior = estruturas recebem menos dano.",
    "StructureDamageRepairCooldown":
        "Tempo de espera em segundos antes que uma estrutura danificada possa ser reparada.",
    "PvEStructureDecayPeriodMultiplier":
        "Multiplicador da duração do temporizador de deterioração de estruturas no PvE.",
    "MaxStructuresInRange":
        "Número máximo de estruturas permitidas em um raio de proximidade para otimizar o desempenho.",
    "bAutoPvETimer":
        "Ativa a alternância automática entre modos PvP e PvE baseada em temporizador.",
    "bAutoPvEUseSystemTime":
        "Usa o horário do sistema para a alternância automática do modo PvE.",
    "AutoPvEStartTimeSeconds":
        "Horário em segundos (0 a 86400) quando a alternância para PvE é ativada.",
    "AutoPvEStopTimeSeconds":
        "Horário em segundos (0 a 86400) quando a alternância para PvE é desativada.",
    "bOnlyAllowSpecifiedEngrams":
        "Permite apenas os engramas explicitamente habilitados na lista de engramas permitidos.",
    "bAutoUnlockAllEngrams":
        "Desbloqueia automaticamente todos os engramas para todos os jogadores ao entrar no servidor.",
    "bShowCreativeMode":
        "Exibe a categoria de engramas do modo criativo no menu de criação.",
    "bUseCorpseLocator":
        "Ativa um marcador na bússola apontando para a localização da última morte do jogador.",
    "bDisableLootCrates":
        "Desativa o surgimento de baús de saque no mapa do servidor.",
    "bDisableDinoRiding":
        "Impede que jogadores montem em dinos no servidor.",
    "bDisableDinoTaming":
        "Desativa completamente a domação de dinos no servidor.",
    "bAllowCustomRecipes":
        "Permite que jogadores criem receitas de criação personalizadas.",
    "DayCycleSpeedScale":
        "Velocidade do ciclo completo de dia e noite. 2.0 = ciclo duas vezes mais rápido.",
    "NightTimeSpeedScale":
        "Velocidade da fase noturna especificamente. 2.0 = noites mais curtas.",
    "DayTimeSpeedScale":
        "Velocidade da fase diurna especificamente. 2.0 = dias mais curtos.",
}

def main():
    path = sys.argv[1] if len(sys.argv) > 1 else FILE

    with open(path, "r", encoding="utf-8") as f:
        content = f.read()

    lines = content.split("\n")
    count = 0

    for i, line in enumerate(lines):
        stripped = line.rstrip()
        # Match: "    ParamName: { ... }" or "    ParamName: { ... },"
        m = re.match(
            r"^(\s+)((?:b[A-Z]|[A-Z])\w+):\s*(\{.*\})([,]?)\s*$",
            stripped,
        )
        if not m:
            continue
        indent = m.group(1)
        name = m.group(2)
        obj_inner = m.group(3)  # e.g. "{ type: ..., default: ... }"
        trailing = m.group(4)   # "," or ""

        if name not in DESCS:
            continue
        if "description:" in stripped:
            continue

        # Insert description before the last '}' inside the braces
        # obj_inner = "{ type: ..., default: ... }"
        # We need: "{ type: ..., default: ..., description: '...' }"
        inner_no_brace = obj_inner.rstrip("}").rstrip()
        new_obj = inner_no_brace + ", description: '" + DESCS[name] + "' }"
        lines[i] = indent + name + ": " + new_obj + trailing
        count += 1

    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines))

    print(f"Adicionadas descrições em português para {count} parâmetros em {path}")

if __name__ == "__main__":
    main()
