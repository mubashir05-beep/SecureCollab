<script>
  import { createEventDispatcher } from "svelte";
  import Button from "./Button.svelte";

  export let visible = false;

  let name = "";
  let description = "";
  let error = "";
  let loading = false;

  const dispatch = createEventDispatcher();

  function submit() {
    if (!name.trim() || loading) return;
    error = "";
    loading = true;
    dispatch("create", { name: name.trim(), description: description.trim() });
  }

  export function setError(msg) { error = msg; loading = false; }
  export function reset() { name = ""; description = ""; error = ""; loading = false; }

  function close() { dispatch("close"); reset(); }
</script>

{#if visible}
  <div
    class="fixed inset-0 z-50 grid place-content-center bg-black/70 p-4 backdrop-blur-sm animate-fade-in"
    on:click={(e) => e.currentTarget === e.target && close()}
    on:keydown={(e) => e.key === "Escape" && close()}
    role="button"
    tabindex="0"
    aria-label="Close dialog"
  >
    <div
      class="w-[min(440px,90vw)] rounded-2xl border border-shell-border bg-shell-elevated p-6 shadow-modal animate-slide-up"
      role="dialog"
      tabindex="-1"
      aria-modal="true"
      aria-labelledby="create-ws-title"
    >
      <!-- Header -->
      <div class="mb-5">
        <div class="mb-3 grid h-10 w-10 place-content-center rounded-xl bg-shell-accent/20 text-shell-accent" aria-hidden="true">
          <svg class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-2 10v-5a1 1 0 00-1-1h-2a1 1 0 00-1 1v5m4 0H9" />
          </svg>
        </div>
        <h3 id="create-ws-title" class="text-lg font-bold text-shell-ink">Create a Workspace</h3>
        <p class="mt-1 text-sm text-shell-muted">Workspaces are where your team collaborates.</p>
      </div>

      {#if error}
        <div class="mb-4 rounded-lg bg-shell-dangerBg px-3 py-2.5 text-sm text-shell-danger" role="alert">{error}</div>
      {/if}

      <form on:submit|preventDefault={submit} class="space-y-4">
        <div>
          <label for="ws-name" class="mb-1.5 block text-xs font-medium text-shell-muted">Workspace name</label>
          <input
            id="ws-name"
            type="text"
            bind:value={name}
            required
            placeholder="e.g. Engineering Team"
            class="w-full rounded-lg border border-shell-border bg-shell-bg px-3 py-2.5 text-sm text-shell-ink placeholder-shell-subtle outline-none transition-colors focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30"
          />
        </div>
        <div>
          <label for="ws-desc" class="mb-1.5 block text-xs font-medium text-shell-muted">Description <span class="text-shell-subtle font-normal">(optional)</span></label>
          <textarea
            id="ws-desc"
            bind:value={description}
            rows="2"
            placeholder="What's this workspace for?"
            class="w-full rounded-lg border border-shell-border bg-shell-bg px-3 py-2.5 text-sm text-shell-ink placeholder-shell-subtle outline-none transition-colors focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30 resize-none"
          ></textarea>
        </div>
        <div class="flex justify-end gap-2 pt-1">
          <Button variant="ghost" on:click={close}>Cancel</Button>
          <Button type="submit" disabled={!name.trim() || loading} {loading}>
            {loading ? "Creating…" : "Create Workspace"}
          </Button>
        </div>
      </form>
    </div>
  </div>
{/if}
