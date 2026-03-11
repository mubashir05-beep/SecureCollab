<script>
  export let padded = true;
</script>

<section class={`rounded-2xl border border-shell-line bg-shell-panel ${padded ? "p-4" : ""}`}>
  <slot />
</section>
