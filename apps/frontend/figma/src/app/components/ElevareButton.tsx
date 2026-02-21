import { ButtonHTMLAttributes, forwardRef } from 'react';

interface ElevareButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'disabled';
  size?: 'default' | 'large';
}

export const ElevareButton = forwardRef<HTMLButtonElement, ElevareButtonProps>(
  ({ variant = 'primary', size = 'default', className = '', children, disabled, ...props }, ref) => {
    const baseStyles = 'rounded-lg font-medium transition-all';
    
    const sizeStyles = {
      default: 'px-6 py-2',
      large: 'px-8 py-3',
    };
    
    const variantStyles = {
      primary: 'bg-primary text-primary-foreground hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed',
      secondary: 'bg-secondary text-secondary-foreground border border-border hover:bg-neutral-200 disabled:opacity-50 disabled:cursor-not-allowed',
      disabled: 'bg-neutral-200 text-neutral-500 cursor-not-allowed',
    };

    const combinedClassName = `${baseStyles} ${sizeStyles[size]} ${variantStyles[variant]} ${className}`;

    return (
      <button
        ref={ref}
        className={combinedClassName}
        disabled={disabled || variant === 'disabled'}
        {...props}
      >
        {children}
      </button>
    );
  }
);

ElevareButton.displayName = 'ElevareButton';
