/**
 * Auth stubs -- v1.0.0 has no authentication concept.
 * These functions are no-ops that allow existing auth code to compile
 * while doing nothing at runtime.
 */

export interface AuthRequestParams {
    address: string;
    session_key: string;
    application: string;
    expires_at: bigint;
    scope: string;
    allowances: any[];
}

export async function createAuthRequestMessage(_params: AuthRequestParams): Promise<string> {
    return JSON.stringify({ req: [0, 'auth_request', {}, Date.now()], sig: '0x' });
}

export async function createAuthVerifyMessage(_signer: any, _response: any): Promise<string> {
    return JSON.stringify({ req: [0, 'auth_verify', {}, Date.now()], sig: '0x' });
}

export async function createAuthVerifyMessageWithJWT(_jwt: string): Promise<string> {
    return JSON.stringify({ req: [0, 'auth_verify', {}, Date.now()], sig: '0x' });
}

export function createEIP712AuthMessageSigner(
    _walletClient: any,
    _params: any,
    _domain: any,
): any {
    return async () => '0x';
}
