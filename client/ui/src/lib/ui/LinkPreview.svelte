<script>
  export let url = "";

  let domain = "";
  let displayUrl = "";

  $: {
    try {
      const u = new URL(url);
      domain = u.hostname.replace(/^www\./, "");
      displayUrl = u.hostname + u.pathname.slice(0, 40) + (u.pathname.length > 40 ? "..." : "");
    } catch {
      domain = url;
      displayUrl = url;
    }
  }
</script>

<a href={url} target="_blank" rel="noopener noreferrer"
  class="mt-1.5 flex items-center gap-2.5 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 transition hover:bg-slate-100 group max-w-md">
  <div class="grid h-8 w-8 flex-shrink-0 place-content-center rounded-md bg-slate-200 text-slate-500">
    <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
    </svg>
  </div>
  <div class="min-w-0 flex-1">
    <p class="text-xs font-semibold text-slate-700 truncate group-hover:text-blue-600">{displayUrl}</p>
    <p class="text-xs text-slate-400">{domain}</p>
  </div>
  <svg class="h-3.5 w-3.5 flex-shrink-0 text-slate-300 group-hover:text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
  </svg>
</a>
