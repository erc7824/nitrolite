import { useEffect, useRef } from 'react';
// Aura effect component (legacy from tic-tac-toe)

interface CellAuraEffectProps {
  value: string;
  activated: boolean;
}

export function CellAuraEffect({ value, activated }: CellAuraEffectProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  
  // Colors based on player symbol
  const color = value === 'X' 
    ? { primary: '#00e5ff', glow: 'rgba(0, 229, 255, ' } 
    : { primary: '#ff49e1', glow: 'rgba(255, 73, 225, ' };
  
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    // Set canvas dimensions to match the parent element (cell)
    canvas.width = canvas.offsetWidth || 100;
    canvas.height = canvas.offsetHeight || 100;
    
    // Create particles
    const particlesArray: Particle[] = [];
    const numberOfParticles = 20;
    
    class Particle {
      x = 0;
      y = 0;
      size = 0;
      speedX = 0;
      speedY = 0;
      opacity = 0;
      age = 0;
      lifespan = 0;
      
      constructor(exploding: boolean = false) {
        // Place particles centrally
        this.x = (canvas?.width || 100) / 2 + (Math.random() - 0.5) * 20;
        this.y = (canvas?.height || 100) / 2 + (Math.random() - 0.5) * 20;
        
        this.size = Math.random() * 3 + 0.5;
        
        if (exploding) {
          // Exploding particles move outward in all directions
          const angle = Math.random() * Math.PI * 2;
          const speed = Math.random() * 3 + 1;
          this.speedX = Math.cos(angle) * speed;
          this.speedY = Math.sin(angle) * speed;
          this.opacity = 1;
        } else {
          // Normal particles drift randomly
          this.speedX = (Math.random() - 0.5) * 0.5;
          this.speedY = (Math.random() - 0.5) * 0.5;
          this.opacity = Math.random() * 0.7;
        }
        
        this.age = 0;
        this.lifespan = 60 + Math.random() * 60; // 1-2 seconds at 60fps
      }
      
      update() {
        this.age++;
        
        // Move particle
        this.x += this.speedX;
        this.y += this.speedY;
        
        // Gradually slow down exploding particles
        this.speedX *= 0.98;
        this.speedY *= 0.98;
        
        // Fade out toward end of life
        let alpha = this.opacity;
        if (this.age > this.lifespan * 0.7) {
          alpha = this.opacity * (1 - (this.age - this.lifespan * 0.7) / (this.lifespan * 0.3));
        }
        
        // Draw particle
        if (ctx) {
          ctx.beginPath();
          ctx.arc(this.x, this.y, this.size, 0, Math.PI * 2);
          ctx.fillStyle = color.glow + alpha + ')';
          
          // Add glow effect
          ctx.shadowColor = color.primary;
          ctx.shadowBlur = 5;
          ctx.fill();
          ctx.shadowBlur = 0;
        }
        
        // Return true if particle is still alive
        return this.age < this.lifespan;
      }
    }
    
    // Initialize particles
    const init = () => {
      particlesArray.length = 0; // Clear any existing particles
      for (let i = 0; i < numberOfParticles; i++) {
        particlesArray.push(new Particle());
      }
    };
    
    // Add explosion particles when the cell is activated
    const addExplosionParticles = () => {
      for (let i = 0; i < 30; i++) {
        particlesArray.push(new Particle(true));
      }
    };
    
    let active = false;
    
    // Animation loop
    const animate = () => {
      if (!canvas || !ctx) return;
      
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      
      // Update and filter particles (remove dead ones)
      for (let i = 0; i < particlesArray.length; i++) {
        if (!particlesArray[i].update()) {
          particlesArray.splice(i, 1);
          i--;
        }
      }
      
      // Add new particles to maintain a minimum count
      while (particlesArray.length < numberOfParticles / 2) {
        particlesArray.push(new Particle());
      }
      
      // Continue animation
      animationFrameId = requestAnimationFrame(animate);
    };
    
    // Initialize and start animation
    init();
    let animationFrameId = requestAnimationFrame(animate);
    
    // Add explosion effect when activated changes to true
    if (activated && !active) {
      addExplosionParticles();
      active = true;
    }
    
    // Handle window resize
    const handleResize = () => {
      if (!canvas) return;
      canvas.width = canvas.offsetWidth || 100;
      canvas.height = canvas.offsetHeight || 100;
    };
    
    window.addEventListener('resize', handleResize);
    
    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize);
      cancelAnimationFrame(animationFrameId);
    };
  }, [color.glow, color.primary, activated]);
  
  return (
    <canvas
      ref={canvasRef}
      className="absolute inset-0 z-5 pointer-events-none"
      style={{ opacity: 0.8 }}
    />
  );
}