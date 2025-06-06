import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "../../lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-lg text-sm font-medium ring-offset-background transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 transform hover:scale-[1.02] active:scale-[0.98] cursor-pointer select-none",
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
        viperGreen: "bg-gradient-to-br from-viper-green via-viper-green to-viper-green-dark text-viper-charcoal shadow-lg shadow-viper-green/25 hover:shadow-viper-green/40 hover:shadow-xl hover:from-viper-green-light hover:to-viper-green border-2 border-viper-green/40 hover:border-viper-green/60 font-bold tracking-wide relative overflow-hidden before:absolute before:inset-0 before:bg-gradient-to-r before:from-transparent before:via-white/20 before:to-transparent before:translate-x-[-100%] hover:before:translate-x-[100%] before:transition-transform before:duration-700",
        viperPurple: "bg-gradient-to-br from-viper-purple via-viper-purple to-viper-purple-dark text-white shadow-lg shadow-viper-purple/25 hover:shadow-viper-purple/40 hover:shadow-xl hover:from-viper-purple-light hover:to-viper-purple border-2 border-viper-purple/40 hover:border-viper-purple/60 font-bold tracking-wide relative overflow-hidden before:absolute before:inset-0 before:bg-gradient-to-r before:from-transparent before:via-white/20 before:to-transparent before:translate-x-[-100%] hover:before:translate-x-[100%] before:transition-transform before:duration-700",
        viperYellow: "bg-gradient-to-br from-viper-yellow via-viper-yellow to-viper-yellow-dark text-viper-charcoal shadow-lg shadow-viper-yellow/25 hover:shadow-viper-yellow/40 hover:shadow-xl hover:from-viper-yellow-light hover:to-viper-yellow border-2 border-viper-yellow/40 hover:border-viper-yellow/60 font-bold tracking-wide relative overflow-hidden before:absolute before:inset-0 before:bg-gradient-to-r before:from-transparent before:via-white/20 before:to-transparent before:translate-x-[-100%] hover:before:translate-x-[100%] before:transition-transform before:duration-700",
        glass: "bg-viper-charcoal/30 backdrop-blur-md border border-viper-charcoal-light/40 text-white shadow-sm hover:bg-viper-charcoal/40 hover:border-viper-green/30",
        glowGreen: "bg-viper-charcoal/60 text-viper-green border border-viper-green/50 shadow-[0_0_10px_rgba(42,255,107,0.1)] hover:shadow-[0_0_15px_rgba(42,255,107,0.2)] hover:bg-viper-charcoal/80 hover:border-viper-green/60",
        glowPurple: "bg-viper-charcoal/60 text-viper-purple border border-viper-purple/50 shadow-[0_0_10px_rgba(180,37,255,0.1)] hover:shadow-[0_0_15px_rgba(180,37,255,0.2)] hover:bg-viper-charcoal/80 hover:border-viper-purple/60",
      },
      size: {
        default: "h-10 px-6 py-2 text-sm",
        sm: "h-8 px-4 text-xs",
        lg: "h-12 px-8 text-base",
        xl: "h-14 px-10 text-lg font-medium",
        xxl: "h-16 px-12 text-xl font-medium",
        icon: "h-10 w-10",
        wide: "h-12 px-16 text-base",
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