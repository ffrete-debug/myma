import { create } from 'zustand';
// Use the configured instance, not bare axios: it carries the 401 handler
// that clears a dead session and sends the user to /login. Calls made with
// bare axios skipped it, so a token for a user that no longer exists (after
// the database is reset, say) produced a console error per caller and left
// the page stranded instead of redirecting once.
import axios from '@/lib/axios';
import Cookies from 'js-cookie';

// Define server object type

/** A server id as it appears anywhere in the app: the API sends a number, but
 *  route params and <select> values are strings. */
export type ServerId = number | string;

/** Compare ids without caring which representation the caller holds. Using ===
 *  compared a number against a string and never matched, so status updates were
 *  silently dropped. */
export function sameServer(a: ServerId, b: ServerId): boolean {
    return String(a) === String(b);
}

export interface Server {
    /** The API returns a JSON number here. It was typed as a string, which let
     *  callers assume string methods were available and produced a runtime
     *  "trim is not a function" crash on the players page. */
    id: number;
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
    updateServer: (serverId: ServerId, updateData: Partial<Server>) => Promise<Server>;
    deleteServer: (serverId: ServerId) => Promise<void>;
    getServer: (serverId: ServerId) => Promise<Server>;
    getImageStatus: () => Promise<void>;
    startServer: (serverId: ServerId) => Promise<void>;
    stopServer: (serverId: ServerId) => Promise<void>;
    restartServer: (serverId: ServerId) => Promise<void>;
    updateServerStatus: (serverId: ServerId, status: Server['status']) => void;
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
                    servers: state.servers.map((s) => (sameServer(s.id, serverId) ? updatedServer : s)),
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
                    servers: state.servers.filter((s) => !sameServer(s.id, serverId)),
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
