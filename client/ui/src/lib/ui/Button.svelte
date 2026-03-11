<script>
  export let variant = "primary";
  export let type = "button";
</script>

<button type={type} class={`rounded-xl px-4 py-2 text-sm font-medium transition ${variant === "primary"
  ? "bg-gradient-to-br from-shell-accent to-shell-success text-white hover:opacity-90"
  : "border border-shell-line bg-white text-shell-ink hover:bg-shell-bg"}`}>
  <slot />
</button>
