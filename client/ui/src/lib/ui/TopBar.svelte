<script>
  import { createEventDispatcher } from "svelte";

  export let channelName = "";
  export let channelTopic = "";
  export let memberCount = 0;
  export let isEncrypted = true;
  export let connectionState = "connected";
  export let liveActivity = "";

  const dispatch = createEventDispatcher();

  const connectionStyles = {
    connected: { label: "Live", dot: "bg-sage", tone: "text-sage bg-sage/10" },
    connecting: { label: "Connecting", dot: "bg-clay", tone: "text-clay bg-clay/10" },
    reconnecting: { label: "Reconnecting", dot: "bg-clay", tone: "text-clay bg-clay/10" },
    disconnected: { label: "Offline", dot: "bg-muted", tone: "text-muted bg-muted/10" },
  };

  $: liveState = connectionStyles[connectionState] || connectionStyles.connected;
</script>

<header
  class="flex h-16 flex-shrink-0 items-center justify-between border-b border-borderSoft/50 glass-header px-6 z-10"
  aria-label="Channel header"
>
  <div class="min-w-0 flex items-center gap-4">
    <div class="flex flex-col">
      <div class="flex items-center gap-2">
        <span class="inline-flex h-2 w-2 rounded-full {liveState.dot}" aria-hidden="true"></span>
        <span class="text-[10px] font-bold uppercase tracking-widest text-muted">{liveState.label}</span>
        {#if liveActivity}
          <span class="text-[10px] font-bold text-muted/60">{liveActivity}</span>
        {/if}
      </div>

      <div class="flex items-center gap-2 mt-0.5">
        {#if channelName}
          <h1 class="text-[15px] font-bold text-charcoal truncate">{channelName}</h1>
          {#if channelTopic}
            <span class="text-borderSoft">|</span>
            <p class="truncate text-[13px] text-muted font-medium" title={channelTopic}>{channelTopic}</p>
          {/if}
        {:else}
          <h1 class="text-[15px] font-bold text-charcoal">SecureCollab Workspace</h1>
        {/if}
      </div>
    </div>
  </div>

  <!-- Right: actions -->
  <div class="flex flex-shrink-0 items-center gap-3">
    <!-- E2E badge -->
    {#if isEncrypted}
      <div
        class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-sage/10 text-sage"
        title="Messages are end-to-end encrypted"
      >
        <iconify-icon icon="lucide:shield-check" class="text-sm"></iconify-icon>
        <span class="text-[11px] font-bold uppercase tracking-tight">Protected</span>
      </div>
    {/if}

    <!-- Members count -->
    {#if memberCount > 0}
      <button
        on:click={() => dispatch("showMembers")}
        title="View members"
        class="flex items-center gap-2 px-3 py-1.5 rounded-xl border border-borderSoft bg-white/50 text-muted hover:border-sage hover:text-sage transition-all"
      >
        <iconify-icon icon="lucide:users" class="text-sm"></iconify-icon>
        <span class="text-[12px] font-bold">{memberCount}</span>
      </button>
    {/if}

    <div class="w-px h-6 bg-borderSoft/60 mx-1"></div>

    <!-- Search -->
    <button
      on:click={() => dispatch("search")}
      class="w-9 h-9 flex items-center justify-center rounded-xl border border-borderSoft bg-white/50 text-muted hover:border-sage hover:text-sage transition-all"
    >
      <iconify-icon icon="lucide:search" class="text-lg"></iconify-icon>
    </button>

    <!-- Productivity Toggle (can trigger task panel) -->
    <button
      on:click={() => dispatch("toggleTasks")}
      class="w-9 h-9 flex items-center justify-center rounded-xl bg-clay text-white shadow-lg shadow-clay/20 hover:scale-[1.05] active:scale-95 transition-all"
    >
      <iconify-icon icon="lucide:check-square" class="text-lg"></iconify-icon>
    </button>
  </div>
</header>
