<script>
  export let variant = "primary";
  export let type = "button";
  export let fullWidth = false;
  export let disabled = false;
</script>

<button {type} {disabled} class={`rounded-xl px-4 py-2 text-sm font-medium transition disabled:opacity-50
  ${fullWidth ? "w-full" : ""}
  ${variant === "primary"
    ? "bg-gradient-to-br from-shell-accent to-shell-success text-white hover:opacity-90"
    : "border border-shell-line bg-white text-shell-ink hover:bg-shell-bg"}`}
  on:click
>
  <slot />
</button>
