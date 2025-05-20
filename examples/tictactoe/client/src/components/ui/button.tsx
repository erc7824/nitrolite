import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "../../lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 transform hover:scale-[1.02] active:scale-[0.98]",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground shadow hover:bg-primary/90 hover:shadow-primary/20 hover:shadow-lg",
        destructive:
          "bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90 hover:shadow-lg",
        outline:
          "border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground",
        secondary:
          "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80 hover:shadow-secondary/20 hover:shadow-lg",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
        cyan: "bg-gradient-to-r from-cyan-600 to-cyan-700 text-white shadow-md hover:shadow-cyan-500/30 hover:shadow-lg hover:from-cyan-500 hover:to-cyan-600 border border-cyan-700/30",
        magenta: "bg-gradient-to-r from-fuchsia-600 to-fuchsia-700 text-white shadow-md hover:shadow-fuchsia-500/30 hover:shadow-lg hover:from-fuchsia-500 hover:to-fuchsia-600 border border-fuchsia-700/30",
        glass: "bg-gray-900/30 backdrop-blur-md border border-gray-800/40 text-white shadow-sm hover:bg-gray-900/40 hover:border-cyan-800/30",
        glowCyan: "bg-gray-900/60 text-cyan-400 border border-cyan-900/50 shadow-[0_0_10px_rgba(0,229,255,0.1)] hover:shadow-[0_0_15px_rgba(0,229,255,0.2)] hover:bg-gray-900/80 hover:border-cyan-800/60",
        glowMagenta: "bg-gray-900/60 text-fuchsia-400 border border-fuchsia-900/50 shadow-[0_0_10px_rgba(255,73,225,0.1)] hover:shadow-[0_0_15px_rgba(255,73,225,0.2)] hover:bg-gray-900/80 hover:border-fuchsia-800/60",
      },
      size: {
        default: "h-9 px-4 py-2",
        sm: "h-8 rounded-md px-3 text-xs",
        lg: "h-10 rounded-md px-8",
        xl: "h-12 rounded-md px-8 text-base font-medium",
        xxl: "h-14 rounded-md px-10 text-lg font-medium",
        icon: "h-9 w-9",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, leftIcon, rightIcon, children, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      >
        {leftIcon && <span className="mr-2">{leftIcon}</span>}
        {children}
        {rightIcon && <span className="ml-2">{rightIcon}</span>}
      </Comp>
    );
  }
);
Button.displayName = "Button";

export { Button, buttonVariants };