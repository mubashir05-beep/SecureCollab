<script>
  export let url = "";

  let domain = "";
  let displayUrl = "";

  $: {
    try {
      const u = new URL(url);
      domain = u.hostname.replace(/^www\./, "");
      const path = u.pathname.slice(0, 48);
      displayUrl = u.hostname + path + (u.pathname.length > 48 ? "…" : "");
    } catch {
      domain = url;
      displayUrl = url;
    }
  }
</script>

<a
  href={url}
  target="_blank"
  rel="noopener noreferrer"
  class="group mt-1.5 flex max-w-sm items-center gap-3 rounded-lg border border-shell-border bg-shell-elevated px-3 py-2 transition-colors hover:border-shell-accent hover:bg-shell-surface"
  aria-label="Link preview: {displayUrl}"
>
  <!-- Link icon -->
  <div class="grid h-8 w-8 flex-shrink-0 place-content-center rounded-md bg-shell-surface text-shell-subtle">
    <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
    </svg>
  </div>

  <!-- Text -->
  <div class="min-w-0 flex-1">
    <p class="truncate text-xs font-medium text-shell-ink group-hover:text-shell-accentText">{displayUrl}</p>
    <p class="text-xs text-shell-subtle">{domain}</p>
  </div>

  <!-- External arrow -->
  <svg class="h-3.5 w-3.5 flex-shrink-0 text-shell-subtle opacity-0 group-hover:opacity-100 transition-opacity" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
    <path stroke-linecap="round" stroke-linejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
  </svg>
</a>
