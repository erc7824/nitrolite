import { DEFAULTS } from '../constants/defaults.js';

/**
 * Calculates progress percentage for a given current/total
 */
export function calculateProgress(current: number, total: number): number {
  if (total === 0) return 0;
  return Math.round((current / total) * DEFAULTS.PROGRESS.COMPLETE);
}

/**
 * Creates a progress updater function for iterative operations
 */
export function createProgressUpdater(
  total: number, 
  onProgress: (percent: number) => void
) {
  return (current: number) => {
    const progress = calculateProgress(current + 1, total);
    onProgress(progress);
  };
}

/**
 * Progress manager for multi-step operations
 */
export class StepProgressManager {
  private readonly stepWeights: number[];
  private readonly onProgress: (percent: number) => void;
  private currentStep = 0;
  private stepProgress = 0;

  constructor(stepWeights: number[], onProgress: (percent: number) => void) {
    this.stepWeights = stepWeights;
    this.onProgress = onProgress;
  }

  /**
   * Moves to the next step and resets step progress
   */
  nextStep(): void {
    this.currentStep = Math.min(this.currentStep + 1, this.stepWeights.length - 1);
    this.stepProgress = 0;
    this.updateProgress();
  }

  /**
   * Updates progress within current step
   */
  updateStepProgress(progress: number): void {
    this.stepProgress = Math.max(0, Math.min(100, progress));
    this.updateProgress();
  }

  /**
   * Calculates and reports overall progress
   */
  private updateProgress(): void {
    const totalWeight = this.stepWeights.reduce((sum, weight) => sum + weight, 0);
    let completedWeight = 0;

    // Add weight of completed steps
    for (let i = 0; i < this.currentStep; i++) {
      completedWeight += this.stepWeights[i];
    }

    // Add current step progress
    if (this.currentStep < this.stepWeights.length) {
      const currentStepWeight = this.stepWeights[this.currentStep];
      completedWeight += (currentStepWeight * this.stepProgress) / DEFAULTS.PROGRESS.COMPLETE;
    }

    const overallProgress = Math.round((completedWeight / totalWeight) * DEFAULTS.PROGRESS.COMPLETE);
    this.onProgress(overallProgress);
  }

  /**
   * Marks all steps as complete
   */
  complete(): void {
    this.currentStep = this.stepWeights.length;
    this.stepProgress = DEFAULTS.PROGRESS.COMPLETE;
    this.onProgress(DEFAULTS.PROGRESS.COMPLETE);
  }
}