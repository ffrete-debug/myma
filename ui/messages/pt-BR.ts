export default {
  // Common
  common: {
    confirm: 'Confirmar',
    cancel: 'Cancelar',
    save: 'Salvar',
    delete: 'Excluir',
    edit: 'Editar',
    add: 'Adicionar',
    create: 'Criar',
    close: 'Fechar',
    loading: 'Carregando...',
    error: 'Erro',
    success: 'Sucesso',
    warning: 'Aviso',
    info: 'Informação',
    yes: 'Sim',
    no: 'Não',
    back: 'Voltar',
    next: 'Próximo',
    previous: 'Anterior',
    submit: 'Enviar',
    reset: 'Redefinir',
    search: 'Buscar',
    filter: 'Filtrar',
    sort: 'Ordenar',
    refresh: 'Atualizar',
    copy: 'Copiar',
    download: 'Baixar',
    upload: 'Enviado',
    rename: 'Renomear',
    export: 'Exportar',
    import: 'Importar',
    settings: 'Configurações',
    help: 'Ajuda',
    about: 'Sobre',
    version: 'Versão',
    language: 'Idioma',
    theme: 'Tema',
    dark: 'Escuro',
    light: 'Claro',
    auto: 'Automático',
    lines: 'linhas',
    all: 'Todos',
    autoRefreshOn: 'Atualização automática ligada',
    autoRefreshOff: 'Atualização automática desligada'
  },

  // Navigation
  navigation: {
    home: 'Início',
    dashboard: 'Painel',
    servers: 'Gerenciamento de Servidores',
    plugins: 'Plugins',
    metrics: 'Métricas',
    mods: 'Mods',
    players: 'Gerenciamento de Jogadores',
    logs: 'Monitoramento de Logs',
    auditLogs: 'Logs de Auditoria',
    settings: 'Configurações',
    logout: 'Sair',
    welcome: 'Bem-vindo',
    user: 'Usuário'
  },

  // Authentication
  auth: {
    login: 'Entrar',
    logout: 'Sair',
    username: 'Usuário',
    password: 'Senha',
    loginTitle: 'Ark Server Commander',
    loginSubtitle: 'Login seguro na sua conta de gestão',
    loginButton: 'Entrar',
    loginLoading: 'Entrando...',
    loginError: 'Falha no login',
    loginSuccess: 'Login bem-sucedido',
    logoutSuccess: 'Logout realizado com sucesso',
    initCheck: 'Verificando status de inicialização do sistema',
    initRequired: 'Inicialização do sistema necessária',
    alreadyLoggedIn: 'Você já está logado',
    enterUsername: 'Por favor, informe o usuário',
    enterPassword: 'Por favor, informe a senha',
    firstTimeTip: 'Primeira vez? O sistema o guiará automaticamente pela inicialização',
    secureLogin: 'Sistema de Login Seguro',
    // Initialization related
    initTitle: 'Inicialização do Sistema',
    initSubtitle: 'Primeiro uso, configure a conta de administrador',
    adminUsername: 'Nome de Administrador',
    enterAdminUsername: 'Por favor, informe o nome de administrador',
    confirmPassword: 'Confirmar Senha',
    enterConfirmPassword: 'Por favor, informe a senha novamente',
    passwordMinLength: 'A senha deve ter no mínimo 6 caracteres',
    passwordMinLengthError: 'A senha deve ter no mínimo 6 caracteres',
    passwordMismatch: 'As senhas não coincidem',
    initButton: 'Inicializar Sistema',
    initLoading: 'Inicializando...',
    initSuccess: 'Sistema inicializado com sucesso',
    initError: 'Falha ao inicializar o sistema',
    initTip: 'Redirecionará automaticamente para a página principal após a inicialização',
    initWizard: 'Assistente de Inicialização do Sistema'
  },

  // Home page
  home: {
    title: 'Bem-vindo ao ARK Server Commander',
    subtitle: 'Login efetuado com sucesso. Você já pode começar a gerenciar seus servidores ARK.',
    systemInfo: 'Informações do Sistema',
    username: 'Usuário',
    userID: 'ID do Usuário',
    imageManagement: 'Gerenciamento de Imagens',
    features: 'Módulos de Recursos',
    serverManagement: 'Gerenciamento de Servidores',
    serverManagementDesc: 'Adicione, configure e gerencie seus servidores ARK com start/stop one-click e monitoramento',
    startManage: 'Começar a Gerenciar',
    playerManagement: 'Gerenciamento de Jogadores',
    playerManagementDesc: 'Gerencie jogadores dos servidores, visualize status online e permissões',
    logMonitoring: 'Monitoramento de Logs',
    logMonitoringDesc: 'Monitoramento de logs em tempo real, visualize status do sistema e métricas de performance',
    comingSoon: 'Em Breve',
    tip: 'Clique nos cards acima para começar a gerenciar seus servidores ARK'
  },

  // Server management
  servers: {
    title: 'Gerenciamento de Servidores',
    serverManagementDesc: 'Gerencie e monitore suas instâncias de servidor ARK',
    addServer: 'Adicionar Servidor',
    editServer: 'Editar Servidor',
    deleteServer: 'Excluir Servidor',
    serverName: 'Nome do Servidor',
    serverPort: 'Porta do Servidor',
    serverPath: 'Caminho do Servidor',
    serverStatus: 'Status do Servidor',
    serverActions: 'Ações',
    startServer: 'Iniciar Servidor',
    stopServer: 'Parar Servidor',
    restartServer: 'Reiniciar Servidor',
    viewLogs: 'Ver Logs',
    serverConfig: 'Configuração do Servidor',
    gameIni: 'Configuração do Game.ini',
    gameUserSettings: 'Configuração do GameUserSettings.ini',
    serverArgs: 'Argumentos de Inicialização',
    running: 'Executando',
    stopped: 'Parado',
    starting: 'Iniciando',
    stopping: 'Parando',
    error: 'Erro',
    unknown: 'Desconhecido',
    confirmDelete: 'Tem certeza que deseja excluir este servidor?',
    deleteWarning: 'Esta ação não pode ser desfeita',
    serverAdded: 'Servidor adicionado com sucesso',
    serverUpdated: 'Servidor atualizado com sucesso',
    serverDeleted: 'Servidor excluído com sucesso',
    serverStartSuccess: 'Servidor iniciado com sucesso',
    serverStopSuccess: 'Servidor parado com sucesso',
    serverRestartSuccess: 'Servidor reiniciado com sucesso',
    serverStartError: 'Falha ao iniciar servidor',
    serverStopError: 'Falha ao parar servidor',
    serverRestartError: 'Falha ao reiniciar servidor',
    noServers: 'Nenhum servidor ainda',
    noServersDesc: 'Clique em "Adicionar Servidor" para criar seu primeiro servidor ARK',
    serverConfigSaved: 'Configuração do servidor salva',
    serverConfigError: 'Falha ao salvar configuração do servidor',
    invalidPort: 'Número de porta inválido',
    invalidPath: 'Caminho de servidor inválido',
    portInUse: 'Porta já está em uso',
    pathNotExists: 'O caminho do servidor não existe',
    serverNameRequired: 'Nome do servidor é obrigatório',
    serverPortRequired: 'Porta do servidor é obrigatória',
    serverPathRequired: 'Caminho do servidor é obrigatório',
    imageStatus: 'Status da Imagem',
    imageDownloading: 'Baixando Imagem',
    imageNotReady: 'Imagem Não Pronta',
    imageDownloadingDesc: 'Baixando imagem, aguarde antes de criar o servidor',
    imageNotReadyDesc: 'Imagem não está pronta, não é possível criar o servidor',
    // Docker images detailed translations
    dockerImages: {
      title: 'Image Download Status',
      overallStatus: 'Overall Status',
      imageReady: 'Images Ready',
      imageNotReady: 'Images Not Ready (Cannot Start Server)',
      imageMissingManualDownload: 'Images Missing, Please Download Manually',
      downloading: 'Downloading',
      ready: 'Ready',
      cancel: 'Cancelar',
      download: 'Baixar',
      update: 'Atualizar',
      updateAvailable: 'Atualização Disponível',
      notReady: 'Not Ready',
      waitingDownload: 'Waiting Download',
      layerProgress: 'Layer Download Progress',
      totalImages: 'Total Images',
      downloadingCount: 'Downloading',
      refreshStatus: 'Refresh Image Status',
      manualDownload: 'Manual Download',
      checkUpdates: 'Check Updates',
      updateConfirm: 'Image Update Confirmation',
      imageInfo: 'Image Information',
      imageName: 'Image Name',
      affectedServers: 'Affected Servers',
      updateWarning: 'Update Risk Warning',
      warningDownloadTime: 'Image download may take a long time, please be patient',
      warningContainerRecreate: 'Container recreation will cause brief server downtime',
      warningDataSafety: 'Please ensure important data is backed up to avoid data loss',
      updateOptions: 'Update Options',
      updateImageOnly: 'Update Image Only',
      updateImageOnlyDesc: 'Only download new image, do not recreate containers. Manual container recreation required to use new image.',
      updateAndRecreate: 'Update Image and Recreate Containers',
      updateAndRecreateDesc: 'Download new image and automatically recreate all affected containers. Servers will be briefly offline.',
      confirmUpdate: 'Confirm Update',
      unknownSize: 'Unknown Size',
      // Image names
      arkServer: 'ARK Server',
      alpineSystem: 'Alpine System',
      // Layer information
      layerDetails: 'Layer Details',
      layers: 'Layers',
      // Layer status
      layerStatus: {
        pending: 'Pending',
        downloading: 'Downloading',
        extracting: 'Extracting',
        verifying: 'Verifying',
        complete: 'Complete'
      }
    },
    cannotDeleteRunning: 'Cannot delete running server, please stop it first',
    serverCreateSuccess: 'Server created successfully',
    serverUpdateSuccess: 'Server updated successfully',
    serverDeleteSuccess: 'Server deleted successfully',
    serverStartInProgress: 'Server starting...',
    serverStopInProgress: 'Server stopping...',
    copyToClipboard: 'Copied to clipboard',
    copyFailed: 'Copy failed, please copy manually',
    authenticationFailed: 'Authentication failed, please login again',
    serverLogs: 'Server Logs',
    noLogs: 'No logs yet',
    getServerListFailed: 'Failed to get server list, please try again later',
    loadServerInfoFailed: 'Failed to load server info, please try again later',
    configUnreadable: 'Não foi possível ler os arquivos de configuração reais deste servidor, então os editores abaixo podem estar vazios. O salvamento foi desativado para não sobrescrever a configuração real. Verifique se o volume do servidor existe e se o Docker está acessível.',
    fetchLogsFailed: 'Falha ao obter os logs, tente novamente mais tarde',
    serverUpdateError: 'Falha ao atualizar o servidor, tente novamente mais tarde',
    operationFailed: 'Operation failed, please try again later',
    deleteFailed: 'Delete failed, please try again later',
    startServerFailed: 'Failed to start server, please try again later',
    stopServerFailed: 'Failed to stop server, please try again later',
    imageStatusError: 'Failed to get image status',
    // Operações em lote
    selected: 'selecionados',
    selectAll: 'Selecionar Todos',
    bulkSelected: '{count} servidor(es) selecionado(s)',
    bulkStart: 'Iniciar em Lote',
    bulkStop: 'Parar em Lote',
    bulkRestart: 'Reiniciar em Lote',
    clearSelection: 'Limpar Seleção',
    page: 'Página',
    next: 'Próximo',
    previous: 'Anterior',
    // Backup e Restauração
    backup: 'Backup',
    restore: 'Restaurar',
    createBackup: 'Criar Backup',
    backupCreated: 'Backup criado com sucesso',
    backupFailed: 'Falha no backup',
    backupDownload: 'Baixar Backup',
    backupDelete: 'Excluir Backup',
    backupDeleteConfirm: 'Tem certeza que deseja excluir este backup?',
    backupList: 'Backups',
    backupNoBackups: 'Nenhum backup ainda',
    backupFileSize: 'Tamanho do arquivo',
    backupStatus: 'Status',
    backupInProgress: 'Em andamento',
    backupCompleted: 'Concluído',
    backupFailedStatus: 'Falhou',
    backupRestore: 'Restaurar a partir do backup',
    backupRestoreConfirm: 'Tem certeza que deseja restaurar a partir deste backup? O servidor será recriado com os dados do backup.',
    // Relacionado ao card do servidor
    card: {
      startServer: 'Iniciar Servidor',
      stopServer: 'Parar Servidor',
      running: 'Executando',
      stopped: 'Parado',
      starting: 'Iniciando',
      stopping: 'Parando',
      restarting: 'Reiniciando',
      error: 'Erro',
      unknown: 'Desconhecido',
      startingEllipsis: 'Iniciando...',
      stoppingEllipsis: 'Parando...',
      plugins: 'Plugins',
      mods: 'Mods',
      unknownStatus: 'Status Desconhecido',
      cannotStartImageNotReady: 'Imagem não pronta, não é possível iniciar',
      rconInfo: 'Informações RCON',
      rconConnectionInfo: 'Informações de Conexão RCON',
      serverIdentifier: 'Identificador do Servidor',
      rconPort: 'Porta RCON',
      serverPort: 'Porta do Servidor',
      adminPassword: 'Senha de Admin',
      editServer: 'Editar Servidor',
      deleteServer: 'Excluir Servidor',
      confirmDelete: 'Confirmar Exclusão',
      confirmDeleteMessage: 'Tem certeza que deseja excluir o servidor "{identifier}"? Esta ação não pode ser desfeita.',
      status: 'Status',
      serverName: 'Nome do Servidor',
      clusterId: 'ID do Cluster',
      map: 'Mapa',
      maxPlayers: 'Máx. Jogadores',
      portConfig: 'Configuração de Portas',
      gamePort: 'Porta de Jogo',
      queryPort: 'Porta de Query',
      rconPortLabel: 'Porta RCON',
      authInfo: 'Info de Autenticação',
      timeInfo: 'Informações de Tempo',
      createdAt: 'Criado em',
      updatedAt: 'Atualizado em',
      serverId: 'ID do Servidor',
      copy: 'Copiar',
      close: 'Fechar',
      showPassword: 'Mostrar Senha',
      hidePassword: 'Ocultar Senha'
    },
    // Server edit related
    edit: {
      title: 'Edição do Servidor',
      loadServerInfoFailed: 'Falha ao carregar as informações do servidor, tente novamente mais tarde',
      createTitle: 'Adicionar Servidor',
      editTitle: 'Editar Servidor',
      createServerDesc: 'Configure e crie uma nova instância de servidor ARK',
      basicParams: 'Parâmetros Básicos',
      gameUserSettings: 'GameUserSettings.ini',
      gameIni: 'Game.ini',
      serverArgs: 'Argumentos de Inicialização (SERVER_ARGS)',
      serverIdentifier: 'Identificador do Servidor',
      serverIdentifierRequired: 'Identificador do Servidor *',
      serverIdentifierPlaceholder: 'Informe o identificador do servidor',
      serverName: 'Nome do Servidor',
      serverNamePlaceholder: 'Informe o nome do servidor',
      serverNameDesc: 'Nome exibido na lista de servidores do jogo',
      clusterId: 'ID do Cluster',
      clusterIdPlaceholder: 'Informe o ID do cluster (opcional)',
      clusterIdDesc: 'Para compartilhamento de dados entre servidores do cluster',
      gamePort: 'Porta de Jogo',
      gamePortRequired: 'Porta de Jogo *',
      gamePortPlaceholder: '7777',
      queryPort: 'Porta de Query',
      queryPortRequired: 'Porta de Query *',
      queryPortPlaceholder: '27015',
      rconPort: 'Porta RCON',
      rconPortRequired: 'Porta RCON *',
      rconPortPlaceholder: '32330',
      map: 'Mapa',
      mapPlaceholder: 'Selecione o mapa',
      maxPlayers: 'Máx. Jogadores',
      maxPlayersPlaceholder: '70',
      maxPlayersDesc: 'Número máximo de jogadores (1-200)',
      modIds: 'IDs de Mods',
      modIdsPlaceholder: 'Informe IDs de mods, separados por vírgulas (ex.: 123456,789012)',
      modIdsDesc: 'IDs da Steam Workshop; vários mods separados por vírgulas',
      adminPassword: 'Senha de Admin',
      adminPasswordRequired: 'Senha de Admin *',
      adminPasswordPlaceholder: 'Informe a senha de admin (também usada como senha RCON)',
      showPassword: 'Mostrar Senha',
      hidePassword: 'Ocultar Senha',
      saveChanges: 'Salvar Alterações',
      createServer: 'Criar Servidor',
      saving: 'Salvando...',
      preparing: 'Preparando...',
      loadingServerInfo: 'Carregando informações do servidor...',
      closeConfirm: 'Tem certeza que deseja fechar? Dados não salvos serão perdidos.',
      range: 'Faixa',
      // Map options
      maps: {
        TheIsland: 'The Island',
        TheCenter: 'The Center',
        ScorchedEarth_P: 'Scorched Earth',
        Aberration_P: 'Aberration',
        Extinction: 'Extinction',
        Valguero_P: 'Valguero',
        Genesis: 'Genesis',
        Genesis2: 'Genesis 2',
        CrystalIsles: 'Crystal Isles',
        LostIsland: 'Lost Island',
        Fjordur: 'Fjordur'
      },
      selectMapPlaceholder: 'Select map or enter map name',
      customMapPlaceholder: 'Enter custom map name',
      customMap: 'Custom Map',
      noMatchingMaps: 'No matching maps',
      mapId: 'Map ID',
      officialMaps: 'Official Maps',
      searchMaps: 'Search maps...',
      serverCreateError: 'Failed to create server. Please check your input and try again.',
      exportAll: 'Export All Config',
      importAll: 'Import All Config',
      exportFile: 'Download',
      importFile: 'Import'
    },
    // Editor related
    editor: {
      visualEdit: 'Visual Edit',
      textEdit: 'Text Edit',
      visualEditMode: 'Visual Edit',
      textEditMode: 'Text Edit',
      syncing: 'Syncing...',
      synced: 'Synced',
      pendingSync: 'Pending Sync',
      resetToDefault: 'Reset to Default',
      format: 'Format',
      content: 'Content',
      placeholder: 'Enter configuration content...',
      placeholders: {
        gameIni: 'Insira o conteúdo da configuração do Game.ini...',
        gameUserSettings: 'Insira o conteúdo da configuração do GameUserSettings.ini...',
      },
      description: 'This file contains basic server settings such as port, password, max players, etc.',
      visualEditModeDesc: 'Visual Edit Mode',
      visualEditModeTip: 'Modify parameters through form controls. Hover over the icon next to parameter names to view detailed descriptions.',
      gameIniTextEditDesc: 'Directly edit Game.ini configuration file content.',
      showPassword: 'Show Password',
      hidePassword: 'Hide Password',
      enabled: 'Enabled',
      disabled: 'Disabled',
      parametersCount: ' parameters',
      otherSettingsTitle: 'Outras configurações neste arquivo',
      otherSettingsHint: '{count} configurações que o ARK gravou e não estão nas predefinições acima. São mostradas para nada ficar oculto e são preservadas ao salvar.',
      defaultValue: 'Default',
      parseGameIniError: 'Failed to parse Game.ini text',
      syncVisualToTextError: 'Failed to sync visual config to text',
      range: 'Range',
      syncTip: {
        visual: 'Modify parameters in visual mode, configuration will be synced to text when switching to text mode.',
        text: 'Edit configuration file directly in text mode, content will be parsed when switching to visual mode.'
      }
    },
    // Server args editor related
    argsEditor: {
      title: 'Startup Arguments Configuration',
      switchParams: 'Switch Parameters',
      numberParams: 'Number Parameters',
      textParams: 'Text Parameters',
      selectParams: 'Select Parameters',
      range: 'Range',
      pleaseSelect: 'Please Select',
      customArgs: 'Custom Arguments',
      customArgsDesc: 'Add custom startup arguments that will be directly added to the startup command',
      addCustomArg: 'Add Custom Argument',
      removeCustomArg: 'Remove Custom Argument',
      customArgPlaceholder: 'Enter custom startup argument, e.g. -nosteam',
      enabled: 'Enabled',
      disabled: 'Disabled'
    },
    // Startup parameter categories
    paramCategories: {
      basic: 'Basic',
      core: 'Core',
      dinos: 'Dinosaurs',
      structures: 'Structures',
      pvp: 'PvP',
      mechanics: 'Game Mechanics',
      transfer: 'Transfer & Cluster',
      performance: 'Performance',
      graphics: 'Graphics',
      security: 'Security',
      logging: 'Logging',
      mods: 'Mods',
      features: 'Features',
      maintenance: 'Maintenance',
      advanced: 'Advanced',
      custom: 'Custom'
    },
    // Query parameters translation
    queryParams: {
      AltSaveDirectoryName: 'Alternative Save Directory Name',
      EventColorsChanceOverride: 'Event Colors Chance Override',
      GameModIds: 'Game Mod IDs',
      NewYear1UTC: 'New Year Event Start Time (UTC)',
      NewYear2UTC: 'New Year Event End Time (UTC)'
    },
    // Command line parameters translation
    commandLineArgs: {
      // Events and Features
      ActiveEvent: 'Active Event',
      NewYearEvent: 'New Year Event',
      UseVivox: 'Use Vivox Voice Chat',
      webalarm: 'Web Alarms',
      AllowChatSpam: 'Allow Chat Spam',
      
      // Mods and Steam
      automanagedmods: 'Auto-Managed Mods',
      MapModID: 'Map Mod ID',
      
      // Crossplay and Network
      crossplay: 'Enable Crossplay',
      epiconly: 'Epic Games Store Only',
      PublicIPForEpic: 'Public IP for Epic Games',
      MULTIHOME: 'Multi-Home IP Address',
      
      // Server Management
      culture: 'Server Language Culture',
      exclusivejoin: 'Exclusive Join (Whitelist)',
      EnableIdlePlayerKick: 'Enable Idle Player Kick',
      MaxNumOfSaveBackups: 'Maximum Number of Save Backups',
      newsaveformat: 'New Save Format',
      NoHangDetection: 'Disable Hang Detection',
      
      // Creatures and Gameplay
      DisableCustomFoldersInTributeInventories: 'Disable Custom Folders in Tribute Inventories',
      ForceAllowCaveFlyers: 'Force Allow Cave Flyers',
      ForceRespawnDinos: 'Force Respawn Dinos',
      NoDinos: 'No Dinosaurs',
      imprintlimit: 'Imprint Limit Percentage',
      MinimumTimeBetweenInventoryRetrieval: 'Minimum Time Between Inventory Retrieval',
      
      // PvP Settings
      DisableRailgunPVP: 'Disable Railgun in PvP',
      pvedisallowtribewar: 'PvE Disallow Tribe War',
      pveallowtribewar: 'PvE Allow Tribe War',
      
      // Security and Anti-Cheat
      insecure: 'Disable VAC (Insecure)',
      NoBattlEye: 'Disable BattlEye',
      noantispeedhack: 'Disable Anti-Speedhack',
      speedhackbias: 'Speedhack Detection Bias',
      noundermeshchecking: 'Disable Under-Mesh Checking',
      noundermeshkilling: 'Disable Under-Mesh Killing',
      SecureSendArKPayload: 'Secure Send ARK Payload',
      UseItemDupeCheck: 'Use Item Duplication Check',
      UseSecureSpawnRules: 'Use Secure Spawn Rules',
      BattlEyeServerRecheck: 'BattlEye Server Recheck',
      
      // Performance Optimization
      nocombineclientmoves: 'Disable Combine Client Moves',
      StasisKeepControllers: 'Stasis Keep Controllers',
      structurememopts: 'Structure Memory Optimizations',
      UseStructureStasisGrid: 'Use Structure Stasis Grid',
      DormancyNetMultiplier: 'Dormancy Network Multiplier',
      nodormancythrottling: 'Disable Dormancy Throttling',
      nitradotest2: 'Nitrado Test Mode 2',
      dedihibernation: 'Dedicated Hibernation',
      
      // Graphics and Client
      ServerAllowAnsel: 'Server Allow NVIDIA Ansel',
      
      // Logging and Admin
      servergamelog: 'Server Game Log',
      servergamelogincludetribelogs: 'Server Game Log Include Tribe Logs',
      ServerRCONOutputTribeLogs: 'Server RCON Output Tribe Logs',
      NotifyAdminCommandsInChat: 'Notify Admin Commands in Chat',
      
      // Transfer and Cluster
      ClusterDirOverride: 'Cluster Directory Override',
      clusterid: 'Cluster ID',
      NoTransferFromFiltering: 'No Transfer From Filtering',
      usestore: 'Use Store',
      BackupTransferPlayerDatas: 'Backup Transfer Player Data',
      converttostore: 'Convert to Store',
      
      // Advanced/Undocumented
      CustomAdminCommandTrackingURL: 'Custom Admin Command Tracking URL',
      CustomMerticsURL: 'Custom Metrics URL',
      CustomNotificationURL: 'Custom Notification URL',
      DisableDupeLogDeletes: 'Disable Duplicate Log Deletes',
      EnableOfficialOnlyVersioningCode: 'Enable Official Only Versioning Code',
      EnableVictoryCoreDupeCheck: 'Enable Victory Core Duplication Check',
      forcedisablemeshchecking: 'Force Disable Mesh Checking',
      ForceDupeLog: 'Force Duplication Log',
      ignoredupeditems: 'Ignore Duplicated Items',
      MaxConnectionsPerIP: 'Maximum Connections Per IP',
      parseservertojson: 'Parse Server to JSON',
      pauseonddos: 'Pause on DDoS',
      PreventTotalConversionSaveDir: 'Prevent Total Conversion Save Directory',
      ReloadedForBackup: 'Reloaded for Backup',
      UnstasisDinoObstructionCheck: 'Unstasis Dino Obstruction Check',
      UseTameEffectivenessClamp: 'Use Tame Effectiveness Clamp',
      UseServerNetSpeedCheck: 'Use Server Network Speed Check'
    },

    // Game.ini parameter categories
    gameIniCategories: {
      gameBasic: 'Basic Game',
      experienceSettings: 'Experience and Level',
      breedingSettings: 'Breeding',
      itemSettings: 'Item and Resource',
      dinoSettings: 'Dinosaur',
      tribeSettings: 'Tribe and Player',
      pvpSettings: 'PvP',
      structureSettings: 'Building and Structure',
      advancedSettings: 'Advanced',
      customSettings: 'Custom Configuration'
    },

    // GameUserSettings.ini parameter categories
    gameUserSettingsCategories: {
      serverBasic: 'Server Basic',
      gameMode: 'Game Mode',
      communication: 'Chat and Communication',
      gameMultipliers: 'Game Multiplier',
      characterSettings: 'Character',
      dinoSettings: 'Dinosaur',
      environmentSettings: 'Environment',
      structureSettings: 'Structure',
      tribeSettings: 'Tribe and Alliance',
      breedingSettings: 'Breeding and Imprinting',
      itemSettings: 'Item and Supply',
      performanceSettings: 'Server Performance',
      diseaseSettings: 'Disease and Status',
      offlineRaidSettings: 'Offline Raid Protection',
      crossArkSettings: 'Cross-ARK Transfer',
      flyerSettings: 'Flyer',
      advancedSettings: 'Advanced Feature'
    },

    // Game.ini parameter translations
    gameIniParams: {
      // Basic settings
      bUseSingleplayerSettings: 'Use Singleplayer Settings',
      bDisableStructurePlacementCollision: 'Disable Structure Placement Collision',
      bAllowFlyerCarryPvE: 'Allow Flyer Carry PvE',
      bDisableStructureDecayPvE: 'Disable Structure Decay PvE',
      bAllowUnlimitedRespecs: 'Allow Unlimited Respecs',
      bAllowPlatformSaddleMultiFloors: 'Allow Platform Saddle Multi Floors',
      bPassiveDefensesDamageRiderlessDinos: 'Passive Defenses Damage Riderless Dinos',
      bPvEDisableFriendlyFire: 'PvE Disable Friendly Fire',
      bDisableFriendlyFire: 'Disable Friendly Fire',
      bEnablePvPGamma: 'Enable PvP Gamma',
      DifficultyOffset: 'Difficulty Offset',
      OverrideOfficialDifficulty: 'Override Official Difficulty',

      // Experience and level settings
      XPMultiplier: 'XP Multiplier',
      PlayerCharacterWaterDrainMultiplier: 'Player Water Drain Multiplier',
      PlayerCharacterFoodDrainMultiplier: 'Player Food Drain Multiplier',
      PlayerCharacterStaminaDrainMultiplier: 'Player Stamina Drain Multiplier',
      PlayerCharacterHealthRecoveryMultiplier: 'Player Health Recovery Multiplier',

      // Breeding settings
      MatingIntervalMultiplier: 'Mating Interval Multiplier',
      EggHatchSpeedMultiplier: 'Egg Hatch Speed Multiplier',
      BabyMatureSpeedMultiplier: 'Baby Mature Speed Multiplier',
      BabyFoodConsumptionSpeedMultiplier: 'Baby Food Consumption Speed Multiplier',
      BabyCuddleIntervalMultiplier: 'Baby Cuddle Interval Multiplier',
      BabyCuddleGracePeriodMultiplier: 'Baby Cuddle Grace Period Multiplier',
      BabyCuddleLoseImprintQualitySpeedMultiplier: 'Baby Cuddle Lose Imprint Quality Speed Multiplier',

      // Item and resource settings
      HarvestAmountMultiplier: 'Harvest Amount Multiplier',
      HarvestHealthMultiplier: 'Harvest Health Multiplier',
      ResourcesRespawnPeriodMultiplier: 'Resources Respawn Period Multiplier',
      ItemStackSizeMultiplier: 'Item Stack Size Multiplier',
      CropGrowthSpeedMultiplier: 'Crop Growth Speed Multiplier',
      GlobalItemDecompositionTimeMultiplier: 'Global Item Decomposition Time Multiplier',
      GlobalCorpseDecompositionTimeMultiplier: 'Global Corpse Decomposition Time Multiplier',

      // Dinosaur settings
      TamingSpeedMultiplier: 'Taming Speed Multiplier',
      DinoCharacterFoodDrainMultiplier: 'Dino Food Drain Multiplier',
      DinoCharacterStaminaDrainMultiplier: 'Dino Stamina Drain Multiplier',
      DinoCharacterHealthRecoveryMultiplier: 'Dino Health Recovery Multiplier',
      DinoCountMultiplier: 'Dino Count Multiplier',
      WildDinoCharacterFoodDrainMultiplier: 'Wild Dino Food Drain Multiplier',
      WildDinoTorporDrainMultiplier: 'Wild Dino Torpor Drain Multiplier',

      // Tribe and player settings
      MaxNumberOfPlayersInTribe: 'Max Number Of Players In Tribe',
      TribeNameChangeCooldown: 'Tribe Name Change Cooldown (Minutes)',
      bPvEAllowTribeWar: 'PvE Allow Tribe War',
      bPvEAllowTribeWarCancel: 'PvE Allow Tribe War Cancel',

      // PvP settings
      bIncreasePvPRespawnInterval: 'Increase PvP Respawn Interval',
      IncreasePvPRespawnIntervalCheckPeriod: 'PvP Respawn Interval Check Period (Seconds)',
      IncreasePvPRespawnIntervalMultiplier: 'PvP Respawn Interval Multiplier',
      IncreasePvPRespawnIntervalBaseAmount: 'PvP Respawn Interval Base Amount (Seconds)',

      // Structure and building settings
      StructureDamageMultiplier: 'Structure Damage Multiplier',
      StructureResistanceMultiplier: 'Structure Resistance Multiplier',
      StructureDamageRepairCooldown: 'Structure Damage Repair Cooldown (Seconds)',
      PvEStructureDecayPeriodMultiplier: 'PvE Structure Decay Period Multiplier',
      MaxStructuresInRange: 'Max Structures In Range',

      // Advanced feature settings
      bAutoPvETimer: 'Auto PvE Timer',
      bAutoPvEUseSystemTime: 'Auto PvE Use System Time',
      AutoPvEStartTimeSeconds: 'Auto PvE Start Time (Seconds)',
      AutoPvEStopTimeSeconds: 'Auto PvE Stop Time (Seconds)',
      bOnlyAllowSpecifiedEngrams: 'Only Allow Specified Engrams',
      bAutoUnlockAllEngrams: 'Auto Unlock All Engrams',
      bShowCreativeMode: 'Show Creative Mode',
      bUseCorpseLocator: 'Use Corpse Locator',
      bDisableLootCrates: 'Disable Loot Crates',
      bDisableDinoRiding: 'Disable Dino Riding',
      bDisableDinoTaming: 'Disable Dino Taming',
      bAllowCustomRecipes: 'Allow Custom Recipes',

      // Custom configuration
      DayCycleSpeedScale: 'Day Cycle Speed Scale',
      NightTimeSpeedScale: 'Night Time Speed Scale',
      DayTimeSpeedScale: 'Day Time Speed Scale'
    },

    // GameUserSettings.ini parameter translations
    gameUserSettingsParams: {
      // Server basic settings
      ServerPassword: 'Server Password',
      SpectatorPassword: 'Spectator Password',
      AdminLogging: 'Admin Logging',

      // Game mode settings
      serverPVE: 'PvE Mode',
      serverHardcore: 'Hardcore Mode',
      ShowMapPlayerLocation: 'Show Player Location',
      allowThirdPersonPlayer: 'Allow Third Person',
      ServerCrosshair: 'Show Crosshair',
      EnablePvPGamma: 'PvP Gamma Adjustment',
      DisablePvEGamma: 'Disable PvE Gamma Adjustment',
      serverForceNoHud: 'Force Hide HUD',
      ShowFloatingDamageText: 'Show Floating Damage Text',
      AllowHitMarkers: 'Allow Hit Markers',

      // Chat and communication settings
      globalVoiceChat: 'Global Voice Chat',
      proximityChat: 'Proximity Chat',
      alwaysNotifyPlayerJoined: 'Always Notify Player Joined',
      alwaysNotifyPlayerLeft: 'Always Notify Player Left',
      DontAlwaysNotifyPlayerJoined: 'Disable Player Join Notification',

      // Game multiplier settings
      XPMultiplier: 'XP Multiplier',
      TamingSpeedMultiplier: 'Taming Speed Multiplier',
      HarvestAmountMultiplier: 'Harvest Amount Multiplier',
      HarvestHealthMultiplier: 'Harvest Health Multiplier',
      ResourcesRespawnPeriodMultiplier: 'Resources Respawn Period Multiplier',
      ItemStackSizeMultiplier: 'Item Stack Size Multiplier',

      // Character settings
      PlayerCharacterHealthRecoveryMultiplier: 'Player Health Recovery Multiplier',
      PlayerCharacterFoodDrainMultiplier: 'Player Food Drain Multiplier',
      PlayerCharacterWaterDrainMultiplier: 'Player Water Drain Multiplier',
      PlayerCharacterStaminaDrainMultiplier: 'Player Stamina Drain Multiplier',
      PlayerDamageMultiplier: 'Player Damage Multiplier',
      PlayerResistanceMultiplier: 'Player Resistance Multiplier',
      OxygenSwimSpeedStatMultiplier: 'Oxygen Swim Speed Stat Multiplier',
      ImplantSuicideCD: 'Implant Suicide Cooldown',

      // Dinosaur settings
      DinoCountMultiplier: 'Dino Count Multiplier',
      DinoCharacterHealthRecoveryMultiplier: 'Dino Health Recovery Multiplier',
      DinoCharacterFoodDrainMultiplier: 'Dino Food Drain Multiplier',
      DinoCharacterStaminaDrainMultiplier: 'Dino Stamina Drain Multiplier',
      DinoDamageMultiplier: 'Dino Damage Multiplier',
      TamedDinoDamageMultiplier: 'Tamed Dino Damage Multiplier',
      DinoResistanceMultiplier: 'Dino Resistance Multiplier',
      TamedDinoResistanceMultiplier: 'Tamed Dino Resistance Multiplier',
      MaxTamedDinos: 'Max Tamed Dinos',
      MaxPersonalTamedDinos: 'Max Personal Tamed Dinos',
      DisableDinoDecayPvE: 'Disable Dino Decay PvE',
      AutoDestroyDecayedDinos: 'Auto Destroy Decayed Dinos',
      PvEDinoDecayPeriodMultiplier: 'PvE Dino Decay Period Multiplier',
      PvPDinoDecay: 'PvP Dino Decay',
      AllowRaidDinoFeeding: 'Allow Raid Dino Feeding',
      RaidDinoCharacterFoodDrainMultiplier: 'Raid Dino Food Drain Multiplier',
      AllowFlyerCarryPvE: 'PvE Allow Flyer Carry',
      bForceCanRideFliers: 'Force Can Ride Fliers',

      // Environment settings
      DayCycleSpeedScale: 'Day Cycle Speed Scale',
      DayTimeSpeedScale: 'Day Time Speed Scale',
      NightTimeSpeedScale: 'Night Time Speed Scale',
      DisableWeatherFog: 'Disable Weather Fog',
      DifficultyOffset: 'Difficulty Offset',
      OverrideOfficialDifficulty: 'Override Official Difficulty',
      RandomSupplyCratePoints: 'Random Supply Crate Points',

      // Structure settings
      StructureDamageMultiplier: 'Structure Damage Multiplier',
      StructureResistanceMultiplier: 'Structure Resistance Multiplier',
      TheMaxStructuresInRange: 'Max Structures In Range',
      NewMaxStructuresInRange: 'New Max Structures In Range',
      MaxStructuresInRange: 'Max Structures In Range',
      DisableStructureDecayPvE: 'Disable Structure Decay PvE',
      PvEStructureDecayPeriodMultiplier: 'PvE Structure Decay Period Multiplier',
      PvEStructureDecayDestructionPeriod: 'PvE Structure Decay Destruction Period',
      PvPStructureDecay: 'PvP Structure Decay',
      StructurePickupTimeAfterPlacement: 'Structure Pickup Time After Placement',
      StructurePickupHoldDuration: 'Structure Pickup Hold Duration',
      AlwaysAllowStructurePickup: 'Always Allow Structure Pickup',
      OnlyAutoDestroyCoreStructures: 'Only Auto Destroy Core Structures',
      OnlyDecayUnsnappedCoreStructures: 'Only Decay Unsnapped Core Structures',
      FastDecayUnsnappedCoreStructures: 'Fast Decay Unsnapped Core Structures',
      DestroyUnconnectedWaterPipes: 'Destroy Unconnected Water Pipes',
      StructurePreventResourceRadiusMultiplier: 'Structure Prevent Resource Radius Multiplier',
      MaxPlatformSaddleStructureLimit: 'Max Platform Saddle Structure Limit',
      PerPlatformMaxStructuresMultiplier: 'Per Platform Max Structures Multiplier',
      PlatformSaddleBuildAreaBoundsMultiplier: 'Platform Saddle Build Area Bounds Multiplier',
      OverrideStructurePlatformPrevention: 'Override Structure Platform Prevention',
      EnableExtraStructurePreventionVolumes: 'Enable Extra Structure Prevention Volumes',
      AllowCaveBuildingPvE: 'Allow Cave Building PvE',
      AllowCaveBuildingPvP: 'Allow Cave Building PvP',
      PvEAllowStructuresAtSupplyDrops: 'PvE Allow Structures At Supply Drops',
      AllowCrateSpawnsOnTopOfStructures: 'Allow Crate Spawns On Top Of Structures',
      bAllowPlatformSaddleMultiFloors: 'Allow Platform Saddle Multi Floors',
      MaxGateFrameOnSaddles: 'Max Gate Frame On Saddles',

      // Tribe and alliance settings
      MaxNumberOfPlayersInTribe: 'Max Number Of Players In Tribe',
      TribeNameChangeCooldown: 'Tribe Name Change Cooldown',
      PreventTribeAlliances: 'Prevent Tribe Alliances',
      MaxAlliancesPerTribe: 'Max Alliances Per Tribe',
      MaxTribesPerAlliance: 'Max Tribes Per Alliance',

      // Breeding and imprinting settings
      AllowAnyoneBabyImprintCuddle: 'Allow Anyone Baby Imprint Cuddle',
      DisableImprintDinoBuff: 'Disable Imprint Dino Buff',
      BabyImprintingStatScaleMultiplier: 'Baby Imprinting Stat Scale Multiplier',

      // Item and supply settings
      ClampItemSpoilingTimes: 'Clamp Item Spoiling Times',
      ClampResourceHarvestDamage: 'Clamp Resource Harvest Damage',
      UseOptimizedHarvestingHealth: 'Use Optimized Harvesting Health',
      BanListURL: 'Ban List URL',

      // Server performance settings
      AutoSavePeriodMinutes: 'Auto Save Period Minutes',
      KickIdlePlayersPeriod: 'Kick Idle Players Period',
      ListenServerTetherDistanceMultiplier: 'Listen Server Tether Distance Multiplier',
      RCONServerGameLogBuffer: 'RCON Server Game Log Buffer',
      NPCNetworkStasisRangeScalePlayerCountStart: 'NPC Network Stasis Range Scale Player Count Start',
      NPCNetworkStasisRangeScalePlayerCountEnd: 'NPC Network Stasis Range Scale Player Count End',
      NPCNetworkStasisRangeScalePercentEnd: 'NPC Network Stasis Range Scale Percent End',

      // Disease and status settings
      PreventDiseases: 'Prevent Diseases',
      NonPermanentDiseases: 'Non Permanent Diseases',
      PreventSpawnAnimations: 'Prevent Spawn Animations',

      // Offline raid protection settings
      PreventOfflinePvP: 'Prevent Offline PvP',
      PreventOfflinePvPInterval: 'Prevent Offline PvP Interval',

      // Cross-ARK transfer settings
      NoTributeDownloads: 'No Tribute Downloads',
      PreventDownloadSurvivors: 'Prevent Download Survivors',
      PreventDownloadItems: 'Prevent Download Items',
      PreventDownloadDinos: 'Prevent Download Dinos',
      PreventUploadSurvivors: 'Prevent Upload Survivors',
      PreventUploadItems: 'Prevent Upload Items',
      PreventUploadDinos: 'Prevent Upload Dinos',
      CrossARKAllowForeignDinoDownloads: 'Cross ARK Allow Foreign Dino Downloads',
      MaxTributeDinos: 'Max Tribute Dinos',
      MaxTributeItems: 'Max Tribute Items',
      MinimumDinoReuploadInterval: 'Minimum Dino Reupload Interval',
      TributeItemExpirationSeconds: 'Tribute Item Expiration Seconds',
      TributeDinoExpirationSeconds: 'Tribute Dino Expiration Seconds',
      TributeCharacterExpirationSeconds: 'Tribute Character Expiration Seconds',

      // Flyer settings
      AllowFlyingStaminaRecovery: 'Allow Flying Stamina Recovery',
      ForceFlyerExplosives: 'Force Flyer Explosives',

      // Advanced feature settings
      AllowMultipleAttachedC4: 'Allow Multiple Attached C4',
      AllowIntegratedSPlusStructures: 'Allow Integrated S+ Structures',
      AllowHideDamageSourceFromLogs: 'Allow Hide Damage Source From Logs',
      AllowSharedConnections: 'Allow Shared Connections',
      bFilterTribeNames: 'Filter Tribe Names',
      bFilterCharacterNames: 'Filter Character Names',
      bFilterChat: 'Filter Chat',
      EnableCryoSicknessPVE: 'Enable Cryo Sickness PVE',
      EnableCryopodNerf: 'Enable Cryopod Nerf',
      CryopodNerfDuration: 'Cryopod Nerf Duration',
      CryopodNerfDamageMult: 'Cryopod Nerf Damage Multiplier',
      CryopodNerfIncomingDamageMultPercent: 'Cryopod Nerf Incoming Damage Multiplier Percent',
      DisableCryopodEnemyCheck: 'Disable Cryopod Enemy Check',
      DisableCryopodFridgeRequirement: 'Disable Cryopod Fridge Requirement',
      AllowCryoFridgeOnSaddle: 'Allow Cryo Fridge On Saddle',
      MaxTrainCars: 'Máximo de Vagões de Trem',
      MaxHexagonsPerCharacter: 'Max Hexagons Per Character',
      AllowTekSuitPowersInGenesis: 'Allow Tek Suit Powers In Genesis',
      CustomDynamicConfigUrl: 'Custom Dynamic Config URL'
    },

    // GameUserSettings.ini editor description
    gameUserSettingsTextEditDesc: 'Edit GameUserSettings.ini configuration file content directly. Changes will be automatically parsed and synchronized to the visual interface. Switch to visual mode to see the parsed parameter settings.',

    // Default values for GameUserSettings parameters
    defaultValues: {
      sessionName: 'My ARK Server',
      message: 'Welcome to ARK Server!'
    },

    // Placeholders
    placeholders: {
      gameUserSettings: `[ServerSettings]
SessionName=My ARK Server
ServerPassword=
MaxPlayers=70

[SessionSettings]
SessionName=My ARK Server

[MessageOfTheDay]
Message=Welcome to ARK Server!

[/Script/Engine.GameSession]
MaxPlayers=70`,
      gameIni: `[/Script/ShooterGame.ShooterGameMode]
DifficultyOffset=0.2
OverrideOfficialDifficulty=5.0
XPMultiplier=1.0
TamingSpeedMultiplier=1.0
HarvestAmountMultiplier=1.0
ResourcesRespawnPeriodMultiplier=1.0
PlayerCharacterWaterDrainMultiplier=1.0
PlayerCharacterFoodDrainMultiplier=1.0
DinoCharacterFoodDrainMultiplier=1.0
PlayerCharacterStaminaDrainMultiplier=1.0
DinoCharacterStaminaDrainMultiplier=1.0
PlayerCharacterHealthRecoveryMultiplier=1.0
DinoCharacterHealthRecoveryMultiplier=1.0
DinoCountMultiplier=1.0
AllowFlyerCarryPvE=False
MaxTamedDinos=4000
StructureDamageMultiplier=1.0
StructureResistanceMultiplier=1.0
TheMaxStructuresInRange=10500
BabyMatureSpeedMultiplier=1.0
EggHatchSpeedMultiplier=1.0
BabyCuddleIntervalMultiplier=1.0
BabyCuddleGracePeriodMultiplier=1.0
BabyImprintAmountMultiplier=1.0`
    },

    // RCON console panel
    rcon: {
      console: 'Console RCON',
      connected: 'Conectado',
      disconnected: 'Desconectado',
      connecting: 'Conectando...',
      serverOffline: 'O servidor parece estar offline — RCON indisponível',
      expand: 'Expandir console',
      collapse: 'Recolher console',
      connectingHint: 'Abra uma sessão de terminal para executar comandos admin remotos neste servidor ARK.',
      commandHint: 'Digite um comando e pressione Enter para enviar. O servidor deve estar em execução com RCON habilitado.',
      connectionError: 'Erro de conexão — tentando novamente...',
      invalidResponse: 'Resposta inválida recebida do servidor',
      sendFailed: 'Falha ao enviar comando',
      closePanel: 'Fechar terminal',
      outputTruncated: '\n... saída truncada ...'
    }
  },

  // Modals
  modals: {
    privacyPolicy: 'Privacy Policy',
    termsOfService: 'Terms of Service',
    privacyPolicyContent: 'We value your privacy. This application only collects necessary server management information and will not leak your personal data.',
    termsOfServiceContent: 'Using this Ark Server Commander means you agree to comply with the relevant terms of use. Please use this tool reasonably for server management.'
  },

  // Footer
  footer: {
    copyright: '© {year} Ark Server Commander developed by {company}',
    privacyPolicy: 'Privacy Policy',
    termsOfService: 'Terms of Service',
    support: 'Technical Support',
    github: 'GitHub'
  },

  // Error messages
  errors: {
    networkError: 'Erro de conexão de rede',
    serverError: 'Erro do servidor',
    unauthorized: 'Acesso não autorizado',
    forbidden: 'Acesso negado',
    notFound: 'Página não encontrada',
    validationError: 'Falha na validação de entrada',
    unknownError: 'Erro desconhecido',
    tryAgain: 'Por favor, tente novamente',
    contactSupport: 'Por favor, contate o suporte técnico'
  },

  // Success messages
  success: {
    operationSuccess: 'Operação bem-sucedida',
    dataSaved: 'Dados salvos',
    dataDeleted: 'Dados excluídos',
    dataUpdated: 'Dados atualizados',
    dataCreated: 'Dados criados'
  },

  // Form
  form: {
    required: 'Este campo é obrigatório',
    invalidFormat: 'Formato inválido',
    minLength: 'Mínimo de {min} caracteres',
    maxLength: 'Máximo de {max} caracteres',
    invalidEmail: 'E-mail inválido',
    invalidUrl: 'URL inválida',
    invalidNumber: 'Por favor, informe um número válido',
    invalidPort: 'A porta deve estar entre 1 e 65535',
    invalidPath: 'Por favor, informe um caminho válido',
    passwordMismatch: 'As senhas não coincidem',
    usernameExists: 'Nome de usuário já existe',
    serverNameExists: 'Nome de servidor já existe',
    portInUse: 'Porta já está em uso'
  },

  // Plugins
  plugins: {
    title: 'Gerenciador de Plugins',
    selectServer: 'Selecione um servidor...',
    selectServerHint: 'Selecione um servidor acima para gerenciar os plugins',
    root: 'Raiz',
    upload: 'Enviado',
    newFolder: 'Nova Pasta',
    folderNamePlaceholder: 'Nome da pasta',
    rename: 'Renomear',
    delete: 'Excluir',
    download: 'Baixar',
    refresh: 'Atualizar',
    uploading: 'Enviando...',
    empty: 'Esta pasta está vazia',
    dragDropHint: 'Arraste e solte arquivos aqui para enviar',
    confirmDelete: 'Excluir "{name}"?',
    editing: 'Editando',
    edit: 'Editar',
    saveSuccess: 'Arquivo salvo com sucesso',
    extract: 'Extrair',
    downloadZip: 'Baixar como ZIP'
  },

  // Jogadores
  players: {
    title: 'Gerenciamento de Jogadores',
    description: 'Visualize e gerencie os jogadores online dos seus servidores ARK',
    selectServer: 'Selecionar Servidor',
    serverIdLabel: 'ID do Servidor',
    serverIdPlaceholder: 'Informe o ID do servidor (ex. 1)',
    fetchPlayers: 'Buscar Jogadores',
    noOnlinePlayers: 'Nenhum jogador online encontrado',
    serverNotRunning: 'Este servidor está parado, então não há jogadores para mostrar.',
    serverStarting: 'Este servidor está iniciando \u2014 os dados de jogadores aparecerão quando estiver pronto.',
    onlineNow: 'Online Agora',
    onlinePlayers: 'Jogadores Online',
    maxPlayers: 'Máx. Jogadores',
    server: 'Servidor',
    session: 'Sessão',
    playerName: 'Nome do Jogador',
    status: 'Status',
    online: 'Online',
    playerHistory: 'Histórico de Jogadores',
    noHistory: 'Nenhum histórico disponível',
    fetched: 'Jogadores buscados com sucesso',
    fetchFailed: 'Falha ao buscar jogadores',
    joinedAt: 'Entrou Em'
  },

  // Logs de auditoria
  rcon: {
    quickActions: 'Ações rápidas',
    actions: 'Ações',
    broadcastPlaceholder: 'Transmitir uma mensagem para todos os jogadores\u2026',
    saveWorld: 'Salvar mundo',
    setTime: 'Definir hora',
    destroyWildDinos: 'Remover dinos selvagens',
    confirmDestroyDinos: 'Isso remove todas as criaturas selvagens. Tem certeza?',
    confirm: 'Confirmar',
    cancel: 'Cancelar',
    kick: 'Expulsar jogador',
    ban: 'Banir jogador',
    actionFailed: 'Falha ao executar o comando',
  },
  backups: {
    scheduleTitle: 'Backups automáticos',
    scheduleHint: 'Execute backups periodicamente e, se quiser, envie-os para fora do host.',
    on: 'Ativo',
    enableAutomatic: 'Ativar backups automáticos',
    everyHours: 'A cada (horas)',
    keepLast: 'Manter os últimos',
    keepZeroHint: '0 mantém todos os backups.',
    uploadToCloud: 'Enviar cada backup para o armazenamento de objetos',
    cloudNotConfigured: 'O armazenamento de objetos não está configurado no servidor.',
    cloudTarget: 'Enviando para {destination} via {provider}',
    lastRun: 'Última execução {time} \u2014 {status}',
    nextRun: 'Próxima execução {time}',
    saveSchedule: 'Salvar agendamento',
    saved: 'Salvo',
    loadFailed: 'Não foi possível carregar o agendamento',
    saveFailed: 'Não foi possível salvar o agendamento',
    uploadNow: 'Enviar para a nuvem',
    uploadFailed: 'Falha no envio',
    uploaded: 'Enviado',
  },
  mods: {
    title: 'Gerenciamento de mods',
    subtitle: 'Explore a Oficina Steam do ARK e gerencie os mods de cada servidor.',
    server: 'Servidor',
    installed: 'Mods instalados ({count})',
    loadOrderHint: 'A ordem de carregamento importa \u2014 mods posteriores substituem os anteriores.',
    restartHint: 'Reinicie o servidor para aplicar as altera\u00e7\u00f5es de mods.',
    noMods: 'Nenhum mod neste servidor.',
    browse: 'Oficina Steam',
    searchPlaceholder: 'Pesquisar na Oficina\u2026',
    workshopIdPlaceholder: 'Adicionar por ID da Oficina',
    addByIdHint: 'Voc\u00ea ainda pode adicionar mods pelo ID da Oficina.',
    add: 'Adicionar',
    added: 'Adicionado',
    remove: 'Remover',
    enable: 'Ativar',
    disable: 'Desativar',
    disabled: 'Desativado',
    moveUp: 'Mover para cima',
    moveDown: 'Mover para baixo',
    viewOnSteam: 'Ver na Steam',
    loadFailed: 'N\u00e3o foi poss\u00edvel carregar os mods',
    searchFailed: 'A pesquisa na Oficina falhou',
    addFailed: 'N\u00e3o foi poss\u00edvel adicionar o mod',
    removeFailed: 'N\u00e3o foi poss\u00edvel remover o mod',
    updateFailed: 'N\u00e3o foi poss\u00edvel atualizar o mod',
    reorderFailed: 'N\u00e3o foi poss\u00edvel salvar a ordem',
  },
  metrics: {
    title: 'Métricas do servidor',
    subtitle: 'CPU, memória e jogadores online dos seus servidores em tempo real.',
    refresh: 'Atualizar agora',
    updatedAt: 'Atualizado às {time}',
    loadFailed: 'Não foi possível carregar as métricas',
    playersOnline: 'Jogadores online',
    serversRunning: 'Servidores ativos',
    averageCpu: 'CPU média',
    totalMemory: 'Memória total',
    noServers: 'Nenhum servidor ainda.',
    partialPlayerCount: 'Alguns servidores não informaram o número de jogadores.',
    cpu: 'CPU',
    memory: 'Memória',
    players: 'Jogadores',
    unknown: 'Desconhecido',
    notRunning: 'Parado',
    acrossCores: 'em {count} núcleos',
    severity: { good: 'Normal', warning: 'Alto', critical: 'Crítico' },
    status: { running: 'Ativo', stopped: 'Parado' },
    errors: {
      'docker unavailable': 'Docker indisponível.',
      'container not found': 'Contêiner não encontrado.',
      'stats unavailable': 'Estatísticas de recursos indisponíveis.',
    },
  },
  auditLogs: {
    title: 'Logs de Auditoria',
    page: 'Página',
    noLogs: 'Nenhum registro de auditoria encontrado',
    fetchFailed: 'Falha ao carregar os logs de auditoria',
    filterUserID: 'ID do Usuário',
    filterUserIDPlaceholder: 'Filtrar por ID do usuário',
    filterAction: 'Ação',
    filterActionPlaceholder: 'Filtrar por ação',
    filterStartDate: 'Data Inicial',
    filterEndDate: 'Data Final',
    userID: 'ID do Usuário',
    action: 'Ação',
    resource: 'Recurso',
    detail: 'Detalhe',
    ip: 'Endereço IP',
    time: 'Horário'
  }
}