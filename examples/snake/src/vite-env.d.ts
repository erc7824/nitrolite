/// <reference types="vite/client" />

interface Window {
  __ENV__?: {
    [key: string]: string;
  };
}
