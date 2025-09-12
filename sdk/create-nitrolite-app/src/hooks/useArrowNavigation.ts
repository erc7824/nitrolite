import { useInput } from 'ink';

interface ArrowNavigationProps {
  selectedIndex: number;
  setSelectedIndex: (index: number) => void;
  totalItems: number;
  onSelect?: () => void;
  disabled?: boolean;
}

export function useArrowNavigation({
  selectedIndex,
  setSelectedIndex,
  totalItems,
  onSelect,
  disabled = false,
}: ArrowNavigationProps) {
  useInput((input, key) => {
    if (disabled) return;
    
    if (key.upArrow) {
      setSelectedIndex(selectedIndex > 0 ? selectedIndex - 1 : totalItems - 1);
    } else if (key.downArrow) {
      setSelectedIndex(selectedIndex < totalItems - 1 ? selectedIndex + 1 : 0);
    } else if (key.return && onSelect) {
      onSelect();
    }
  });
}