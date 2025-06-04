const getEnvVar = (key: string, defaultValue: string): string => {
    try {
        const envValue = (import.meta.env?.[`VITE_${key}`] || window?.__ENV__?.[key] || null) as string | null;
        return envValue || defaultValue;
    } catch (e) {
        console.warn(`Could not access environment variable ${key}, using default value`);
        return defaultValue;
    }
};

export const BROKER_WS_URL = getEnvVar("BROKER_WS_URL", "");
export const GAMESERVER_WS_URL = getEnvVar("GAMESERVER_WS_URL", "ws://localhost:3001");
export const CHAIN_ID = parseInt(getEnvVar("CHAIN_ID", "137"), 10);
