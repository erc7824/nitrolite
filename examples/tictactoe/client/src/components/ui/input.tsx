import * as React from "react";
import { cva } from "class-variance-authority";
import { cn } from "../../lib/utils";

const inputVariants = cva(
  "flex h-10 w-full rounded-md px-3 py-2 text-sm file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-gray-500 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200",
  {
    variants: {
      variant: {
        default: 
          "border border-gray-700/50 bg-gray-800/70 text-gray-200 shadow-inner focus-visible:ring-1 focus-visible:ring-cyan-500 focus-visible:border-cyan-500/50",
        cyan: 
          "border border-cyan-800/40 bg-gray-800/70 text-gray-200 shadow-inner focus-visible:ring-1 focus-visible:ring-cyan-500 focus-visible:border-cyan-500/50",
        magenta: 
          "border border-fuchsia-800/40 bg-gray-800/70 text-gray-200 shadow-inner focus-visible:ring-1 focus-visible:ring-fuchsia-500 focus-visible:border-fuchsia-500/50",
        glass: 
          "border border-gray-700/30 bg-gray-900/40 backdrop-blur-md text-gray-200 shadow-inner focus-visible:ring-1 focus-visible:ring-cyan-500/50",
      },
      size: {
        default: "h-10",
        sm: "h-8 text-xs px-2.5",
        lg: "h-12 text-base",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface InputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size' | 'prefix'> {
  icon?: React.ReactNode;
  prefix?: string;
  variant?: 'default' | 'cyan' | 'magenta' | 'glass';
  size?: 'default' | 'sm' | 'lg';
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, variant, size, icon, prefix, ...props }, ref) => {
    return (
      <div className="relative">
        {icon && (
          <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500">
            {icon}
          </div>
        )}
        {prefix && (
          <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 font-medium">
            {prefix}
          </div>
        )}
        <input
          type={type}
          className={cn(
            inputVariants({ variant, size }),
            icon && "pl-10",
            prefix && "pl-7",
            className
          )}
          ref={ref}
          {...props}
        />
      </div>
    );
  }
);
Input.displayName = "Input";

export { Input, inputVariants };