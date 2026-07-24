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

// Define server state type
interface ServersState {
    servers: Server[];
    isLoading: boolean;
    error: string | null;
    imageStatus: ImageStatus | null;
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
    fetchServers: () => Promise<void>;
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
        throw new Error('Authentication token not found');
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
    actions: {
        fetchServers: async () => {
            if (get().isLoading) return;
            set({ isLoading: true, error: null });
            try {
                const response = await axios.get('/api/servers', { headers: getAuthHeaders() });
                set({ servers: response.data.data || [] });
            } catch (error) {
                set({ error: 'Failed to fetch server list' });
                throw error;
            } finally {
                set({ isLoading: false });
            }
        },
        createServer: async (serverData) => {
            try {
                const response = await axios.post('/api/servers', serverData, { headers: getAuthHeaders() });
                const newServer = response.data.data;
                set((state) => ({ servers: [...state.servers, newServer] }));
                return newServer;
            } catch (error) {
                set({ error: 'Failed to create server' });
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
                set({ error: 'Failed to update server' });
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
                set({ error: 'Failed to delete server' });
                throw error;
            }
        },
        getServer: async (serverId) => {
            try {
                const response = await axios.get(`/api/servers/${serverId}`, { headers: getAuthHeaders() });
                return response.data.data;
            } catch (error) {
                set({ error: 'Failed to get server info' });
                throw error;
            }
        },
        getImageStatus: async () => {
            try {
                const response = await axios.get('/api/images/status', { headers: getAuthHeaders() });
                set({ imageStatus: response.data.data });
            } catch (error) {
                set({ error: 'Failed to get image status' });
                throw error;
            }
        },
        startServer: async (serverId) => {
            try {
                await axios.post(`/api/servers/${serverId}/start`, {}, { headers: getAuthHeaders() });
                get().actions.updateServerStatus(serverId, 'starting');
                setTimeout(() => get().actions.updateServerStatus(serverId, 'running'), 3000);
            } catch (error) {
                set({ error: 'Failed to start server' });
                throw error;
            }
        },
        stopServer: async (serverId) => {
            try {
                await axios.post(`/api/servers/${serverId}/stop`, {}, { headers: getAuthHeaders() });
                get().actions.updateServerStatus(serverId, 'stopping');
                setTimeout(() => get().actions.updateServerStatus(serverId, 'stopped'), 2000);
            } catch (error) {
                set({ error: 'Failed to stop server' });
                throw error;
            }
        },
        restartServer: async (serverId) => {
            try {
                await axios.post(`/api/servers/${serverId}/restart`, {}, { headers: getAuthHeaders() });
                get().actions.updateServerStatus(serverId, 'restarting');
                setTimeout(() => get().actions.updateServerStatus(serverId, 'running'), 5000);
            } catch (error) {
                set({ error: 'Failed to restart server' });
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
export const serversActions = useServersStore.getState().actions;

export default useServersStore;
