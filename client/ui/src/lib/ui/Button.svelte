<script>
  /** @type {"primary" | "secondary" | "ghost" | "danger"} */
  export let variant = "primary";
  /** @type {"sm" | "md" | "lg"} */
  export let size = "md";
  export let type = "button";
  export let fullWidth = false;
  export let disabled = false;
  export let loading = false;

  const base = "inline-flex items-center justify-center gap-2 font-medium rounded-lg transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-shell-accent focus-visible:ring-offset-2 focus-visible:ring-offset-shell-sidebar disabled:opacity-40 disabled:cursor-not-allowed select-none";

  const variants = {
    primary:   "bg-shell-accent text-white hover:bg-shell-accentHov active:bg-shell-accentHov",
    secondary: "bg-shell-elevated text-shell-ink border border-shell-border hover:bg-shell-surface",
    ghost:     "text-shell-muted hover:bg-shell-surface hover:text-shell-ink",
    danger:    "bg-shell-danger text-white hover:opacity-90 active:opacity-80",
  };

  const sizes = {
    sm: "px-3 py-1.5 text-xs h-7",
    md: "px-4 py-2 text-sm h-8",
    lg: "px-5 py-2.5 text-sm h-10",
  };
</script>

<button
  {type}
  disabled={disabled || loading}
  class="{base} {variants[variant]} {sizes[size]} {fullWidth ? 'w-full' : ''}"
  on:click
>
  {#if loading}
    <!-- Minimal spinner -->
    <svg class="h-3.5 w-3.5 animate-spin opacity-70" viewBox="0 0 24 24" fill="none">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
    </svg>
  {/if}
  <slot />
</button>
