<script>
  import { createEventDispatcher } from "svelte";

  export let channelName = "";
  export let channelTopic = "";
  export let memberCount = 0;
  export let isEncrypted = true;

  const dispatch = createEventDispatcher();
</script>

<header class="flex h-12 items-center justify-between border-b border-slate-200 bg-white px-4">
  <div class="flex items-center gap-3">
    <h1 class="text-base font-bold text-slate-900">
      {#if channelName}
        <span class="text-slate-400 font-normal">#</span> {channelName}
      {:else}
        SecureCollab
      {/if}
    </h1>
    {#if channelTopic}
      <span class="hidden text-sm text-slate-400 sm:inline">|</span>
      <span class="hidden truncate text-sm text-slate-500 sm:inline">{channelTopic}</span>
    {/if}
  </div>

  <div class="flex items-center gap-3">
    {#if isEncrypted}
      <span class="flex items-center gap-1 rounded-md bg-green-50 px-2 py-0.5 text-xs font-medium text-green-700">
        <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
        </svg>
        E2E Encrypted
      </span>
    {/if}
    {#if memberCount > 0}
      <button
        on:click={() => dispatch("showMembers")}
        class="flex items-center gap-1 text-sm text-slate-500 hover:text-slate-700"
      >
        <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
        </svg>
        {memberCount}
      </button>
    {/if}
    <button
      on:click={() => dispatch("search")}
      class="rounded-md p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
      title="Search"
    >
      <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
    </button>
  </div>
</header>
