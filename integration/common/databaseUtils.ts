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
            const tables = ['app_sessions', 'channels', 'contract_events', 'ledger', 'rpc_store', 'signers', 'transactions'];

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

    async close(): Promise<void> {
        await this.pool.end();
    }
}
