<script>
  export let padded = true;
  export let elevated = false;
</script>

<section class="rounded-xl border border-shell-border {elevated ? 'bg-shell-elevated' : 'bg-shell-surface'} {padded ? 'p-4' : ''}">
  <slot />
</section>
