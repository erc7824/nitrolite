export interface ApiResource {
  path: string;
  pathParams?: Array<string>;
  filterFields?: Array<string>;
  paginated?: boolean;
  headers?: RequestInit['headers'];
}

export type ApiName =
  | 'admin'
  | 'bens'
  | 'channels'
  | 'contractInfo'
  | 'general'
  | 'metadata'
  | 'rewards'
  | 'stats'
  | 'visualize';
