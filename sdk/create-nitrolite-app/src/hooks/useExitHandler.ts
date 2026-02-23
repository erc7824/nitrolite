import { useInput } from 'ink';

export function useExitHandler() {
  useInput((input, key) => {
    if (key.ctrl && input === 'c') {
      process.exit(0);
    }
  });
}