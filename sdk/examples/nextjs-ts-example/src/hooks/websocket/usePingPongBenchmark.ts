import { useRef, useCallback } from 'react';
import { WebSocketClient } from '@/websocket';
import { useMessageService } from '../useMessageService';

// Custom hook to manage ping-pong benchmarking
export function usePingPongBenchmark(clientRef: React.RefObject<WebSocketClient>, messageService: ReturnType<typeof useMessageService>) {
    const { addSystemMessage, addErrorMessage, addPingMessage, addPongMessage } = messageService;
    const benchmarkInProgress = useRef<boolean>(false);
    
    // Session counters
    const pingCount = useRef<number>(0);
    const pongCount = useRef<number>(0);
    
    // Total counts for current user
    const userPingCount = useRef<number>(0);
    const userPongCount = useRef<number>(0);
    
    // Total counts for guest/server
    const guestPingCount = useRef<number>(0);
    const guestPongCount = useRef<number>(0);
    
    const getTotalStats = useCallback(() => {
        return {
            user: {
                pings: userPingCount.current,
                pongs: userPongCount.current,
                total: userPingCount.current + userPongCount.current
            },
            guest: {
                pings: guestPingCount.current,
                pongs: guestPongCount.current,
                total: guestPingCount.current + guestPongCount.current
            },
            total: userPingCount.current + userPongCount.current + guestPingCount.current + guestPongCount.current
        };
    }, []);
    
    const runPingPongBenchmark = useCallback(async (count: number) => {
        if (!clientRef.current?.isConnected) return;
        
        // Reset session counters
        pingCount.current = 0;
        pongCount.current = 0;
        benchmarkInProgress.current = true;
        const startTime = Date.now();
        const stats = getTotalStats();
        
        addSystemMessage(`Starting ping-pong benchmark with ${count} pings at ${new Date().toISOString()}...`);
        addSystemMessage(`Current P2P message counts - User: ${stats.user.total} (${stats.user.pings} pings, ${stats.user.pongs} pongs), Guest: ${stats.guest.total} (${stats.guest.pings} pings, ${stats.guest.pongs} pongs)`);
        
        try {
            // First ping to start the cycle
            const pingStartTime = Date.now();
            try {
                await clientRef.current.ping();
                const pingDuration = Date.now() - pingStartTime;
                pingCount.current++;
                userPingCount.current++;
                
                // Show starting ping with updated stats
                const timestamp = new Date().toISOString();
                const statsAfterPing = getTotalStats();
                
                addPingMessage(`[${timestamp}] PING #${pingCount.current}/${count} (${pingDuration}ms)`, 'user');
                addSystemMessage(`Initiated ping-pong loop. Waiting for responses...`);
                addSystemMessage(`P2P message counts - User: ${statsAfterPing.user.total} (${statsAfterPing.user.pings} pings, ${statsAfterPing.user.pongs} pongs), Guest: ${statsAfterPing.guest.total} (${statsAfterPing.guest.pings} pings, ${statsAfterPing.guest.pongs} pongs)`);
            } catch (pingError) {
                await handleInitialPingError(pingError, count);
            }
            
            setupProgressTracker(count, startTime);
        } catch (error) {
            handleBenchmarkFailure(error, count);
        }
    }, [clientRef, addSystemMessage, addErrorMessage, addPingMessage, getTotalStats]);
    
    const handleInitialPingError = useCallback(async (pingError: unknown, count: number) => {
        const errorMsg = pingError instanceof Error ? pingError.message : String(pingError);
        addErrorMessage(`[${new Date().toISOString()}] INITIAL PING ERROR: ${errorMsg}`);
        
        if (errorMsg.includes('timeout')) {
            addSystemMessage('Timeout on initial ping. Trying again in 2 seconds...');
            
            try {
                await clientRef.current.ping();
                pingCount.current++;
                userPingCount.current++;
                
                const stats = getTotalStats();
                addPingMessage(
                    `[${new Date().toISOString()}] RETRY PING #${pingCount.current}/${count}`,
                    'user',
                );
                addSystemMessage(`Retry successful. Ping-pong loop started.`);
                addSystemMessage(`P2P message counts - User: ${stats.user.total} (${stats.user.pings} pings, ${stats.user.pongs} pongs), Guest: ${stats.guest.total} (${stats.guest.pings} pings, ${stats.guest.pongs} pongs)`);
            } catch (retryError) {
                addErrorMessage(
                    `[${new Date().toISOString()}] RETRY FAILED: ${retryError instanceof Error ? retryError.message : String(retryError)}`,
                );
                throw retryError;
            }
        } else {
            throw pingError;
        }
    }, [clientRef, addErrorMessage, addSystemMessage, addPingMessage, getTotalStats]);
    
    const setupProgressTracker = useCallback((count: number, startTime: number) => {
        const progressInterval = setInterval(() => {
            if (!benchmarkInProgress.current) {
                clearInterval(progressInterval);
                return;
            }
            
            const elapsedTime = (Date.now() - startTime) / 100;
            const pingsPerSecond = pingCount.current / elapsedTime;
            const pongsPerSecond = pongCount.current / elapsedTime;
            const stats = getTotalStats();
            
            addSystemMessage(
                `Progress: ${pingCount.current}/${count} pings (${pingsPerSecond.toFixed(2)}/sec), ${pongCount.current} pongs (${pongsPerSecond.toFixed(2)}/sec)`
            );
            addSystemMessage(
                `P2P message counts - User: ${stats.user.total} (${stats.user.pings} pings, ${stats.user.pongs} pongs), Guest: ${stats.guest.total} (${stats.guest.pings} pings, ${stats.guest.pongs} pongs)`
            );
            
            if (pingCount.current >= count) {
                finishBenchmark(count, startTime, progressInterval);
            }
        }, 100);
    }, [addSystemMessage, getTotalStats]);
    
    const finishBenchmark = useCallback((count: number, startTime: number, progressInterval: NodeJS.Timeout) => {
        clearInterval(progressInterval);
        benchmarkInProgress.current = false;
        
        const endTime = Date.now();
        const totalTime = (endTime - startTime) / 1000;
        const stats = getTotalStats();
        
        addSystemMessage(
            `Benchmark complete: ${pingCount.current} pings, ${pongCount.current} pongs in ${totalTime.toFixed(2)} seconds.`
        );
        addSystemMessage(
            `Performance: ${(pingCount.current / totalTime).toFixed(2)} pings/sec, ${(pongCount.current / totalTime).toFixed(2)} pongs/sec`
        );
        addSystemMessage(
            `P2P message counts - User: ${stats.user.total} (${stats.user.pings} pings, ${stats.user.pongs} pongs), Guest: ${stats.guest.total} (${stats.guest.pings} pings, ${stats.guest.pongs} pongs), Total: ${stats.total}`
        );
    }, [addSystemMessage, getTotalStats]);
    
    const handleBenchmarkFailure = useCallback((error: unknown, count: number) => {
        const errorMsg = error instanceof Error ? error.message : String(error);
        console.error(`Ping attempts failed:`, error);
        addErrorMessage(`[${new Date().toISOString()}] PING FAILURE: ${errorMsg}`);
        addSystemMessage(`Benchmark aborted due to ping failures. Please check your connection and try again.`);
        
        if (errorMsg.includes('timeout')) {
            addSystemMessage(`Will attempt to restart benchmark in 5 seconds...`);
            setTimeout(() => {
                if (clientRef.current?.isConnected) {
                    addSystemMessage(`Restarting ping-pong benchmark automatically...`);
                    benchmarkInProgress.current = false;
                    pingCount.current = 0;
                    pongCount.current = 0;
                    runPingPongBenchmark(count).catch((e) => {
                        addErrorMessage(`Auto-restart failed: ${e instanceof Error ? e.message : String(e)}`);
                    });
                }
            }, 300);
        } else {
            benchmarkInProgress.current = false;
        }
    }, [clientRef, runPingPongBenchmark, addSystemMessage, addErrorMessage]);
    
    const handlePongResponse = useCallback((timestamp: string) => {
        pongCount.current++;
        guestPongCount.current++;
        
        const stats = getTotalStats();
        addPongMessage(`[${timestamp}] SERVER PONG received #${pongCount.current}/1000 (Guest pongs: ${guestPongCount.current})`, 'guest');
        
        if (benchmarkInProgress.current && pingCount.current < 1000) {
            sendNextPingInBenchmark();
        }
    }, [addPongMessage, getTotalStats]);
    
    const sendNextPingInBenchmark = useCallback(() => {
        try {
            clientRef.current
                .ping()
                .then(() => {
                    pingCount.current++;
                    userPingCount.current++;
                    const responseTime = new Date().toISOString();
                    addPingMessage(`[${responseTime}] USER PING #${pingCount.current}/1000 (User pings: ${userPingCount.current})`, 'user');
                })
                .catch((error) => {
                    handlePingError(error);
                });
        } catch (error) {
            console.error('Failed to send ping during benchmark:', error);
        }
    }, [clientRef, addPingMessage]);
    
    const handlePingError = useCallback((error: unknown) => {
        console.error('Ping error during benchmark:', error);
        const errorMsg = error instanceof Error ? error.message : String(error);
        addErrorMessage(`[${new Date().toISOString()}] PING ERROR: ${errorMsg}`);
        
        if (errorMsg.includes('timeout') && benchmarkInProgress.current) {
            setTimeout(() => {
                if (benchmarkInProgress.current && clientRef.current?.isConnected) {
                    addSystemMessage(`Attempting to resume ping sequence after timeout...`);
                    clientRef.current
                        .ping()
                        .then(() => {
                            pingCount.current++;
                            userPingCount.current++;
                            const stats = getTotalStats();
                            addPingMessage(
                                `[${new Date().toISOString()}] RESUMED PING #${pingCount.current}/1000 (User pings: ${userPingCount.current})`,
                                'user'
                            );
                        })
                        .catch((e) => {
                            addErrorMessage(
                                `[${new Date().toISOString()}] Failed to resume: ${e instanceof Error ? e.message : String(e)}`
                            );
                        });
                }
            }, 0);
        }
    }, [clientRef, addSystemMessage, addErrorMessage, addPingMessage, getTotalStats]);
    
    // Handle receipt of a ping message from a guest
    const handleGuestPingMessage = useCallback((sender: string, timestamp: string) => {
        guestPingCount.current++;
        const stats = getTotalStats();
        
        addPingMessage(`[${timestamp}] GUEST PING from ${sender} (Guest pings: ${guestPingCount.current})`, 'guest');
        
        // Send a response pong
        if (clientRef.current?.isConnected) {
            clientRef.current
                .ping()
                .then(() => {
                    userPongCount.current++;
                    const responseTime = new Date().toISOString();
                    const updatedStats = getTotalStats();
                    addPongMessage(`[${responseTime}] USER PONG response (User pongs: ${userPongCount.current})`, 'user');
                })
                .catch((error) => {
                    console.error('Pong error:', error);
                });
        }
    }, [clientRef, addPingMessage, addPongMessage, getTotalStats]);
    
    // Handle receipt of a pong message from a guest
    const handleGuestPongMessage = useCallback((sender: string, timestamp: string) => {
        guestPongCount.current++;
        const stats = getTotalStats();
        
        addPongMessage(`[${timestamp}] GUEST PONG from ${sender} (Guest pongs: ${guestPongCount.current})`, 'guest');
        
        // Send a new ping, but only if not in benchmark
        if (clientRef.current?.isConnected && !benchmarkInProgress.current) {
            clientRef.current
                .ping()
                .then(() => {
                    userPingCount.current++;
                    const responseTime = new Date().toISOString();
                    const updatedStats = getTotalStats();
                    addPingMessage(`[${responseTime}] USER PING after pong (User pings: ${userPingCount.current})`, 'user');
                })
                .catch((error) => {
                    console.error('Ping error:', error);
                });
        }
    }, [clientRef, benchmarkInProgress, addPingMessage, addPongMessage, getTotalStats]);
    
    return {
        benchmarkInProgress,
        pingCount,
        pongCount,
        userPingCount,
        userPongCount,
        guestPingCount,
        guestPongCount,
        runPingPongBenchmark,
        handlePongResponse,
        handleGuestPingMessage,
        handleGuestPongMessage,
        getTotalStats,
    };
}