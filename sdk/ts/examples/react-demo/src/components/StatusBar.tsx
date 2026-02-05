import type { StatusMessage } from '../types';

interface StatusBarProps {
  status: StatusMessage;
  onClose: () => void;
}

export default function StatusBar({ status, onClose }: StatusBarProps) {
  const bgColors = {
    success: 'bg-green-100 border-green-400 text-green-800',
    error: 'bg-red-100 border-red-400 text-red-800',
    info: 'bg-blue-100 border-blue-400 text-blue-800',
  };

  return (
    <div className={`fixed top-4 right-4 max-w-md border-l-4 p-4 rounded shadow-lg z-50 ${bgColors[status.type]}`}>
      <div className="flex justify-between items-start">
        <div className="flex-1">
          <p className="font-semibold">{status.message}</p>
          {status.details && <p className="text-sm mt-1">{status.details}</p>}
        </div>
        <button onClick={onClose} className="ml-4 text-xl leading-none">&times;</button>
      </div>
    </div>
  );
}
