import { config, isDevelopment } from '../config/index.js';

/**
 * Simple logger utility
 */
class Logger {
  private getTimestamp(): string {
    return new Date().toISOString();
  }

  private formatMessage(level: string, message: string, ...args: any[]): string {
    const timestamp = this.getTimestamp();
    const formattedArgs =
      args.length > 0
        ? ' ' +
        args.map((arg) => (typeof arg === 'object' ? JSON.stringify(arg, null, 2) : String(arg))).join(' ')
        : '';

    return `[${timestamp}] [${level}] ${message}${formattedArgs}`;
  }

  info(message: string, ...args: any[]): void {
    console.log(this.formatMessage('INFO', message, ...args));
  }

  warn(message: string, ...args: any[]): void {
    console.warn(this.formatMessage('WARN', message, ...args));
  }

  error(message: string, ...args: any[]): void {
    console.error(this.formatMessage('ERROR', message, ...args));
  }

  debug(message: string, ...args: any[]): void {
    if (isDevelopment) {
      console.log(this.formatMessage('DEBUG', message, ...args));
    }
  }

  trace(message: string, ...args: any[]): void {
    if (isDevelopment) {
      console.trace(this.formatMessage('TRACE', message, ...args));
    }
  }
}

export const logger = new Logger();
