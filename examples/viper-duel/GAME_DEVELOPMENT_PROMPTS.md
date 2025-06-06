# Game Development Prompts & Requirements Template

This document contains comprehensive prompts and patterns for converting tic-tac-toe games to other multiplayer games with betting systems. Originally built for Viper Duel Snake game, these templates work for any real-time multiplayer game.

## Initial Setup & Architecture Analysis

### 1. Codebase Analysis
```
init is analyzing your codebase…
Please analyze this codebase and create a CLAUDE.md file, which will be given to future instances of Claude Code to operate in this repository.
```

**Purpose**: Understand existing codebase structure and create documentation for future development.

## Game Conversion Framework

### Converting from Tic-Tac-Toe to Any Multiplayer Game

Use this systematic approach to convert a tic-tac-toe base to any real-time multiplayer game:

#### Step 1: Core Game Logic Replacement
```
I want to convert this tic-tac-toe game to [YOUR GAME TYPE]. Please:
1. Replace the game logic in services/[game].js
2. Update the game state interfaces
3. Implement [GAME-SPECIFIC] mechanics
4. Keep the existing room management and WebSocket infrastructure
```

#### Step 2: Visual Component Updates
```
Replace the tic-tac-toe board with [YOUR GAME] visuals:
1. Create new game rendering component (Board.tsx → [Game]Canvas.tsx)
2. Update game state display components
3. Implement [GAME-SPECIFIC] visual effects
4. Maintain responsive design and theming
```

#### Step 3: Input/Control System
```
Update the input system for [YOUR GAME]:
1. Replace click-based moves with [INPUT TYPE] (keyboard/mouse/touch)
2. Implement real-time input handling
3. Add input validation and anti-cheat measures
4. Support multiple input methods if needed
```

## Betting System Implementation

### Complete Betting System Setup

Use this comprehensive prompt to add betting functionality to any game:

```
Add a flexible betting system to this game with the following requirements:

BETTING AMOUNTS:
- Support bet amounts: free, 0.01, 0.1, 1, 2 (in USDC)
- Winner-takes-all pot system (betAmount × 2)
- Display currency as $ symbol (not USDC)

CLIENT IMPLEMENTATION:
1. Add bet amount types (BetAmount, BetOption) to types/index.ts
2. Update interfaces: GameState, JoinRoomPayload, AvailableRoom
3. Create bet selection UI in lobby with visual feedback
4. Add bet amount badges to available rooms with proper alignment
5. Update all WebSocket message payloads to include betAmount
6. Fix TypeScript errors in game state management

SERVER IMPLEMENTATION:
1. Update room management to support bet amounts in roomManager.js
2. Add bet amount validation in room creation and joining
3. Implement bet amount matching (players can only join matching stakes)
4. Update formatGameState to include bet amounts
5. Add comprehensive bet amount validation in validators.js
6. Update all WebSocket message handlers

UI REQUIREMENTS:
- 3x2 grid layout for bet selection with visual feedback
- Bet amount badges on available rooms (Free, $0.01, $0.1, $1, $2)
- Winner-takes-all pot calculation display
- Proper alignment and responsive design
- Remove any hardcoded betting messages

VALIDATION REQUIREMENTS:
- Client and server-side bet amount validation
- Room bet amount matching enforcement
- Error handling for bet mismatches
- Type safety throughout the system
```

### Betting UI Components Pattern

```
Create a professional betting interface with:

BET SELECTION:
- Grid layout with buttons for each bet amount
- Visual feedback (glow effect for selected amount)
- Winner pot calculation display
- Clean typography and spacing

ROOM DISPLAY:
- Bet amount badges on available rooms
- Proper alignment and truncation handling
- Color coding (gray for free, amber for paid)
- Responsive layout for different stake amounts

CURRENCY DISPLAY:
- Use $ symbol consistently (not USDC)
- Proper decimal formatting ($0.01, not $0.010)
- Clear pot calculations (Winner takes all! Total pot: $2.00)
```

## Visual Enhancement Patterns

### Canvas-Based Game Rendering

For smooth, professional game visuals, replace basic CSS grids with canvas rendering:

```
Replace the blocky [GAME ELEMENTS] with smooth, glowing, brand-consistent visuals:

Visual spec:
- [MAIN ELEMENTS] – [SPECIFIC DESIGN REQUIREMENTS]
- [SECONDARY ELEMENTS] – [DESIGN DETAILS]
- Neon fill/glow with brand colors
- Enhanced animations and effects

Create radial gradients per [GAME ELEMENT]:
- [COLOR SCHEME 1]: #[HEX] → #[HEX] → rgba([RGB],0)
- [COLOR SCHEME 2]: #[HEX] → #[HEX] → rgba([RGB],0)

Effects:
- ctx.shadowBlur = 12; ctx.shadowColor = color;
- Highlight pass with CRT shine effect
- Smooth transitions and animations
```

### Brand Identity Implementation

```
Complete rebrand from [OLD NAME] to [NEW NAME] with:

BRAND ELEMENTS:
- Update all references in codebase
- Package names and descriptions
- Documentation and README files
- Authentication domains and app identifiers

VISUAL DESIGN:
- Logo references and taglines
- Color palette and themes
- Visual effects and particles
- Typography and spacing

TECHNICAL UPDATES:
- EIP-712 domain names
- WebSocket app names
- Error messages and labels
- File and component names
```

## Architecture Patterns

### Real-Time Game Loop Implementation

```
Convert from turn-based to real-time gameplay:

MOVEMENT SYSTEM:
- Implement automatic [GAME MECHANICS] (every Xms)
- Player input only changes [CONTROL VARIABLES]
- Continuous [GAME STATE] updates
- Remove turn-based validation

GAME LOOP:
- Server-side game state management
- Broadcast updates to all players
- Handle disconnections gracefully
- Implement game over detection

TIMING:
- Configurable game speed (200ms-1000ms intervals)
- Input buffering and validation
- Smooth client-side prediction
- Server authoritative state
```

### WebSocket Message Architecture

```
Implement comprehensive real-time messaging:

CLIENT → SERVER:
- joinRoom (with bet amount)
- [GAME_ACTION] (player input)
- getAvailableRooms
- appSession signatures

SERVER → CLIENT:
- room:state (game state updates)
- room:ready (room full notification)
- game:started (game begin)
- game:update (real-time updates)
- game:over (end state)
- error (validation failures)

VALIDATION:
- Server-side input validation
- Rate limiting and anti-cheat
- State synchronization
- Error handling and recovery
```

## Core Game Mechanics Issues

### Snake Game Specific Examples

These examples show the progression from tic-tac-toe to Snake game mechanics:

### 2. Movement Controls Fix
```
ok now i receive error Cannot reverse direction

movements doesnt work we need support wasd and arrows movements
```

**Issues to Fix**:
- Remove "Cannot reverse direction" restrictions
- Add support for both WASD and arrow key controls
- Fix keyboard event detection (case sensitivity issues)

**Solution Applied**:
- Fixed arrow key detection (`ArrowUp` vs `arrowup`)
- Removed reverse direction validation
- Added dual control support (WASD + arrows)

### 3. Automatic Movement Implementation
```
and also when game start in defult game snakes is move imidiatly
```

**Requirement**: Implement automatic Snake movement like classic arcade games
- Snakes should move automatically every second
- Players only change direction, don't trigger movement
- Continuous movement in current direction

### 4. Movement Logic Clarification
```
no it start move imidiatly movements and user controle it
```

**Clarification**: User wants classic Snake behavior:
- Automatic movement every second in current direction
- User input only changes direction
- Not manual movement per key press

### 5. UI Movement Issues
```
0 movements on ui, maybe issue on client
```

**Problem**: No visual movement on the client side
**Root Cause**: Game loop wasn't starting in app session flow
**Solution**: Added missing `startGameLoop()` call to app session initialization

### 6. Game Ending & Scoring Issues
```
Final Scores
Player 1: 1
Player 2: 0
Game Time: 17 seconds
Player 1 wins!

now it works can 

but i just got 1 score, and game end only when snake touch each other, it can have n amoutn of score and make more fast change of direction
```

**Issues Identified**:
- Limited scoring (only 1 point possible)
- Game only ends on snake-to-snake collision
- Need faster direction changes
- Want unlimited scoring potential

**Fixes Applied**:
- Multiple food items (3+ always available)
- Faster game loop (500ms instead of 1000ms)
- Fixed collision detection for walls and self-collision
- Unlimited food respawning

### 7. Real-Time Movement Request
```
lets do movements update in real time without delay, and i just collect first item to eat and game was end
```

**Final Requirements**:
- **Real-time movement**: No delays, instant response to key presses
- **Bug fix**: Game ending after first food collection

**Implementation**:
- Removed automatic timer-based movement
- Movement happens instantly on key press
- Added comprehensive debugging for collision detection
- Fixed premature game ending bugs

## Technical Implementation Patterns

### Movement System Evolution

**Phase 1: Automatic Timer-Based**
```javascript
setInterval(() => {
  // Move all snakes automatically every X ms
  updateGameState();
}, intervalTime);
```

**Phase 2: Real-Time Input-Based**
```javascript
// On direction change:
changeDirection(direction) {
  updateDirection(direction);
  moveSnake(playerId); // Immediate movement
  broadcastUpdate();
}
```

### Collision Detection Requirements
- Wall collisions should end game
- Self-collisions should end game  
- Snake-to-snake collisions should end game
- Food collisions should increase score and grow snake

### Food & Scoring System
- Multiple food items on board (3+ minimum)
- Unlimited scoring potential
- Automatic food respawning
- Score increases with each food eaten

### Network Architecture
- Real-time WebSocket communication
- Client-server state synchronization
- Immediate broadcast of movement updates
- Game over detection and cleanup

## Common Game Development Issues & Solutions

### 1. Keyboard Input Handling
**Problem**: Arrow keys not working
**Solution**: Check case sensitivity (`ArrowUp` vs `arrowup`)

### 2. Game Loop Management
**Problem**: No visual updates
**Solution**: Ensure game loop starts in all code paths (regular + app session)

### 3. Collision Detection
**Problem**: Game ending unexpectedly
**Solution**: Add comprehensive logging for all collision types

### 4. Real-Time vs Turn-Based
**Problem**: Delayed responses
**Solution**: Move from timer-based to event-driven movement

### 5. Multiplayer Synchronization
**Problem**: Players seeing different states
**Solution**: Immediate broadcast after each player action

## Debug Logging Patterns

### Movement Debugging
```javascript
console.log(`Key pressed: ${event.key}, Direction: ${direction}`);
console.log(`🐍 GAME UPDATE:`, { gameTime, player1Pos, player2Pos });
```

### Collision Debugging
```javascript
console.log(`💥 ${playerId} hit wall at (${x}, ${y})`);
console.log(`🔄 ${playerId} collided with self`);
console.log(`🐍 ${playerId} collided with ${otherPlayer}`);
```

### Game State Debugging
```javascript
console.log(`🍎 ${playerId} ate food! New score: ${score}`);
console.log(`🔍 Game state: player1 alive=${alive1}, player2 alive=${alive2}`);
```

## Performance Optimization Patterns

### Network Optimization
- Only broadcast state changes, not full state every frame
- Use efficient message types (`room:state`, `game:update`, `game:over`)
- Implement game over detection loops separate from movement

### Client-Side Optimization
- Prevent default browser behavior for game keys
- Use React state updates efficiently
- Implement proper cleanup for event listeners

## Future Game Development Checklist

### Initial Setup
- [ ] Analyze existing codebase structure
- [ ] Create comprehensive CLAUDE.md documentation
- [ ] Set up client-server architecture

### Movement System
- [ ] Implement keyboard input handling (WASD + arrows)
- [ ] Choose movement model (real-time vs timer-based)
- [ ] Add collision detection (walls, self, others)
- [ ] Test movement responsiveness

### Game Mechanics
- [ ] Implement scoring system
- [ ] Add multiple objectives (food items)
- [ ] Set up win/lose conditions
- [ ] Add game over detection

### Multiplayer Features
- [ ] WebSocket real-time communication
- [ ] State synchronization between players
- [ ] Room management system
- [ ] Player authentication

### Debug & Testing
- [ ] Add comprehensive logging
- [ ] Test all collision scenarios
- [ ] Verify multiplayer synchronization
- [ ] Performance testing

### Polish & UX
- [ ] Visual feedback for actions
- [ ] Sound effects (optional)
- [ ] Responsive controls
- [ ] Clear game state indicators

## Template Commands for Claude

### Quick Fixes
```
"Fix [specific issue] - [brief description]"
"Add support for [feature] - [requirements]"
"Debug [problem] - showing [symptoms]"
```

### Feature Requests
```
"Implement [feature] with [specific requirements]"
"Make [system] work like [reference/example]"
"Add [functionality] that [specific behavior]"
```

### Technical Issues
```
"No [expected behavior] on [platform/component]"
"[Component] not working - [specific symptoms]"
"[Feature] only works [partial condition], need [full requirement]"
```

This template captures the iterative development process and can be used as a reference for building similar real-time multiplayer games.

## Complete Development Workflow

### Recommended Development Sequence

Use this step-by-step workflow for converting any tic-tac-toe base to a new multiplayer game:

#### Phase 1: Foundation Setup (1-2 hours)
```
1. Analyze existing codebase and create CLAUDE.md
2. Update brand identity (name, colors, taglines)
3. Replace core game logic (tic-tac-toe → your game)
4. Update game state interfaces and types
```

#### Phase 2: Core Mechanics (2-4 hours)
```
1. Implement game-specific input system
2. Add real-time game loop and state updates
3. Replace visual components (Board → GameCanvas)
4. Add game-specific validation and rules
```

#### Phase 3: Betting System (1-2 hours)
```
1. Add betting types and interfaces
2. Update UI with bet selection components
3. Implement server-side bet validation
4. Test room matching and error handling
```

#### Phase 4: Visual Polish (2-3 hours)
```
1. Replace basic visuals with canvas rendering
2. Add neon effects, gradients, and animations
3. Implement smooth visual transitions
4. Add responsive design and accessibility
```

#### Phase 5: Testing & Deployment (1 hour)
```
1. Test all betting scenarios and game mechanics
2. Verify WebSocket message handling
3. Test multiplayer synchronization
4. Deploy and monitor performance
```

### Quick Start Templates

#### Basic Game Conversion
```
Convert this tic-tac-toe game to [GAME NAME]:

1. Replace game logic in services/snake.js with [GAME] mechanics
2. Update GameState interface for [GAME] requirements
3. Create [GAME]Canvas.tsx for visual rendering
4. Implement [INPUT_TYPE] controls (keyboard/mouse/touch)
5. Add real-time game loop with [TIMING] updates
6. Keep existing room management and WebSocket infrastructure
```

#### Betting System Integration
```
Add flexible betting (free, $0.01, $0.1, $1, $2) to this game:

CLIENT SIDE:
- Add BetAmount type and betting interfaces
- Create bet selection UI with visual feedback
- Update all message payloads to include betAmount
- Add bet badges to available rooms

SERVER SIDE:
- Update room creation/joining with bet validation
- Implement bet matching enforcement
- Add comprehensive error handling
- Update game state formatting

UI POLISH:
- Use $ symbol for currency display
- Add winner-takes-all pot calculations
- Ensure proper alignment and responsive design
```

#### Visual Enhancement Upgrade
```
Replace basic [GAME] elements with smooth, glowing visuals:

CANVAS RENDERING:
- Create smooth [GAME ELEMENTS] with neon effects
- Implement radial gradients and glow effects
- Add directional animations and transitions
- Use brand-consistent color schemes

EFFECTS:
- ctx.shadowBlur = 12 for glow effects
- Highlight passes with CRT shine
- Smooth animations between game states
- Particle effects for game events
```

### Best Practices Checklist

#### Code Quality
- [ ] TypeScript strict mode enabled
- [ ] Comprehensive input validation (client + server)
- [ ] Error handling for all edge cases
- [ ] Consistent code formatting and naming

#### Game Design
- [ ] Balanced game mechanics and timing
- [ ] Clear win/lose conditions
- [ ] Smooth player experience and feedback
- [ ] Mobile-responsive design

#### Betting System
- [ ] Secure bet amount validation
- [ ] Room stake matching enforcement
- [ ] Clear UI for bet selection and display
- [ ] Proper currency formatting

#### Visual Polish
- [ ] Brand-consistent color schemes
- [ ] Smooth animations and transitions
- [ ] Professional typography and spacing
- [ ] Accessibility considerations

#### Testing
- [ ] Multiplayer synchronization testing
- [ ] Bet amount validation scenarios
- [ ] Edge case handling (disconnections, etc.)
- [ ] Performance testing with multiple rooms

This framework enables rapid development of professional multiplayer games with betting systems from any tic-tac-toe foundation.