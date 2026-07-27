import { create } from 'zustand';
import axios from 'axios';
import Cookies from 'js-cookie';

// Define server object type
export interface Server {
    id: string;
    identifier: string;
    session_name: string;
    status: 'running' | 'stopped' | 'starting' | 'stopping' | 'restarting';
    port: number;
    query_port: number;
    rcon_port: number;
    admin_password: string;
    map: string;
    max_players: number;
    game_user_settings?: string;
    game_ini?: string;
    server_args?: Record<string, unknown>;
    created_at: string;
    updated_at: string;
}

/**
 * Stable, translatable error codes. Stores cannot call `useTranslations`, so we
 * expose a code and let components resolve it against the `errors` namespace
 * (e.g. `useTranslations('errors')(code)`).
 */
export type ServersErrorCode =
    | 'authTokenMissing'
    | 'fetchServersFailed'
    | 'createServerFailed'
    | 'updateServerFailed'
    | 'deleteServerFailed'
    | 'getServerFailed'
    | 'getImageStatusFailed'
    | 'startServerFailed'
    | 'stopServerFailed'
    | 'restartServerFailed';

/** Error thrown by this store, carrying a translatable code. */
export class ServersError extends Error {
    readonly code: ServersErrorCode;

    constructor(code: ServersErrorCode) {
        super(code);
        this.name = 'ServersError';
        this.code = code;
    }
}

const toErrorCode = (error: unknown, fallback: ServersErrorCode): ServersErrorCode =>
    error instanceof ServersError ? error.code : fallback;

// Define server state type
interface ServersState {
    servers: Server[];
    isLoading: boolean;
    error: ServersErrorCode | null;
    imageStatus: ImageStatus | null;
    currentPage: number;
    pageSize: number;
    totalServers: number;
    actions: ServersActions;
}

export interface ImageStatus {
    can_create_server: boolean;
    can_start_server: boolean;
    any_pulling: boolean;
    any_not_ready: boolean;
    overall_status: string;
    pulling_count: number;
    total_images: number;
    images: {
        [imageName: string]: {
            ready: boolean;
            pulling: boolean;
            has_update?: boolean;
            layers?: {
                [layerId: string]: {
                    id: string;
                    status: 'pending' | 'downloading' | 'extracting' | 'verifying' | 'complete';
                    progress: number;
                    size: number;
                };
            };
        };
    };
}

// Define server actions type
interface ServersActions {
    fetchServers: (page?: number, limit?: number) => Promise<void>;
    setPage: (page: number, limit?: number) => void;
    createServer: (serverData: Partial<Server>) => Promise<Server>;
    updateServer: (serverId: string, updateData: Partial<Server>) => Promise<Server>;
    deleteServer: (serverId: string) => Promise<void>;
    getServer: (serverId: string) => Promise<Server>;
    getImageStatus: () => Promise<void>;
    startServer: (serverId: string) => Promise<void>;
    stopServer: (serverId: string) => Promise<void>;
    restartServer: (serverId: string) => Promise<void>;
    updateServerStatus: (serverId: string, status: Server['status']) => void;
}

const getAuthHeaders = () => {
    const token = Cookies.get('auth-token');
    if (!token) {
        throw new ServersError('authTokenMissing');
    }
    return {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
    };
};

const useServersStore = create<ServersState>((set, get) => ({
    servers: [],
    isLoading: false,
    error: null,
    imageStatus: null,
    currentPage: 1,
    pageSize: 20,
    totalServers: 0,
    actions: {
        fetchServers: async (page = 1, limit = 20) => {
            if (get().isLoading) return;
            set({ isLoading: true, error: null, currentPage: page, pageSize: limit });
            try {
                const response = await axios.get('/api/servers', {
                    headers: getAuthHeaders(),
                    params: { page, limit },
                });
                set({
                    servers: response.data.data || [],
                    totalServers: response.data.total || 0,
                });
            } catch (error) {
                set({ error: toErrorCode(error, 'fetchServersFailed') });
                throw error;
            } finally {
                set({ isLoading: false });
            }
        },
        setPage: (page, limit = 20) => {
            get().actions.fetchServers(page, limit);
        },
        createServer: async (serverData) => {
            try {
                const response = await axios.post('/api/servers', serverData, { headers: getAuthHeaders() });
                const newServer = response.data.data;
                set((state) => ({ servers: [...state.servers, newServer] }));
                return newServer;
            } catch (error) {
                set({ error: toErrorCode(error, 'createServerFailed') });
                throw error;
            }
        },
        updateServer: async (serverId, updateData) => {
            try {
                const response = await axios.put(`/api/servers/${serverId}`, updateData, { headers: getAuthHeaders() });
                const updatedServer = response.data.data;
                set((state) => ({
                    servers: state.servers.map((s) => (s.id === serverId ? updatedServer : s)),
                }));
                return updatedServer;
            } catch (error) {
                set({ error: toErrorCode(error, 'updateServerFailed') });
                throw error;
            }
        },
        deleteServer: async (serverId) => {
            try {
                await axios.delete(`/api/servers/${serverId}`, { headers: getAuthHeaders() });
                set((state) => ({
                    servers: state.servers.filter((s) => s.id !== serverId),
                }));
            } catch (error) {
                set({ error: toErrorCode(error, 'deleteServerFailed') });
                throw error;
            }
        },
        getServer: async (serverId) => {
            try {
                const response = await axios.get(`/api/servers/${serverId}`, { headers: getAuthHeaders() });
                return response.data.data;
            } catch (error) {
                set({ error: toErrorCode(error, 'getServerFailed') });
                throw error;
            }
        },
        getImageStatus: async () => {
            try {
                const response = await axios.get('/api/images/status', { headers: getAuthHeaders() });
                set({ imageStatus: response.data.data });
            } catch (error) {
                set({ error: toErrorCode(error, 'getImageStatusFailed') });
                throw error;
            }
        },
        startServer: async (serverId) => {
            try {
                await axios.post(`/api/servers/${serverId}/start`, {}, { headers: getAuthHeaders() });
                // Optimistic intermediate state only. The terminal status ('running'
                // or 'stopped' if the container failed) arrives via the status socket
                // or the next fetch - never fabricate it on a timer.
                get().actions.updateServerStatus(serverId, 'starting');
            } catch (error) {
                set({ error: toErrorCode(error, 'startServerFailed') });
                throw error;
            }
        },
        stopServer: async (serverId) => {
            try {
                await axios.post(`/api/servers/${serverId}/stop`, {}, { headers: getAuthHeaders() });
                // Optimistic intermediate state only; the real status follows.
                get().actions.updateServerStatus(serverId, 'stopping');
            } catch (error) {
                set({ error: toErrorCode(error, 'stopServerFailed') });
                throw error;
            }
        },
        restartServer: async (serverId) => {
            try {
                await axios.post(`/api/servers/${serverId}/restart`, {}, { headers: getAuthHeaders() });
                // Optimistic intermediate state only; the real status follows.
                get().actions.updateServerStatus(serverId, 'restarting');
            } catch (error) {
                set({ error: toErrorCode(error, 'restartServerFailed') });
                throw error;
            }
        },
        updateServerStatus: (serverId, status) => {
            set((state) => ({
                servers: state.servers.map((s) => (s.id === serverId ? { ...s, status } : s)),
            }));
        },
    },
}));

export const useServers = () => useServersStore((state) => state.servers);
export const useServersIsLoading = () => useServersStore((state) => state.isLoading);
export const useServersError = () => useServersStore((state) => state.error);
export const useImageStatus = () => useServersStore((state) => state.imageStatus);
export const useServersCurrentPage = () => useServersStore((state) => state.currentPage);
export const useServersPageSize = () => useServersStore((state) => state.pageSize);
export const useServersTotal = () => useServersStore((state) => state.totalServers);
export const serversActions = useServersStore.getState().actions;

export default useServersStore;
