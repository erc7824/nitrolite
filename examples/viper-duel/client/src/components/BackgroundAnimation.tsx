import { useEffect, useRef } from 'react';

const COLORS = {
  GREEN: 'rgba(42, 255, 107, 0.25)',
  PURPLE: 'rgba(180, 37, 255, 0.25)',
  YELLOW: 'rgba(255, 237, 0, 0.3)',
  GREEN_BRIGHT: 'rgba(42, 255, 107, 0.5)',
  PURPLE_BRIGHT: 'rgba(180, 37, 255, 0.5)',
  YELLOW_BRIGHT: 'rgba(255, 237, 0, 0.6)',
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
    
    // Snake-like segments array
    const snakeSegments: SnakeSegment[] = [];
    const foodParticles: FoodParticle[] = [];
    const numberOfSnakes = Math.min(Math.max(Math.floor(window.innerWidth / 300), 4), 12);
    const numberOfFood = Math.min(Math.max(Math.floor(window.innerWidth / 150), 8), 20);
    
    // Track mouse position for interactive effects
    let mouseX = 0;
    let mouseY = 0;
    
    const updateMousePosition = (e: MouseEvent) => {
      mouseX = e.clientX;
      mouseY = e.clientY;
    };
    
    window.addEventListener('mousemove', updateMousePosition);
    
    // Snake segment class for flowing snake-like movement
    class SnakeSegment {
      x = 0;
      y = 0;
      targetX = 0;
      targetY = 0;
      size = 0;
      color = '';
      direction = 0;
      speed = 0;
      segments: { x: number; y: number; size: number }[] = [];
      segmentLength = 15;
      age = 0;
      changeDirectionCooldown = 0;
      glowing = false;
      
      constructor() {
        if (!canvas) return;
        this.x = Math.random() * canvas.width;
        this.y = Math.random() * canvas.height;
        this.targetX = this.x;
        this.targetY = this.y;
        this.size = Math.random() * 4 + 3;
        this.color = Math.random() > 0.5 ? COLORS.GREEN : COLORS.PURPLE;
        this.direction = Math.random() * Math.PI * 2;
        this.speed = Math.random() * 0.5 + 0.3;
        
        // Initialize snake segments
        for (let i = 0; i < this.segmentLength; i++) {
          this.segments.push({
            x: this.x - i * (this.size * 2),
            y: this.y,
            size: this.size * (1 - i * 0.05)
          });
        }
      }
      
      update() {
        if (!canvas || !ctx) return;
        this.age++;
        this.changeDirectionCooldown--;
        
        // Calculate distance to mouse for interactive effects
        const dx = this.x - mouseX;
        const dy = this.y - mouseY;
        const distance = Math.sqrt(dx * dx + dy * dy);
        const mouseInfluenceRadius = 200;
        
        // Snake avoids mouse cursor (like it's afraid)
        if (distance < mouseInfluenceRadius && distance > 50) {
          const avoidFactor = (1 - distance / mouseInfluenceRadius) * 0.02;
          this.direction += (Math.atan2(dy, dx) - this.direction) * avoidFactor;
          this.glowing = true;
          this.color = this.color === COLORS.GREEN ? COLORS.GREEN_BRIGHT : COLORS.PURPLE_BRIGHT;
        } else {
          this.glowing = false;
          this.color = this.color === COLORS.GREEN_BRIGHT ? COLORS.GREEN : 
                       this.color === COLORS.PURPLE_BRIGHT ? COLORS.PURPLE : this.color;
        }
        
        // Random direction changes (snake-like behavior)
        if (this.changeDirectionCooldown <= 0 && Math.random() < 0.005) {
          this.direction += (Math.random() - 0.5) * 0.5;
          this.changeDirectionCooldown = 60 + Math.random() * 120;
        }
        
        // Subtle direction wobble for organic movement
        this.direction += Math.sin(this.age * 0.01) * 0.002;
        
        // Move snake head
        this.x += Math.cos(this.direction) * this.speed;
        this.y += Math.sin(this.direction) * this.speed;
        
        // Boundary wrapping with smooth transition
        if (this.x < -50) this.x = canvas.width + 50;
        if (this.x > canvas.width + 50) this.x = -50;
        if (this.y < -50) this.y = canvas.height + 50;
        if (this.y > canvas.height + 50) this.y = -50;
        
        // Update snake segments (follow the head)
        this.segments[0] = { x: this.x, y: this.y, size: this.size };
        
        for (let i = 1; i < this.segments.length; i++) {
          const prev = this.segments[i - 1];
          const current = this.segments[i];
          
          const segmentDx = prev.x - current.x;
          const segmentDy = prev.y - current.y;
          const segmentDistance = Math.sqrt(segmentDx * segmentDx + segmentDy * segmentDy);
          
          if (segmentDistance > this.size * 1.5) {
            const moveX = (segmentDx / segmentDistance) * 0.3;
            const moveY = (segmentDy / segmentDistance) * 0.3;
            current.x += moveX;
            current.y += moveY;
          }
          
          current.size = this.size * (1 - i * 0.03);
        }
        
        this.draw();
      }
      
      draw() {
        if (!ctx) return;
        
        // Draw snake segments from tail to head
        for (let i = this.segments.length - 1; i >= 0; i--) {
          const segment = this.segments[i];
          const alpha = (this.segments.length - i) / this.segments.length;
          
          ctx.beginPath();
          ctx.arc(segment.x, segment.y, segment.size, 0, Math.PI * 2);
          
          // Color with fading alpha for tail effect
          const baseColor = this.color.slice(0, -4);
          const segmentOpacity = alpha * (this.glowing ? 0.8 : 0.4);
          ctx.fillStyle = baseColor + segmentOpacity + ')';
          
          // Add glow effect for snake segments
          if (this.glowing || i < 5) {
            ctx.shadowColor = this.color;
            ctx.shadowBlur = this.glowing ? 25 : 15;
          }
          
          ctx.fill();
          ctx.shadowBlur = 0;
        }
      }
    }
    
    // Food particle class for floating food items
    class FoodParticle {
      x = 0;
      y = 0;
      size = 0;
      age = 0;
      pulsePhase = 0;
      floatPhase = 0;
      glowing = false;
      
      constructor() {
        if (!canvas) return;
        this.x = Math.random() * canvas.width;
        this.y = Math.random() * canvas.height;
        this.size = Math.random() * 3 + 2.5;
        this.pulsePhase = Math.random() * Math.PI * 2;
        this.floatPhase = Math.random() * Math.PI * 2;
      }
      
      update() {
        if (!canvas || !ctx) return;
        this.age++;
        
        // Calculate distance to mouse
        const dx = this.x - mouseX;
        const dy = this.y - mouseY;
        const distance = Math.sqrt(dx * dx + dy * dy);
        
        // Food glows when mouse is near
        this.glowing = distance < 150;
        
        // Gentle floating motion
        this.x += Math.sin(this.age * 0.008 + this.floatPhase) * 0.1;
        this.y += Math.cos(this.age * 0.006 + this.floatPhase) * 0.08;
        
        // Wrap at boundaries
        if (this.x < 0) this.x = canvas.width;
        if (this.x > canvas.width) this.x = 0;
        if (this.y < 0) this.y = canvas.height;
        if (this.y > canvas.height) this.y = 0;
        
        this.draw();
      }
      
      draw() {
        if (!ctx) return;
        
        // Pulsing size
        const pulseSize = this.size * (1 + Math.sin(this.age * 0.05 + this.pulsePhase) * 0.3);
        
        ctx.beginPath();
        ctx.arc(this.x, this.y, pulseSize, 0, Math.PI * 2);
        
        const color = this.glowing ? COLORS.YELLOW_BRIGHT : COLORS.YELLOW;
        ctx.fillStyle = color;
        
        if (this.glowing) {
          ctx.shadowColor = COLORS.YELLOW;
          ctx.shadowBlur = 20;
        } else {
          ctx.shadowColor = COLORS.YELLOW;
          ctx.shadowBlur = 8;
        }
        
        ctx.fill();
        ctx.shadowBlur = 0;
        
        // Draw small sparkle effect
        if (this.glowing && Math.random() < 0.1) {
          const sparkleX = this.x + (Math.random() - 0.5) * pulseSize * 2;
          const sparkleY = this.y + (Math.random() - 0.5) * pulseSize * 2;
          
          ctx.beginPath();
          ctx.arc(sparkleX, sparkleY, 0.5, 0, Math.PI * 2);
          ctx.fillStyle = COLORS.YELLOW_BRIGHT;
          ctx.fill();
        }
      }
    }
    
    // Create snake segments and food
    const init = () => {
      for (let i = 0; i < numberOfSnakes; i++) {
        snakeSegments.push(new SnakeSegment());
      }
      for (let i = 0; i < numberOfFood; i++) {
        foodParticles.push(new FoodParticle());
      }
    };
    
    init();
    
    // Animation loop
    const animate = () => {
      if (!canvas || !ctx) return;
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      
      // Draw food particles first (behind snakes)
      foodParticles.forEach(food => food.update());
      
      // Draw snake segments
      snakeSegments.forEach(snake => snake.update());
      
      // Optional: connect nearby snakes with faint lines
      connectSnakes();
      
      requestAnimationFrame(animate);
    };
    
    // Connect nearby snake heads with faint energy lines
    const connectSnakes = () => {
      if (!ctx) return;
      const maxDistance = 150;
      
      for (let a = 0; a < snakeSegments.length; a++) {
        for (let b = a + 1; b < snakeSegments.length; b++) {
          const snakeA = snakeSegments[a];
          const snakeB = snakeSegments[b];
          
          const dx = snakeA.x - snakeB.x;
          const dy = snakeA.y - snakeB.y;
          const distance = Math.sqrt(dx * dx + dy * dy);
          
          if (distance < maxDistance) {
            const opacity = (1 - distance / maxDistance) * 0.08;
            
            // Different colored connection based on snake types
            const colorA = snakeA.color === COLORS.GREEN || snakeA.color === COLORS.GREEN_BRIGHT ? 'rgba(42, 255, 107, ' : 'rgba(180, 37, 255, ';
            const colorB = snakeB.color === COLORS.GREEN || snakeB.color === COLORS.GREEN_BRIGHT ? 'rgba(42, 255, 107, ' : 'rgba(180, 37, 255, ';
            
            if (snakeA.color !== snakeB.color) {
              // Different colored snakes create special connections
              const gradient = ctx.createLinearGradient(snakeA.x, snakeA.y, snakeB.x, snakeB.y);
              gradient.addColorStop(0, colorA + opacity + ')');
              gradient.addColorStop(0.5, 'rgba(255, 237, 0, ' + (opacity * 0.5) + ')'); // Yellow middle
              gradient.addColorStop(1, colorB + opacity + ')');
              
              ctx.strokeStyle = gradient;
              ctx.lineWidth = 1.2;
              ctx.beginPath();
              ctx.moveTo(snakeA.x, snakeA.y);
              ctx.lineTo(snakeB.x, snakeB.y);
              ctx.stroke();
            }
          }
        }
      }
    };
    
    // Start animation
    animate();
    
    // Create special effects for game events
    const createSnakeExplosion = (x: number, y: number, isPlayer1: boolean, count = 8) => {
      const color = isPlayer1 ? COLORS.GREEN_BRIGHT : COLORS.PURPLE_BRIGHT;
      
      // Create temporary explosion snake segments
      for (let i = 0; i < count; i++) {
        const angle = (Math.PI * 2 * i) / count;
        const snake = new SnakeSegment();
        snake.x = x;
        snake.y = y;
        snake.direction = angle;
        snake.speed = Math.random() * 2 + 1;
        snake.color = color;
        snake.segmentLength = 8;
        snake.glowing = true;
        
        // Remove after short time
        setTimeout(() => {
          const index = snakeSegments.indexOf(snake);
          if (index > -1) snakeSegments.splice(index, 1);
        }, 2000);
        
        snakeSegments.push(snake);
      }
    };
    
    // Listen for game events
    const handleGameEvent = (e: CustomEvent) => {
      const { x, y, type, isPlayer1 } = e.detail;
      
      if (type === 'snakeMove') {
        createSnakeExplosion(x, y, isPlayer1, 5);
      } else if (type === 'foodEaten') {
        // Create food explosion
        for (let i = 0; i < 6; i++) {
          const food = new FoodParticle();
          food.x = x;
          food.y = y;
          food.glowing = true;
          foodParticles.push(food);
          
          setTimeout(() => {
            const index = foodParticles.indexOf(food);
            if (index > -1) foodParticles.splice(index, 1);
          }, 1500);
        }
      }
    };
    
    window.addEventListener('gameEvent', handleGameEvent as EventListener);
    
    // Cleanup on component unmount
    return () => {
      window.removeEventListener('resize', resizeCanvas);
      window.removeEventListener('mousemove', updateMousePosition);
      window.removeEventListener('gameEvent', handleGameEvent as EventListener);
    };
  }, []);
  
  return (
    <canvas
      ref={canvasRef}
      className="fixed inset-0 z-1 pointer-events-none"
      style={{ opacity: 0.9 }}
    />
  );
}