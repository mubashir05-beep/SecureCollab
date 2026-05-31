<script>
  import { createEventDispatcher } from "svelte";

  export let workspaces = [];
  export let channels = [];
  export let directMessages = [];
  export let activeWorkspace = null;
  export let activeChannel = null;
  export let currentUser = null;

  const dispatch = createEventDispatcher();
</script>

<aside class="flex h-screen flex-shrink-0 select-none sidebar-transition bg-sidebar border-r border-borderSoft" aria-label="Sidebar">
  <!-- ── Workspace Rail (Icon-only rail) ── -->
  <div class="flex w-[70px] flex-col items-center gap-4 py-6 border-r border-borderSoft/50 bg-sidebar/50">
    <!-- App Logo -->
    <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-sage text-white shadow-lg shadow-sage/20 font-bold text-lg mb-2">
      SC
    </div>

    <div class="flex-1 w-full flex flex-col items-center gap-3 overflow-y-auto custom-scrollbar px-2">
      {#each workspaces as ws}
        <button
          on:click={() => dispatch("selectWorkspace", ws)}
          title={ws.name}
          class="group relative flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-2xl transition-all duration-300
            {activeWorkspace?.id === ws.id
              ? 'bg-white text-sage shadow-md'
              : 'text-muted hover:bg-white/60 hover:text-charcoal'}"
        >
          <span class="text-sm font-bold">{ws.name?.charAt(0)?.toUpperCase() || "?"}</span>
          
          {#if activeWorkspace?.id === ws.id}
            <div class="absolute -left-2 w-1.5 h-6 bg-sage rounded-r-full"></div>
          {/if}

          {#if ws.unreadCount > 0}
            <span class="absolute -right-1 -top-1 flex h-4.5 min-w-[18px] items-center justify-center rounded-full bg-clay text-[10px] font-bold text-white shadow-sm px-1 border-2 border-sidebar">
              {ws.unreadCount > 9 ? "9+" : ws.unreadCount}
            </span>
          {/if}
        </button>
      {/each}

      <!-- Add Workspace -->
      <button
        on:click={() => dispatch("createWorkspace")}
        class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-2xl border-2 border-dashed border-borderSoft text-muted transition-all hover:border-sage hover:text-sage hover:bg-white/40"
      >
        <iconify-icon icon="lucide:plus" class="text-xl"></iconify-icon>
      </button>
    </div>

    <!-- User Profile & Settings -->
    {#if currentUser}
      <div class="flex flex-col items-center gap-4 mt-auto">
        <button
          on:click={() => dispatch("logout")}
          class="text-muted hover:text-clay transition-colors"
          title="Sign out"
        >
          <iconify-icon icon="lucide:log-out" class="text-xl"></iconify-icon>
        </button>
        
        <div class="relative">
          <div class="w-10 h-10 rounded-2xl bg-white shadow-sm flex items-center justify-center border border-borderSoft group overflow-hidden cursor-pointer hover:border-sage transition-colors">
            <span class="text-sm font-bold text-charcoal">{currentUser.username?.charAt(0)?.toUpperCase()}</span>
          </div>
          <span class="absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 rounded-full border-2 border-sidebar bg-sage"></span>
        </div>
      </div>
    {/if}
  </div>

  <!-- ── Channel/Project Panel ── -->
  <div class="flex w-[210px] flex-col overflow-hidden">
    <!-- Header -->
    <div class="h-[72px] flex flex-col justify-center px-6 border-b border-borderSoft/50">
      <div class="flex items-center justify-between group cursor-pointer">
        <div class="min-w-0">
          <h2 class="truncate text-[15px] font-bold text-charcoal leading-tight">
            {activeWorkspace?.name || "Select Workspace"}
          </h2>
          <div class="flex items-center gap-1">
            <span class="w-1.5 h-1.5 rounded-full bg-sage"></span>
            <span class="text-[11px] font-semibold text-muted uppercase tracking-tight">Active Team</span>
          </div>
        </div>
        <iconify-icon icon="lucide:chevron-down" class="text-muted opacity-0 group-hover:opacity-100 transition-opacity"></iconify-icon>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto custom-scrollbar py-6 px-3">
      <!-- Navigation Sections -->
      <div class="space-y-6">
        <!-- Main Nav -->
        <div class="space-y-1">
          <button class="flex w-full items-center gap-2.5 px-3 py-2 rounded-xl text-[13px] font-semibold bg-white text-sage shadow-sm transition-all">
            <iconify-icon icon="lucide:message-square" class="text-lg"></iconify-icon>
            <span>Team Chat</span>
          </button>
          <button class="flex w-full items-center gap-2.5 px-3 py-2 rounded-xl text-[13px] font-medium text-muted hover:bg-white/60 transition-all">
            <iconify-icon icon="lucide:sparkles" class="text-lg opacity-60"></iconify-icon>
            <span>AI Assistant</span>
          </button>
        </div>

        <!-- Channels -->
        <div class="space-y-2">
          <div class="flex items-center justify-between px-3 mb-1">
            <span class="text-[11px] font-bold uppercase tracking-widest text-muted/60">Channels</span>
            <button
              on:click={() => dispatch("createChannel")}
              class="w-5 h-5 flex items-center justify-center rounded-lg text-muted hover:bg-white hover:text-sage transition-all"
            >
              <iconify-icon icon="lucide:plus" class="text-sm"></iconify-icon>
            </button>
          </div>

          <div class="space-y-0.5">
            {#each channels as ch}
              <button
                on:click={() => dispatch("selectChannel", ch)}
                class="group flex w-full items-center gap-2.5 px-3 py-2 rounded-xl text-left transition-all
                  {activeChannel?.id === ch.id
                    ? 'bg-white text-sage shadow-sm font-semibold'
                    : 'text-muted hover:bg-white/40 hover:text-charcoal'}"
              >
                <iconify-icon
                  icon={ch.is_private ? "lucide:lock" : "lucide:hash"}
                  class="text-lg {activeChannel?.id === ch.id ? 'text-sage' : 'opacity-40'}"
                ></iconify-icon>
                <span class="flex-1 truncate text-[13px]">{ch.name}</span>
                {#if ch.unreadCount > 0}
                  <span class="h-5 min-w-[20px] flex items-center justify-center rounded-full bg-clay text-[10px] font-bold text-white px-1">
                    {ch.unreadCount}
                  </span>
                {/if}
              </button>
            {/each}
          </div>
        </div>

        <!-- Direct Messages -->
        {#if directMessages.length > 0}
          <div class="space-y-2">
            <div class="px-3 mb-1">
              <span class="text-[11px] font-bold uppercase tracking-widest text-muted/60">Collaborators</span>
            </div>
            <div class="space-y-0.5">
              {#each directMessages as dm}
                <button
                  on:click={() => dispatch("selectChannel", dm)}
                  class="group flex w-full items-center gap-2.5 px-3 py-2 rounded-xl text-left transition-all
                    {activeChannel?.id === dm.id
                      ? 'bg-white text-sage shadow-sm font-semibold'
                      : 'text-muted hover:bg-white/40 hover:text-charcoal'}"
                >
                  <div class="relative flex-shrink-0">
                    <div class="w-6 h-6 rounded-lg bg-white border border-borderSoft flex items-center justify-center text-[10px] font-bold text-charcoal">
                      {dm.name?.charAt(0)}
                    </div>
                    <span class="absolute -bottom-0.5 -right-0.5 h-2 w-2 rounded-full border border-sidebar {dm.online ? 'bg-sage' : 'bg-muted/40'}"></span>
                  </div>
                  <span class="flex-1 truncate text-[13px]">{dm.name}</span>
                </button>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    </div>

    <!-- Invite Button at bottom of panel -->
    <div class="p-4 border-t border-borderSoft/30">
      <button
        on:click={() => dispatch("invite")}
        class="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl border border-borderSoft text-muted text-[12px] font-semibold hover:border-sage hover:text-sage hover:bg-white transition-all"
      >
        <iconify-icon icon="lucide:user-plus" class="text-lg"></iconify-icon>
        Invite Teammates
      </button>
    </div>
  </div>
</aside>
