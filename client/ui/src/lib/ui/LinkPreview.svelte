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
  class="group mt-2 flex max-w-md items-center gap-4 p-3 rounded-2xl border border-borderSoft bg-white hover:bg-sidebar/30 hover:border-sage/30 hover:shadow-lg hover:shadow-stone-200/40 transition-all duration-200"
>
  <!-- Link icon -->
  <div class="w-10 h-10 flex-shrink-0 flex items-center justify-center rounded-xl bg-sidebar text-muted/60 group-hover:bg-sage/10 group-hover:text-sage transition-all">
    <iconify-icon icon="lucide:link" class="text-xl"></iconify-icon>
  </div>

  <!-- Text -->
  <div class="min-w-0 flex-1">
    <p class="truncate text-[13px] font-bold text-charcoal leading-none mb-1 group-hover:text-sage transition-colors">{displayUrl}</p>
    <p class="text-[11px] font-bold text-muted/40 uppercase tracking-widest">{domain}</p>
  </div>

  <!-- External arrow -->
  <div class="w-8 h-8 flex items-center justify-center rounded-lg text-muted/30 opacity-0 group-hover:opacity-100 group-hover:translate-x-1 transition-all">
    <iconify-icon icon="lucide:external-link" class="text-lg"></iconify-icon>
  </div>
</a>
