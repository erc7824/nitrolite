import { parseAnyRPCResponse, RPCMethod } from '@erc7824/nitrolite';
import { isDevelopment } from '../../config/index.js';
import { logger } from '../../utils/logger.js';
import { EventEmitter } from './event-emitter.js';
import type { MessageEvent as WSMessageEvent } from 'ws';

export interface MessageEvents {
    authChallenge: any;
    authVerify: any;
    pong: any;
    error: any;
    assets: any;
    general: any;
}

export interface ParsedMessage {
    type: 'rpc' | 'raw' | 'invalid';
    method?: string;
    data: any;
    originalEvent: WSMessageEvent;
}

export class MessageRouter {
    private readonly authChallengeEmitter = new EventEmitter<any>();
    private readonly authVerifyEmitter = new EventEmitter<any>();
    private readonly pongEmitter = new EventEmitter<any>();
    private readonly errorEmitter = new EventEmitter<any>();
    private readonly assetsEmitter = new EventEmitter<any>();
    private readonly generalEmitter = new EventEmitter<any>();

    routeMessage(event: WSMessageEvent): void {
        const parsed = this.parseMessage(event);
        
        logger.info(`🎯 Routing message type: ${parsed.type}, method: ${parsed.method || 'none'}`);
        
        switch (parsed.type) {
            case 'rpc':
                logger.info(`🔀 Routing RPC message: ${parsed.method}`);
                this.routeRPCMessage(parsed.data, parsed.method!);
                break;
            case 'raw':
                logger.info(`📤 Routing raw message`);
                this.routeRawMessage(parsed.data);
                break;
            case 'invalid':
                logger.warn(`❌ Invalid message skipped`);
                if (isDevelopment) {
                    logger.debug('Skipped invalid message:', parsed.data);
                }
                break;
        }
    }

    private parseMessage(event: WSMessageEvent): ParsedMessage {
        const dataStr = event.data.toString();
        
        // Add debug logging to track parsing
        logger.debug('🔍 Parsing message:', dataStr.substring(0, 100) + '...');
        
        try {
            const rawData = JSON.parse(dataStr);
            
            // Try parseAnyRPCResponse first - this should handle Yellow network format
            try {
                const rpcResponse = parseAnyRPCResponse(dataStr);
                logger.debug('✅ Successfully parsed as RPC:', rpcResponse.method);
                return {
                    type: 'rpc',
                    method: rpcResponse.method,
                    data: rpcResponse,
                    originalEvent: event
                };
            } catch (rpcError) {
                logger.debug('❌ parseAnyRPCResponse failed:', rpcError instanceof Error ? rpcError.message : rpcError);
                
                // Fallback for assets messages that might not parse as RPC
                if (rawData.method === RPCMethod.Assets) {
                    logger.debug('📦 Handling as raw assets message');
                    return {
                        type: 'raw',
                        method: RPCMethod.Assets,
                        data: rawData,
                        originalEvent: event
                    };
                }
                
                logger.debug('📤 Handling as generic raw message');
                return {
                    type: 'raw',
                    data: rawData,
                    originalEvent: event
                };
            }
        } catch (parseError) {
            logger.debug('💥 JSON parse failed:', parseError instanceof Error ? parseError.message : parseError);
            return {
                type: 'invalid',
                data: dataStr,
                originalEvent: event
            };
        }
    }

    private routeRPCMessage(data: any, method: string): void {
        logger.info(`🚀 Routing RPC method: ${method}`);
        
        switch (method) {
            case RPCMethod.AuthChallenge:
                logger.info('🤝 Emitting auth_challenge to handlers');
                this.authChallengeEmitter.emit(data);
                break;
            case RPCMethod.AuthVerify:
                logger.info('✅ Emitting auth_verify to handlers');
                this.authVerifyEmitter.emit(data);
                break;
            case RPCMethod.Pong:
                logger.info('🏓 Emitting pong to handlers');
                this.pongEmitter.emit(data);
                break;
            case RPCMethod.Error:
                logger.info('❌ Emitting error to handlers');
                this.errorEmitter.emit(data);
                break;
            default:
                logger.info(`📨 Emitting ${method} to general handlers`);
                this.generalEmitter.emit(data);
                break;
        }
    }

    private routeRawMessage(data: any): void {
        if (data.method === RPCMethod.Assets) {
            this.assetsEmitter.emit(data);
        } else {
            this.generalEmitter.emit(data);
        }
    }

    onAuthChallenge(listener: (data: any) => void): () => void {
        return this.authChallengeEmitter.add(listener);
    }

    onAuthVerify(listener: (data: any) => void): () => void {
        return this.authVerifyEmitter.add(listener);
    }

    onPong(listener: (data: any) => void): () => void {
        return this.pongEmitter.add(listener);
    }

    onError(listener: (data: any) => void): () => void {
        return this.errorEmitter.add(listener);
    }

    onAssets(listener: (data: any) => void): () => void {
        return this.assetsEmitter.add(listener);
    }

    onGeneralMessage(listener: (data: any) => void): () => void {
        return this.generalEmitter.add(listener);
    }

    destroy(): void {
        this.authChallengeEmitter.clear();
        this.authVerifyEmitter.clear();
        this.pongEmitter.clear();
        this.errorEmitter.clear();
        this.assetsEmitter.clear();
        this.generalEmitter.clear();
    }
}