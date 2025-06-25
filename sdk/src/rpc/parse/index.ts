import { RPCResponse, RPCMethod } from "../types";
import { paramsParsers, ParamsParser } from "./response";

export function parseRPCResponse(response: string): RPCResponse {
  try {
    const parsed = JSON.parse(response);
    if (!Array.isArray(parsed.res) || parsed.res.length !== 4) {
      throw new Error('Invalid RPC response format');
    }
    const method = parsed.res[1] as RPCMethod;
    const parse = paramsParsers[method] as ParamsParser<any>;
    if (!parse) {
      throw new Error(`No parser found for method ${method}`);
    }
    const params = parse(parsed.res[2]);
    const responseObj = {
      method,
      requestId: parsed.res[0],
      timestamp: parsed.res[3],
      signatures: parsed.sig || [],
      params,
    } as RPCResponse;
    return responseObj;
  } catch (e) {
    throw new Error(`Failed to parse RPC response: ${e instanceof Error ? e.message : e}`);
  }
}
