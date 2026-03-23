<script>
  import { createEventDispatcher } from "svelte";

  export let visible = false;
  export let members = [];
  export let currentUserId = "";
  export let isAdmin = false;

  let addUserId = "";
  let addRole = "member";
  let error = "";

  const dispatch = createEventDispatcher();

  const roleBadge = {
    owner:  "bg-amber-400/15 text-amber-400 border-amber-400/20",
    admin:  "bg-violet-400/15 text-violet-400 border-violet-400/20",
    member: "bg-shell-elevated text-shell-muted border-shell-border",
    viewer: "bg-shell-elevated text-shell-subtle border-shell-border",
  };

  // Avatar color deterministic from username
  const avatarPalette = [
    "bg-indigo-500", "bg-violet-500", "bg-sky-500",
    "bg-teal-500",   "bg-rose-500",   "bg-amber-500",
    "bg-green-600",  "bg-pink-500",
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
    class="flex w-72 flex-shrink-0 flex-col border-l border-shell-borderSub bg-shell-bg animate-fade-in"
    aria-label="Members panel"
  >
    <!-- Header -->
    <div class="flex h-12 flex-shrink-0 items-center justify-between border-b border-shell-borderSub px-4">
      <div>
        <h3 class="text-sm font-bold text-shell-ink">Members</h3>
        <p class="text-xs text-shell-subtle">
          {members.length} {members.length === 1 ? "member" : "members"}
        </p>
      </div>
      <button
        on:click={() => dispatch("close")}
        class="rounded-md p-1.5 text-shell-subtle transition-colors hover:bg-shell-surface hover:text-shell-ink"
        aria-label="Close members panel"
      >
        <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Add member form (admin only) -->
    {#if isAdmin}
      <div class="border-b border-shell-borderSub px-4 py-3">
        <p class="mb-2 text-xs font-semibold uppercase tracking-wider text-shell-subtle">Add member</p>
        {#if error}
          <p class="mb-2 rounded-md bg-shell-dangerBg px-2 py-1 text-xs text-shell-danger">{error}</p>
        {/if}
        <div class="flex gap-1.5">
          <input
            type="text"
            bind:value={addUserId}
            placeholder="Username"
            class="flex-1 rounded-lg border border-shell-border bg-shell-elevated px-2.5 py-1.5 text-xs text-shell-ink placeholder-shell-subtle outline-none transition-colors focus:border-shell-accent"
            on:keydown={(e) => e.key === "Enter" && handleAdd()}
            aria-label="User ID or username to add"
          />
          <select
            bind:value={addRole}
            class="rounded-lg border border-shell-border bg-shell-elevated px-1.5 py-1.5 text-xs text-shell-muted outline-none focus:border-shell-accent"
            aria-label="Role"
          >
            <option value="member">Member</option>
            <option value="admin">Admin</option>
            <option value="viewer">Viewer</option>
          </select>
        </div>
        <button
          on:click={handleAdd}
          disabled={!addUserId.trim()}
          class="mt-2 w-full rounded-lg bg-shell-accent py-1.5 text-xs font-medium text-white transition-colors hover:bg-shell-accentHov disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Add Member
        </button>
      </div>
    {/if}

    <!-- Member list -->
    <div class="flex-1 overflow-y-auto py-2" role="list" aria-label="Workspace members">
      {#each members as member (member.user_id)}
        {@const displayName = member.username || member.user_id}
        {@const isCurrentUser = member.user_id === currentUserId}
        <div
          class="group flex items-center gap-3 px-4 py-2 transition-colors hover:bg-shell-surface"
          role="listitem"
        >
          <!-- Avatar -->
          <div
            class="grid h-8 w-8 flex-shrink-0 place-content-center rounded-lg text-xs font-bold text-white {avatarBg(displayName)}"
            aria-hidden="true"
          >
            {displayName.charAt(0).toUpperCase()}
          </div>

          <!-- Name + role -->
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium text-shell-ink">
              {displayName}
              {#if isCurrentUser}
                <span class="ml-1 text-xs font-normal text-shell-subtle">(you)</span>
              {/if}
            </p>
            <span class="inline-flex items-center rounded border px-1.5 py-px text-[10px] font-semibold {roleBadge[member.role] || roleBadge.member}">
              {member.role || "member"}
            </span>
          </div>

          <!-- Remove (admin only, not self, not owner) -->
          {#if isAdmin && !isCurrentUser && member.role !== "owner"}
            <button
              on:click={() => handleRemove(member.user_id)}
              class="rounded-md p-1 text-shell-subtle opacity-0 transition-all group-hover:opacity-100 hover:bg-shell-dangerBg hover:text-shell-danger"
              title="Remove {displayName}"
              aria-label="Remove {displayName} from workspace"
            >
              <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          {/if}
        </div>
      {:else}
        <div class="p-6 text-center">
          <p class="text-sm text-shell-subtle">No members found.</p>
        </div>
      {/each}
    </div>
  </aside>
{/if}
