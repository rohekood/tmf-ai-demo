export interface DebugMessage {
    id: string;
    timestamp: string;
    type: 'command' | 'event' | 'query' | 'unknown';
    topic: string;
    correlationId?: string;
    replyTo?: string;
    payload: Record<string, unknown>;
    service: string;
}

export interface DebugFilterState {
    search: string;
    services: string[];
    types: string[]; // command, event, query
}
