<script>
  import { createEventDispatcher } from "svelte";

  export let workspaces = [];
  export let channels = [];
  export let directMessages = [];
  export let activeWorkspace = null;
  export let activeChannel = null;
  export let currentUser = null;

  const dispatch = createEventDispatcher();

  // Avatar color palette — deterministic per workspace initial
  const railColors = [
    "bg-indigo-600", "bg-violet-600", "bg-blue-600",
    "bg-teal-600",   "bg-rose-600",   "bg-amber-600",
  ];

  function railColor(name = "") {
    const idx = (name.charCodeAt(0) || 0) % railColors.length;
    return railColors[idx];
  }
</script>

<aside class="flex h-screen flex-shrink-0 select-none" aria-label="Sidebar">

  <!-- ── Workspace rail (64px) ── -->
  <div class="flex w-16 flex-col items-center gap-1.5 overflow-y-auto bg-shell-sidebar py-3">

    {#each workspaces as ws}
      <button
        on:click={() => dispatch("selectWorkspace", ws)}
        title={ws.name}
        aria-label="Switch to workspace: {ws.name}"
        class="relative flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-2xl text-sm font-bold transition-all duration-150
          {activeWorkspace?.id === ws.id
            ? 'rounded-xl bg-shell-accent text-white shadow-panel'
            : 'bg-white/10 text-white/80 hover:rounded-xl hover:bg-shell-accent hover:text-white'}"
      >
        {ws.name?.charAt(0)?.toUpperCase() || "?"}
        {#if ws.unreadCount > 0}
          <span class="absolute -right-0.5 -top-0.5 h-3.5 w-3.5 rounded-full border-2 border-shell-sidebar bg-shell-danger text-[0.5rem] font-bold text-white flex items-center justify-center">
            {ws.unreadCount > 9 ? "9+" : ws.unreadCount}
          </span>
        {/if}
      </button>
    {/each}

    <!-- Create workspace -->
    <button
      on:click={() => dispatch("createWorkspace")}
      title="Create a new workspace"
      aria-label="Create workspace"
      class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-2xl border-2 border-dashed border-white/20 text-xl text-white/40 transition-all duration-150 hover:rounded-xl hover:border-shell-accent hover:text-shell-accent"
    >+</button>

    <!-- Bottom controls: invite + user avatar + logout -->
    {#if currentUser}
      <div class="mt-auto flex flex-col items-center gap-1.5 pb-2 pt-4">
        <button
          on:click={() => dispatch("invite")}
          title="Invite to workspace"
          aria-label="Invite members"
          class="flex h-9 w-9 items-center justify-center rounded-xl text-white/40 transition-colors hover:bg-white/10 hover:text-white/80"
        >
          <svg class="h-4.5 w-4.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
          </svg>
        </button>

        <!-- User avatar -->
        <div
          class="relative flex h-9 w-9 items-center justify-center rounded-full bg-shell-accent text-xs font-bold text-white"
          title={currentUser.username}
          aria-label="Logged in as {currentUser.username}"
        >
          {currentUser.username?.charAt(0)?.toUpperCase() || "U"}
          <!-- Online dot -->
          <span class="absolute -bottom-0 -right-0 h-3 w-3 rounded-full border-2 border-shell-sidebar bg-shell-success"></span>
        </div>

        <button
          on:click={() => dispatch("logout")}
          title="Sign out"
          aria-label="Sign out"
          class="flex h-9 w-9 items-center justify-center rounded-xl text-white/40 transition-colors hover:bg-shell-danger/20 hover:text-shell-danger"
        >
          <svg class="h-4.5 w-4.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </button>
      </div>
    {/if}
  </div>

  <!-- ── Channel panel (192px) ── -->
  <nav class="flex w-48 flex-col bg-shell-panel" aria-label="Channels">

    <!-- Workspace name header -->
    <div class="flex h-12 items-center justify-between border-b border-shell-borderSub px-3">
      <div class="min-w-0">
        <h2 class="truncate text-sm font-bold text-shell-ink leading-tight">
          {activeWorkspace?.name || "SecureCollab"}
        </h2>
        {#if currentUser}
          <p class="truncate text-xs text-shell-subtle leading-tight">{currentUser.username}</p>
        {/if}
      </div>
      <svg class="ml-1 h-4 w-4 flex-shrink-0 text-shell-subtle" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
      </svg>
    </div>

    <!-- Scrollable channel list -->
    <div class="flex-1 overflow-y-auto py-2">

      <!-- Channels section -->
      <div class="mb-1">
        <div class="flex items-center justify-between px-3 pb-1 pt-0.5">
          <span class="text-[11px] font-semibold uppercase tracking-widest text-shell-subtle">Channels</span>
          <button
            on:click={() => dispatch("createChannel")}
            title="Add channel"
            aria-label="Create channel"
            class="flex h-5 w-5 items-center justify-center rounded text-shell-subtle transition-colors hover:bg-shell-elevated hover:text-shell-ink text-base leading-none"
          >+</button>
        </div>

        {#each channels as ch}
          <button
            on:click={() => dispatch("selectChannel", ch)}
            aria-current={activeChannel?.id === ch.id ? "page" : undefined}
            class="group flex w-full items-center gap-1.5 rounded-md px-3 py-1 text-left transition-colors duration-100
              {activeChannel?.id === ch.id
                ? 'bg-white/10 text-white'
                : 'text-shell-muted hover:bg-white/5 hover:text-shell-ink'}"
          >
            <span class="flex-shrink-0 text-sm {activeChannel?.id === ch.id ? 'text-white/70' : 'text-shell-subtle'}">
              {ch.is_private ? "🔒" : "#"}
            </span>
            <span class="flex-1 truncate text-sm">{ch.name}</span>
            {#if ch.unreadCount > 0}
              <span class="flex h-4 min-w-4 items-center justify-center rounded-full bg-shell-danger px-1 text-[10px] font-bold text-white">
                {ch.unreadCount > 99 ? "99+" : ch.unreadCount}
              </span>
            {/if}
          </button>
        {/each}
      </div>

      <!-- Direct Messages section -->
      {#if directMessages.length > 0}
        <div class="mt-4">
          <div class="px-3 pb-1 pt-0.5">
            <span class="text-[11px] font-semibold uppercase tracking-widest text-shell-subtle">Direct Messages</span>
          </div>

          {#each directMessages as dm}
            <button
              on:click={() => dispatch("selectChannel", dm)}
              aria-current={activeChannel?.id === dm.id ? "page" : undefined}
              class="flex w-full items-center gap-2 rounded-md px-3 py-1 text-left transition-colors duration-100
                {activeChannel?.id === dm.id
                  ? 'bg-white/10 text-white'
                  : 'text-shell-muted hover:bg-white/5 hover:text-shell-ink'}"
            >
              <span class="h-2 w-2 flex-shrink-0 rounded-full {dm.online ? 'bg-shell-success' : 'bg-shell-subtle'}"></span>
              <span class="flex-1 truncate text-sm">{dm.name}</span>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </nav>
</aside>
