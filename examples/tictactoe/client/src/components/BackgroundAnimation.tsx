import { useEffect, useRef } from 'react';

const COLORS = {
  CYAN: 'rgba(0, 229, 255, 0.05)',
  MAGENTA: 'rgba(255, 73, 225, 0.05)',
  CYAN_BRIGHT: 'rgba(0, 229, 255, 0.15)',
  MAGENTA_BRIGHT: 'rgba(255, 73, 225, 0.15)'
};

export function BackgroundAnimation() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    // Set canvas dimensions to match the window
    const resizeCanvas = () => {
      if (!canvas) return;
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
    };
    
    window.addEventListener('resize', resizeCanvas);
    resizeCanvas();
    
    // Create particles
    const particlesArray: Particle[] = [];
    const numberOfParticles = Math.min(Math.max(window.innerWidth / 15, 20), 120); // Increased particle count
    
    // Track mouse position for interactive effects
    let mouseX = 0;
    let mouseY = 0;
    
    const updateMousePosition = (e: MouseEvent) => {
      mouseX = e.clientX;
      mouseY = e.clientY;
    };
    
    window.addEventListener('mousemove', updateMousePosition);
    
    class Particle {
      x = 0;
      y = 0;
      size = 0;
      speedX = 0;
      speedY = 0;
      color = '';
      moveAmplitude = 0;
      moveFrequency = 0;
      movePhase = 0;
      age = 0;
      lifespan = 0;
      
      constructor() {
        if (!canvas) return;
        this.x = Math.random() * canvas.width;
        this.y = Math.random() * canvas.height;
        this.size = Math.random() * 6 + 1;
        this.speedX = (Math.random() - 0.5) * 0.2;
        this.speedY = (Math.random() - 0.5) * 0.2;
        this.color = Math.random() > 0.5 ? COLORS.CYAN : COLORS.MAGENTA;
        this.moveAmplitude = Math.random() * 2;
        this.moveFrequency = Math.random() * 0.01;
        this.movePhase = Math.random() * Math.PI * 2;
        this.age = 0;
        this.lifespan = 500 + Math.random() * 1000;
      }
      
      update() {
        if (!canvas || !ctx) return;
        this.age++;
        
        // Calculate distance to mouse for interactive effects
        const dx = this.x - mouseX;
        const dy = this.y - mouseY;
        const distance = Math.sqrt(dx * dx + dy * dy);
        const mouseInfluenceRadius = 150;
        
        // Apply mouse influence if within radius
        if (distance < mouseInfluenceRadius) {
          // Particles move away from mouse cursor
          const repelFactor = (1 - distance / mouseInfluenceRadius) * 0.3;
          this.x += (dx / distance) * repelFactor;
          this.y += (dy / distance) * repelFactor;
          
          // Particles near mouse are brighter
          this.color = this.color === COLORS.CYAN ? COLORS.CYAN_BRIGHT : COLORS.MAGENTA_BRIGHT;
        } else {
          // Restore original color when away from mouse
          this.color = this.color === COLORS.CYAN_BRIGHT ? COLORS.CYAN : 
                       this.color === COLORS.MAGENTA_BRIGHT ? COLORS.MAGENTA : this.color;
        }
        
        // Oscillating movement
        this.x += this.speedX + Math.sin(this.age * this.moveFrequency + this.movePhase) * this.moveAmplitude;
        this.y += this.speedY + Math.cos(this.age * this.moveFrequency + this.movePhase) * this.moveAmplitude;
        
        // Fade in and out based on age
        let opacity = 1;
        const fadeInDuration = 50;
        const fadeOutStart = this.lifespan - 200;
        
        if (this.age < fadeInDuration) {
          opacity = this.age / fadeInDuration;
        } else if (this.age > fadeOutStart) {
          opacity = (this.lifespan - this.age) / (this.lifespan - fadeOutStart);
        }
        
        // Wrap particles at screen edges
        if (this.x < 0) this.x = canvas.width;
        if (this.x > canvas.width) this.x = 0;
        if (this.y < 0) this.y = canvas.height;
        if (this.y > canvas.height) this.y = 0;
        
        // Reset particle after lifespan ends
        if (this.age >= this.lifespan) {
          this.reset();
        }
        
        // Draw particle
        ctx.beginPath();
        ctx.arc(this.x, this.y, this.size, 0, Math.PI * 2);
        
        // Adjust color opacity
        const color = this.color.slice(0, -4) + opacity + ')';
        
        ctx.fillStyle = color;
        ctx.fill();
        
        // Draw glow
        ctx.shadowColor = this.color;
        ctx.shadowBlur = 10;
        ctx.fillStyle = color;
        ctx.fill();
        
        // Reset shadow
        ctx.shadowBlur = 0;
      }
      
      reset() {
        if (!canvas) return;
        this.x = Math.random() * canvas.width;
        this.y = Math.random() * canvas.height;
        this.size = Math.random() * 6 + 1;
        this.age = 0;
        this.lifespan = 500 + Math.random() * 1000;
      }
    }
    
    // Create particles
    const init = () => {
      for (let i = 0; i < numberOfParticles; i++) {
        particlesArray.push(new Particle());
      }
    };
    
    init();
    
    // Animation loop
    const animate = () => {
      if (!canvas || !ctx) return;
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      
      // Draw each particle
      particlesArray.forEach(particle => particle.update());
      
      // Optional: connect particles with lines
      connectParticles();
      
      requestAnimationFrame(animate);
    };
    
    // Connect particles with faint lines if close enough
    const connectParticles = () => {
      if (!ctx) return;
      const maxDistance = 120;
      
      for (let a = 0; a < particlesArray.length; a++) {
        for (let b = a; b < particlesArray.length; b++) {
          const particleA = particlesArray[a];
          const particleB = particlesArray[b];
          
          const dx = particleA.x - particleB.x;
          const dy = particleA.y - particleB.y;
          const distance = Math.sqrt(dx * dx + dy * dy);
          
          if (distance < maxDistance) {
            // Calculate opacity based on distance
            const opacity = 1 - (distance / maxDistance);
            
            // Determine line color as a mix of the two particles
            const colorA = particleA.color === COLORS.CYAN ? 'rgba(0, 229, 255, ' : 'rgba(255, 73, 225, ';
            const colorB = particleB.color === COLORS.CYAN ? 'rgba(0, 229, 255, ' : 'rgba(255, 73, 225, ';
            
            // Create a gradient for the line
            const gradient = ctx.createLinearGradient(particleA.x, particleA.y, particleB.x, particleB.y);
            gradient.addColorStop(0, colorA + (opacity * 0.05) + ')');
            gradient.addColorStop(1, colorB + (opacity * 0.05) + ')');
            
            ctx.strokeStyle = gradient;
            ctx.lineWidth = 0.3;
            ctx.beginPath();
            ctx.moveTo(particleA.x, particleA.y);
            ctx.lineTo(particleB.x, particleB.y);
            ctx.stroke();
          }
        }
      }
    };
    
    // Start animation
    animate();
    
    // Add a method to create a burst of particles
    const createParticleBurst = (x: number, y: number, color: string, count = 10) => {
      for (let i = 0; i < count; i++) {
        const particle = new Particle();
        particle.x = x;
        particle.y = y;
        particle.color = color;
        particle.size = Math.random() * 4 + 1;
        particle.speedX = (Math.random() - 0.5) * 3;
        particle.speedY = (Math.random() - 0.5) * 3;
        particle.lifespan = 150 + Math.random() * 100;
        particlesArray.push(particle);
      }
    };
    
    // Listen for custom events from cell marks
    const handleCellMarked = (e: CustomEvent) => {
      const { x, y, isX } = e.detail;
      createParticleBurst(
        x, 
        y, 
        isX ? COLORS.CYAN_BRIGHT : COLORS.MAGENTA_BRIGHT,
        15
      );
    };
    
    window.addEventListener('cellMarked', handleCellMarked as EventListener);
    
    // Cleanup on component unmount
    return () => {
      window.removeEventListener('resize', resizeCanvas);
      window.removeEventListener('mousemove', updateMousePosition);
      window.removeEventListener('cellMarked', handleCellMarked as EventListener);
    };
  }, []);
  
  return (
    <canvas
      ref={canvasRef}
      className="fixed inset-0 z-1 pointer-events-none"
      style={{ opacity: 0.6 }}
    />
  );
}
