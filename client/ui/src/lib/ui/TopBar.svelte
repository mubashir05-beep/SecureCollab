<script>
  import { createEventDispatcher } from "svelte";

  export let channelName = "";
  export let channelTopic = "";
  export let memberCount = 0;
  export let isEncrypted = true;

  const dispatch = createEventDispatcher();
</script>

<header
  class="flex h-12 flex-shrink-0 items-center justify-between border-b border-shell-borderSub bg-shell-bg px-4"
  aria-label="Channel header"
>
  <!-- Left: channel identity -->
  <div class="flex min-w-0 items-center gap-3">
    {#if channelName}
      <div class="flex min-w-0 items-center gap-2">
        <span class="flex-shrink-0 text-shell-muted text-base font-normal" aria-hidden="true">#</span>
        <h1 class="truncate text-sm font-bold text-shell-ink">{channelName}</h1>
      </div>
      {#if channelTopic}
        <span class="hidden h-4 w-px flex-shrink-0 bg-shell-border sm:block" aria-hidden="true"></span>
        <p class="hidden truncate text-sm text-shell-muted sm:block" title={channelTopic}>{channelTopic}</p>
      {/if}
    {:else}
      <h1 class="text-sm font-bold text-shell-ink">SecureCollab</h1>
    {/if}
  </div>

  <!-- Right: actions -->
  <div class="flex flex-shrink-0 items-center gap-1">
    <!-- E2E badge -->
    {#if isEncrypted}
      <span
        class="flex items-center gap-1 rounded-md bg-shell-success/10 px-2 py-0.5 text-xs font-medium text-shell-success"
        title="Messages are end-to-end encrypted"
      >
        <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
          <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
        </svg>
        E2E
      </span>
    {/if}

    <!-- Members count -->
    {#if memberCount > 0}
      <button
        on:click={() => dispatch("showMembers")}
        title="View members"
        aria-label="Show {memberCount} members"
        class="flex items-center gap-1.5 rounded-md px-2 py-1 text-sm text-shell-muted transition-colors hover:bg-shell-surface hover:text-shell-ink"
      >
        <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197" />
        </svg>
        <span>{memberCount}</span>
      </button>
    {/if}

    <!-- Search -->
    <button
      on:click={() => dispatch("search")}
      title="Search (coming soon)"
      aria-label="Search"
      class="rounded-md p-1.5 text-shell-muted transition-colors hover:bg-shell-surface hover:text-shell-ink"
    >
      <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
    </button>
  </div>
</header>
