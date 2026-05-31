<script>
  import { createEventDispatcher } from "svelte";
  import Button from "./Button.svelte";

  export let visible = false;
  export let members = [];
  export let currentUserId = "";
  export let isAdmin = false;

  let addUserId = "";
  let addRole = "member";
  let error = "";

  const dispatch = createEventDispatcher();

  const roleBadge = {
    owner:  "bg-clay/10 text-clay border-clay/20",
    admin:  "bg-sage/10 text-sage border-sage/20",
    member: "bg-sidebar text-muted border-borderSoft",
    viewer: "bg-sidebar text-muted/60 border-borderSoft",
  };

  // Premium Avatar Palette
  const avatarPalette = [
    "bg-sage", "bg-clay", "bg-charcoal",
    "bg-stone-400", "bg-[#A8A29E]", "bg-[#78716C]",
    "bg-[#44403C]", "bg-[#1C1917]",
  ];
  function avatarBg(name = "") {
    return avatarPalette[(name.charCodeAt(0) || 0) % avatarPalette.length];
  }

  function handleAdd() {
    if (!addUserId.trim()) return;
    error = "";
    dispatch("addMember", { userId: addUserId.trim(), role: addRole });
    addUserId = "";
  }

  function handleRemove(userId) {
    dispatch("removeMember", { userId });
  }

  export function setError(msg) { error = msg; }
</script>

{#if visible}
  <aside
    class="w-80 h-full flex flex-col border-l border-borderSoft/50 bg-sidebar overflow-hidden animate-fade-in"
    aria-label="Members panel"
  >
    <!-- Header -->
    <div class="h-[72px] flex items-center justify-between px-6 border-b border-borderSoft/50">
      <div class="flex flex-col">
        <h3 class="text-[15px] font-bold text-charcoal tracking-tight">Workspace Members</h3>
        <span class="text-[11px] font-bold text-muted/40 uppercase tracking-widest">
          {members.length} {members.length === 1 ? "User" : "Users"} Online
        </span>
      </div>
      <button
        on:click={() => dispatch("close")}
        class="w-10 h-10 flex items-center justify-center rounded-xl hover:bg-white text-muted hover:text-clay transition-all"
      >
        <iconify-icon icon="lucide:x" class="text-xl"></iconify-icon>
      </button>
    </div>

    <!-- Add member section (admin only) -->
    {#if isAdmin}
      <div class="p-6 border-b border-borderSoft/30 bg-white/40">
        <h4 class="text-[11px] font-bold text-muted uppercase tracking-widest mb-4">Add Team Member</h4>
        {#if error}
          <div class="mb-4 p-3 rounded-xl bg-red-50 text-[12px] font-bold text-red-500 border border-red-100">
            {error}
          </div>
        {/if}
        <div class="space-y-3">
          <input
            type="text"
            bind:value={addUserId}
            placeholder="Search by ID or email..."
            class="w-full px-4 py-3 rounded-xl border border-borderSoft bg-white text-[13px] font-medium focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all outline-none"
            on:keydown={(e) => e.key === "Enter" && handleAdd()}
          />
          <div class="flex gap-2">
            <select
              bind:value={addRole}
              class="flex-1 px-4 py-3 rounded-xl border border-borderSoft bg-white text-[13px] font-medium outline-none focus:border-sage transition-all"
            >
              <option value="member">Member</option>
              <option value="admin">Admin</option>
              <option value="viewer">Viewer</option>
            </select>
            <Button variant="sage" on:click={handleAdd} disabled={!addUserId.trim()}>
              <iconify-icon icon="lucide:plus" class="text-lg"></iconify-icon>
            </Button>
          </div>
        </div>
      </div>
    {/if}

    <!-- Member list -->
    <div class="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-1">
      {#each members as member (member.user_id)}
        {@const displayName = member.username || member.user_id}
        {@const isCurrentUser = member.user_id === currentUserId}
        <div
          class="group flex items-center gap-3 p-3 rounded-2xl hover:bg-white hover:shadow-sm transition-all"
        >
          <!-- Avatar -->
          <div
            class="w-10 h-10 flex items-center justify-center rounded-2xl text-xs font-bold text-white shadow-sm {avatarBg(displayName)} transition-transform group-hover:scale-105"
          >
            {displayName.charAt(0).toUpperCase()}
          </div>

          <!-- Name + role -->
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="truncate text-[14px] font-bold text-charcoal">
                {displayName}
              </span>
              {#if isCurrentUser}
                <span class="text-[10px] font-bold text-sage bg-sage/10 px-1.5 py-0.5 rounded-md uppercase tracking-tight">You</span>
              {/if}
            </div>
            <div class="flex items-center gap-1.5 mt-0.5">
              <span class="inline-flex items-center px-2 py-0.5 rounded-lg border text-[10px] font-bold uppercase tracking-tight {roleBadge[member.role] || roleBadge.member}">
                {member.role || "member"}
              </span>
            </div>
          </div>

          <!-- Actions -->
          {#if isAdmin && !isCurrentUser && member.role !== "owner"}
            <button
              on:click={() => handleRemove(member.user_id)}
              class="w-8 h-8 flex items-center justify-center rounded-lg text-muted opacity-0 group-hover:opacity-100 hover:bg-red-50 hover:text-red-500 transition-all"
              title="Remove member"
            >
              <iconify-icon icon="lucide:user-minus" class="text-lg"></iconify-icon>
            </button>
          {/if}
        </div>
      {:else}
        <div class="p-12 text-center space-y-3">
          <iconify-icon icon="lucide:users" class="text-4xl text-muted/20"></iconify-icon>
          <p class="text-[13px] font-bold text-muted/40 uppercase tracking-widest">No members found</p>
        </div>
      {/each}
    </div>
  </aside>
{/if}
