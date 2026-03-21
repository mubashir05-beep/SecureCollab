<script>
  import { createEventDispatcher } from "svelte";

  export let visible = false;
  export let inviteCode = "";
  export let workspaceName = "";

  let joinCode = "";
  let error = "";
  let copied = false;

  const dispatch = createEventDispatcher();

  export function setError(msg) { error = msg; }
  export function reset() { joinCode = ""; error = ""; copied = false; }

  function copyCode() {
    navigator.clipboard.writeText(inviteCode).then(() => {
      copied = true;
      setTimeout(() => (copied = false), 2000);
    });
  }

  function handleJoin() {
    if (!joinCode.trim()) return;
    error = "";
    dispatch("join", joinCode.trim());
  }
</script>

{#if visible}
  <div class="fixed inset-0 z-50 grid place-content-center bg-black/50" on:click|self={() => dispatch("close")}>
    <div class="w-96 rounded-2xl bg-white p-6 shadow-xl">
      <h2 class="mb-4 text-lg font-bold text-slate-900">Invite & Join</h2>

      {#if error}
        <div class="mb-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">{error}</div>
      {/if}

      <!-- Share invite code -->
      {#if inviteCode}
        <div class="mb-5">
          <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-400">
            Share this invite code for {workspaceName}
          </label>
          <div class="flex gap-2">
            <input
              type="text"
              value={inviteCode}
              readonly
              class="flex-1 rounded-lg border border-slate-300 bg-slate-50 px-3 py-2 text-sm font-mono text-slate-700"
            />
            <button
              on:click={copyCode}
              class="rounded-lg bg-shell-accent px-3 py-2 text-sm font-medium text-white hover:opacity-90 transition"
            >
              {copied ? "Copied!" : "Copy"}
            </button>
          </div>
        </div>
      {/if}

      <hr class="my-4 border-slate-200" />

      <!-- Join with code -->
      <div>
        <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-400">
          Join a workspace with invite code
        </label>
        <div class="flex gap-2">
          <input
            type="text"
            bind:value={joinCode}
            placeholder="Paste invite code..."
            class="flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-700 placeholder-slate-400 focus:border-shell-accent focus:outline-none focus:ring-1 focus:ring-shell-accent/30"
            on:keydown={(e) => e.key === "Enter" && handleJoin()}
          />
          <button
            on:click={handleJoin}
            disabled={!joinCode.trim()}
            class="rounded-lg bg-slate-800 px-4 py-2 text-sm font-medium text-white hover:bg-slate-700 transition disabled:opacity-40"
          >
            Join
          </button>
        </div>
      </div>

      <button
        on:click={() => dispatch("close")}
        class="mt-4 w-full rounded-lg border border-slate-200 py-2 text-sm text-slate-500 hover:bg-slate-50 transition"
      >
        Close
      </button>
    </div>
  </div>
{/if}
