<script>
  import { createEventDispatcher } from "svelte";

  export let workspaces = [];
  export let channels = [];
  export let directMessages = [];
  export let activeWorkspace = null;
  export let activeChannel = null;
  export let currentUser = null;
  export let collapsed = false;

  const dispatch = createEventDispatcher();

  function selectWorkspace(ws) {
    dispatch("selectWorkspace", ws);
  }

  function selectChannel(ch) {
    dispatch("selectChannel", ch);
  }

  function createWorkspace() {
    dispatch("createWorkspace");
  }

  function createChannel() {
    dispatch("createChannel");
  }
</script>

<aside class="flex h-screen flex-shrink-0 {collapsed ? 'w-16' : 'w-64'} transition-all duration-200">
  <!-- Workspace rail -->
  <div class="flex w-16 flex-col items-center gap-2 bg-slate-900 py-3">
    {#each workspaces as ws}
      <button
        on:click={() => selectWorkspace(ws)}
        class="group relative grid h-10 w-10 place-content-center rounded-xl text-sm font-bold transition-all
          {activeWorkspace?.id === ws.id
            ? 'bg-shell-accent text-white rounded-xl'
            : 'bg-slate-700 text-slate-300 rounded-2xl hover:rounded-xl hover:bg-shell-accent hover:text-white'}"
        title={ws.name}
      >
        {ws.name?.charAt(0)?.toUpperCase() || "?"}
        {#if ws.unreadCount > 0}
          <span class="absolute -right-0.5 -top-0.5 h-3 w-3 rounded-full bg-red-500 border-2 border-slate-900"></span>
        {/if}
      </button>
    {/each}

    <button
      on:click={createWorkspace}
      class="grid h-10 w-10 place-content-center rounded-2xl bg-slate-800 text-slate-400 transition hover:rounded-xl hover:bg-shell-accent hover:text-white"
      title="Create workspace"
    >
      +
    </button>

    <!-- Bottom: user avatar + logout -->
    {#if currentUser}
      <div class="mt-auto flex flex-col items-center gap-2 pb-2">
        <button
          on:click={() => dispatch("invite")}
          class="grid h-9 w-9 place-content-center rounded-xl bg-slate-800 text-slate-400 transition hover:bg-slate-700 hover:text-white"
          title="Invite to workspace"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
          </svg>
        </button>
        <div class="grid h-10 w-10 place-content-center rounded-full bg-slate-700 text-xs font-bold text-slate-300"
          title={currentUser.username}>
          {currentUser.username?.charAt(0)?.toUpperCase() || "U"}
        </div>
        <button
          on:click={() => dispatch("logout")}
          class="grid h-9 w-9 place-content-center rounded-xl bg-slate-800 text-slate-400 transition hover:bg-red-500/20 hover:text-red-400"
          title="Logout"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </button>
      </div>
    {/if}
  </div>

  <!-- Channel panel -->
  {#if !collapsed}
    <div class="flex w-48 flex-col bg-slate-800 text-slate-200">
      <!-- Workspace header -->
      <div class="flex items-center justify-between border-b border-slate-700 px-3 py-3">
        <div class="truncate">
          <h2 class="text-sm font-bold truncate">{activeWorkspace?.name || "SecureCollab"}</h2>
          <p class="text-xs text-slate-400 truncate">{currentUser?.username || ""}</p>
        </div>
      </div>

      <!-- Channel list -->
      <div class="flex-1 overflow-y-auto px-1 py-2">
        <div class="mb-1 flex items-center justify-between px-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-400">Channels</span>
          <button
            on:click={createChannel}
            class="text-slate-400 hover:text-white text-sm leading-none"
            title="Create channel"
          >+</button>
        </div>

        {#each channels as ch}
          <button
            on:click={() => selectChannel(ch)}
            class="flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-sm transition
              {activeChannel?.id === ch.id
                ? 'bg-shell-accent/20 text-white'
                : 'text-slate-300 hover:bg-slate-700'}"
          >
            <span class="text-slate-400">{ch.is_private ? "🔒" : "#"}</span>
            <span class="truncate">{ch.name}</span>
            {#if ch.unreadCount > 0}
              <span class="ml-auto flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-bold text-white">
                {ch.unreadCount}
              </span>
            {/if}
          </button>
        {/each}

        <!-- Direct Messages -->
        {#if directMessages.length > 0}
          <div class="mb-1 mt-4 flex items-center justify-between px-2">
            <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-400">Direct Messages</span>
          </div>

          {#each directMessages as dm}
            <button
              on:click={() => selectChannel(dm)}
              class="flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-sm transition
                {activeChannel?.id === dm.id
                  ? 'bg-shell-accent/20 text-white'
                  : 'text-slate-300 hover:bg-slate-700'}"
            >
              <span class="h-2 w-2 rounded-full {dm.online ? 'bg-green-400' : 'bg-slate-500'}"></span>
              <span class="truncate">{dm.name}</span>
            </button>
          {/each}
        {/if}
      </div>
    </div>
  {/if}
</aside>
