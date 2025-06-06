import { AlertTriangle } from 'lucide-react';
import { Button } from './ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from './ui/dialog';

interface ErrorModalProps {
  message: string;
  onClose: () => void;
}

export function ErrorModal({ message, onClose }: ErrorModalProps) {
  // Use a simple implementation without local state to avoid possible state conflicts
  return (
    <Dialog open={true} modal={true}>
      <DialogContent 
        className="bg-viper-charcoal/90 border-red-800/30 relative overflow-hidden animate-fadeIn"
        style={{
          boxShadow: '0 0 30px rgba(255, 0, 0, 0.2), 0 0 15px rgba(180, 37, 255, 0.15), 0 4px 20px rgba(0, 0, 0, 0.4)',
          maxWidth: '28rem'
        }}
      >
        {/* Error glow effect */}
        <div className="absolute inset-0 bg-gradient-to-br from-red-900/30 to-viper-purple/20 z-0"></div>
        
        <div className="relative z-10">
          <DialogHeader className="flex flex-row items-center gap-3 pb-2">
            <div className="p-2 bg-red-900/30 rounded-full shadow-lg shadow-red-900/20">
              <AlertTriangle className="h-7 w-7 text-red-400" />
            </div>
            <DialogTitle className="text-xl font-bold text-red-400 text-shadow-sm">
              Connection Error
            </DialogTitle>
          </DialogHeader>
          
          <div className="mt-2 text-gray-300 bg-gray-800/60 p-4 rounded-md border border-red-900/30 shadow-inner">
            <p>{message}</p>
          </div>
          
          <DialogFooter className="mt-6">
            <Button
              onClick={onClose}
              type="button"
              variant="destructive"
              size="lg"
              className="w-full sm:w-auto bg-gradient-to-r from-red-600 to-viper-purple hover:from-red-500 hover:to-viper-purple-light shadow-lg transform transition-transform hover:scale-105 active:scale-95"
            >
              Return to Lobby
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  );
}