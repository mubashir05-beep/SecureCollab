<script>
  /** @type {"primary" | "secondary" | "ghost" | "danger" | "sage" | "clay"} */
  export let variant = "primary";
  /** @type {"sm" | "md" | "lg"} */
  export let size = "md";
  export let type = "button";
  export let fullWidth = false;
  export let disabled = false;
  export let loading = false;

  const base = "inline-flex items-center justify-center gap-2 font-bold rounded-2xl transition-all duration-200 focus:outline-none focus:ring-4 focus:ring-sage/10 disabled:opacity-40 disabled:cursor-not-allowed disabled:scale-100 select-none hover:scale-[1.03] active:scale-95 shadow-sm";

  const variants = {
    primary:   "bg-charcoal text-white hover:bg-charcoal/90",
    sage:      "bg-sage text-white shadow-lg shadow-sage/20",
    clay:      "bg-clay text-white shadow-lg shadow-clay/20",
    secondary: "bg-white text-charcoal border border-borderSoft hover:bg-sidebar",
    ghost:     "text-muted hover:bg-sidebar hover:text-charcoal shadow-none",
    danger:    "bg-red-500 text-white shadow-lg shadow-red-500/20",
  };

  const sizes = {
    sm: "px-4 py-2 text-[12px]",
    md: "px-6 py-2.5 text-[13px]",
    lg: "px-8 py-3.5 text-[15px]",
  };
</script>

<button
  {type}
  disabled={disabled || loading}
  class="{base} {variants[variant === 'primary' ? 'sage' : variant]} {sizes[size]} {fullWidth ? 'w-full' : ''}"
  on:click
>
  {#if loading}
    <!-- Minimal spinner -->
    <iconify-icon icon="lucide:loader-2" class="animate-spin text-lg"></iconify-icon>
  {/if}
  <slot />
</button>
