import { Pool } from 'pg';
import { CONFIG } from './setup';

export class DatabaseUtils {
    private pool: Pool;

    constructor() {
        this.pool = new Pool({
            database: CONFIG.DATABASE_NAME,
            user: CONFIG.DATABASE_USER,
            password: CONFIG.DATABASE_PASSWORD,
            host: CONFIG.DATABASE_HOST,
            port: CONFIG.DATABASE_PORT,
        });
    }

    async cleanupDatabaseData(): Promise<void> {
        try {
            const tables = ['app_sessions', 'channels', 'contract_events', 'ledger', 'rpc_store', 'signers', 'session_keys', 'ledger_transactions', 'blockchain_actions'];

            const client = await this.pool.connect();
            try {
                await client.query('BEGIN');

                await client.query('SET session_replication_role = replica');

                for (const tableName of tables) {

                    await client.query(`TRUNCATE TABLE "${tableName}" CASCADE`);
                }

                await client.query('SET session_replication_role = DEFAULT');

                await client.query('COMMIT');
            } catch (error) {
                await client.query('ROLLBACK');
                throw error;
            } finally {
                client.release();
            }
        } catch (error) {
            console.error('Error during database data cleanup:', error);
            throw error;
        }
    }

    async resetClearnodeState(): Promise<void> {
        await this.cleanupDatabaseData();

        // Future-proof: if Clearnode adds in-memory caching, add cache-clear API call here
        // await this.clearClearnodeCache(); // Uncomment when caching is added
    }

    async getBlockchainActions(filters: { channel_id?: string; action_type?: string }): Promise<any[]> {
        const client = await this.pool.connect();
        try {
            let query = 'SELECT * FROM blockchain_actions WHERE 1=1';
            const values: any[] = [];
            let paramIndex = 1;

            if (filters.channel_id) {
                query += ` AND channel_id = $${paramIndex++}`;
                values.push(filters.channel_id);
            }

            if (filters.action_type) {
                query += ` AND action_type = $${paramIndex++}`;
                values.push(filters.action_type);
            }

            const result = await client.query(query, values);
            return result.rows;
        } finally {
            client.release();
        }
    }

    async close(): Promise<void> {
        await this.pool.end();
    }
}
