import type CspDev from 'csp-dev';

export function app(): CspDev.DirectiveDescriptor {
  return {
    'default-src': [
      '*',
    ],
  };
}
