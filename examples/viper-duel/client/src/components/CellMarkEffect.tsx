import { useState, useEffect } from 'react';
// Mark effect component (legacy from tic-tac-toe)
import { cn } from '../lib/utils';

interface CellMarkEffectProps {
  value: string;
  activated: boolean;
}

export function CellMarkEffect({ value, activated }: CellMarkEffectProps) {
  const [showEffect, setShowEffect] = useState(false);
  
  // Reset and show the effect when activated changes
  useEffect(() => {
    if (activated) {
      setShowEffect(true);
      
      // Remove elements after animation completes
      const timer = setTimeout(() => {
        setShowEffect(false);
      }, 1500); // Match duration with animation
      
      return () => clearTimeout(timer);
    }
  }, [activated]);
  
  if (!showEffect) return null;
  
  const isX = value === 'X';
  // const baseColor = isX ? 'cyan' : 'magenta';
  
  return (
    <div className="absolute inset-0 pointer-events-none overflow-hidden">
      {/* Center burst */}
      <div className={cn(
        "absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-16 h-16 rounded-full",
        isX ? "bg-cyan-500/30 animate-explode-cyan" : "bg-fuchsia-500/30 animate-explode-magenta"
      )}></div>
      
      {/* Orbiting particles */}
      {Array.from({ length: 6 }).map((_, i) => (
        <div
          key={i}
          className={cn(
            "absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2",
            "w-2 h-2 rounded-full opacity-70",
            isX ? "bg-cyan-400" : "bg-fuchsia-400",
            i % 2 === 0 ? "animate-orbit" : "animate-orbit-reverse",
          )}
          style={{
            animationDelay: `${i * 0.1}s`,
            transform: `rotate(${i * 60}deg) translateX(${20 + i * 5}px)`,
          }}
        ></div>
      ))}
      
      {/* Random floating particles */}
      {Array.from({ length: 12 }).map((_, i) => {
        // Calculate random positions and delays
        const angle = Math.random() * Math.PI * 2;
        const distance = 20 + Math.random() * 40;
        const size = 1 + Math.random() * 2;
        const delay = Math.random() * 0.5;
        const duration = 0.5 + Math.random() * 1;
        
        const x = Math.cos(angle) * distance;
        const y = Math.sin(angle) * distance;
        
        return (
          <div
            key={i + 'float'}
            className={cn(
              "absolute left-1/2 top-1/2 rounded-full opacity-70",
              isX ? "bg-cyan-400" : "bg-fuchsia-400"
            )}
            style={{
              width: `${size}px`,
              height: `${size}px`,
              transform: `translate(${x}px, ${y}px)`,
              opacity: 0,
              animation: `fadeIn 0.2s ease-out forwards ${delay}s, fadeOut 0.7s ease-in forwards ${delay + duration}s`,
            }}
          ></div>
        );
      })}
      
      {/* Lines radiating outward */}
      {Array.from({ length: 8 }).map((_, i) => {
        const angle = (i * Math.PI) / 4; // 8 directions
        const delay = i * 0.05;
        
        return (
          <div
            key={i + 'line'}
            className={cn(
              "absolute left-1/2 top-1/2 h-0.5 origin-left",
              isX ? "bg-cyan-500" : "bg-fuchsia-500"
            )}
            style={{
              width: '1px',
              transform: `rotate(${angle}rad)`,
              animation: `lineGrow 0.4s ease-out forwards ${delay}s`,
            }}
          ></div>
        );
      })}
      
      {/* Add custom keyframes for line growth with a style tag */}
      <style>{`
        @keyframes lineGrow {
          0% {
            width: 0;
            opacity: 0.8;
          }
          100% {
            width: 50px;
            opacity: 0;
          }
        }
        
        @keyframes fadeOut {
          0% {
            opacity: 0.7;
          }
          100% {
            opacity: 0;
          }
        }
      `}</style>
    </div>
  );
}