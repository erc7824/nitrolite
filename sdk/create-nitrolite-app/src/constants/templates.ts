export interface Template {
  id: string;
  name: string;
  description: string;
  features: string[];
}

export const TEMPLATES: Template[] = [
  {
    id: 'nextjs-app',
    name: 'Next.js App Router',
    description: 'Next.js with App Router, TypeScript, and TailwindCSS',
    features: ['Next.js 15', 'App Router', 'TypeScript', 'TailwindCSS', 'SSR support'],
  },
  {
    id: 'nodejs-server',
    name: 'Node.js Server',
    description: 'Express server with WebSocket support and Nitrolite SDK',
    features: ['Express.js', 'WebSocket Server', 'TypeScript', 'Nitrolite SDK', 'Hot Reload'],
  },
];