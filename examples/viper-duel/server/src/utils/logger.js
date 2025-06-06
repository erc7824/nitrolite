/**
 * Logger utility with color-coded console outputs
 */
import chalk from 'chalk';

// Log level colors
const colors = {
  info: chalk.blue,
  success: chalk.green,
  warn: chalk.yellow,
  error: chalk.red,
  debug: chalk.magenta,
  system: chalk.cyan,
  auth: chalk.hex('#FF8800'),
  ws: chalk.hex('#00AAFF'),
  nitro: chalk.hex('#9900FF'),
  game: chalk.hex('#00FF99')
};

// Timestamp generator
const timestamp = () => {
  const now = new Date();
  return chalk.gray(`[${now.toISOString().split('T')[1].slice(0, -1)}]`);
};

// Logger implementation
export const logger = {
  info: (message, ...args) => console.log(timestamp(), colors.info('INFO'), message, ...args),
  success: (message, ...args) => console.log(timestamp(), colors.success('SUCCESS'), message, ...args),
  warn: (message, ...args) => console.warn(timestamp(), colors.warn('WARNING'), message, ...args),
  error: (message, ...args) => console.error(timestamp(), colors.error('ERROR'), message, ...args),
  debug: (message, ...args) => console.debug(timestamp(), colors.debug('DEBUG'), message, ...args),
  system: (message, ...args) => console.log(timestamp(), colors.system('SYSTEM'), message, ...args),
  auth: (message, ...args) => console.log(timestamp(), colors.auth('AUTH'), message, ...args),
  ws: (message, ...args) => console.log(timestamp(), colors.ws('WEBSOCKET'), message, ...args),
  nitro: (message, ...args) => console.log(timestamp(), colors.nitro('NITROLITE'), message, ...args),
  game: (message, ...args) => console.log(timestamp(), colors.game('GAME'), message, ...args),
  
  // Special format for objects/data
  data: (label, data) => {
    console.log(
      timestamp(), 
      chalk.hex('#888888')('DATA'), 
      chalk.cyan(label + ':'),
      typeof data === 'object' ? '\n' + JSON.stringify(data, null, 2) : data
    );
  }
};

export default logger;