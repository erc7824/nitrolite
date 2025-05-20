import { useEffect, useRef, useState } from 'react';

// Define sound types for type safety
type SoundType = 'mark-x' | 'mark-o' | 'game-over' | 'win' | 'draw';

export function useSoundEffects() {
  const [loaded, setLoaded] = useState(false);
  const audioContext = useRef<AudioContext | null>(null);
  const buffers = useRef<Map<SoundType, AudioBuffer>>(new Map());
  
  // Initialize audio context and load sounds
  useEffect(() => {
    // Create audio context
    const ctx = new (window.AudioContext || (window as any).webkitAudioContext)();
    audioContext.current = ctx;
    
    // Define the sounds to preload
    // const soundsToLoad: Array<{ type: SoundType; url: string }> = [
    //   // These URLs would point to your actual sound files
    //   { type: 'mark-x', url: '/sounds/mark-x.mp3' },
    //   { type: 'mark-o', url: '/sounds/mark-o.mp3' },
    //   { type: 'game-over', url: '/sounds/game-over.mp3' },
    //   { type: 'win', url: '/sounds/win.mp3' },
    //   { type: 'draw', url: '/sounds/draw.mp3' },
    // ];
    
    // Instead of loading real files, we'll generate audio buffers programmatically
    // This is a fallback for development when sound files might not exist
    const generateToneBuffer = (frequency: number, duration: number, fadeOut = true): AudioBuffer => {
      const sampleRate = ctx.sampleRate;
      const buffer = ctx.createBuffer(2, sampleRate * duration, sampleRate);
      
      // Generate sound data for both channels
      for (let channel = 0; channel < 2; channel++) {
        const channelData = buffer.getChannelData(channel);
        
        for (let i = 0; i < channelData.length; i++) {
          // Basic sine wave
          const t = i / sampleRate;
          let value = Math.sin(2 * Math.PI * frequency * t);
          
          // Add a bit of randomness for more texture
          const noise = Math.random() * 0.05;
          value = value * 0.5 + value * noise;
          
          // Apply fade out if needed
          if (fadeOut && t > duration * 0.7) {
            const fadePosition = (t - duration * 0.7) / (duration * 0.3);
            value *= 1 - fadePosition;
          }
          
          channelData[i] = value;
        }
      }
      
      return buffer;
    };
    
    // Generate different sounds for different actions
    const xMarkBuffer = generateToneBuffer(880, 0.3); // Higher pitch for X
    const oMarkBuffer = generateToneBuffer(440, 0.3); // Lower pitch for O
    const gameOverBuffer = generateToneBuffer(220, 1, true); // Low tone for game over
    const winBuffer = generateToneBuffer(880, 0.8, true); // Celebratory tone
    const drawBuffer = generateToneBuffer(440, 0.5, true); // Neutral tone
    
    // Store the buffers
    buffers.current.set('mark-x', xMarkBuffer);
    buffers.current.set('mark-o', oMarkBuffer);
    buffers.current.set('game-over', gameOverBuffer);
    buffers.current.set('win', winBuffer);
    buffers.current.set('draw', drawBuffer);
    
    setLoaded(true);
    
    // Clean up
    return () => {
      if (audioContext.current && audioContext.current.state !== 'closed') {
        audioContext.current.close();
      }
    };
  }, []);
  
  // Function to play a sound
  const playSound = (type: SoundType, volume = 0.5) => {
    if (!loaded || !audioContext.current) return;
    
    const buffer = buffers.current.get(type);
    if (!buffer) return;
    
    // Create source and gain nodes
    const source = audioContext.current.createBufferSource();
    const gainNode = audioContext.current.createGain();
    
    // Connect nodes
    source.buffer = buffer;
    source.connect(gainNode);
    gainNode.connect(audioContext.current.destination);
    
    // Set volume
    gainNode.gain.value = volume;
    
    // Play the sound
    source.start();
  };
  
  return { playSound, loaded };
}