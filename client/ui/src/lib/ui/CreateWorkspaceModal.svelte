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
    class="fixed inset-0 z-50 grid place-content-center bg-slate-900/60 p-4 backdrop-blur-sm"
    on:click={(e) => e.currentTarget === e.target && close()}
    on:keydown={(e) => e.key === "Escape" && close()}
    role="button" tabindex="0" aria-label="Close"
  >
    <div class="w-[min(420px,90vw)] rounded-2xl bg-white p-6 shadow-2xl" role="dialog" tabindex="-1" aria-modal="true">
      <h3 class="mb-1 text-lg font-bold text-slate-900">Create a Workspace</h3>
      <p class="mb-4 text-sm text-slate-500">Workspaces are where your team collaborates.</p>

      {#if error}
        <div class="mb-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{error}</div>
      {/if}

      <form on:submit|preventDefault={submit} class="space-y-3">
        <div>
          <label for="ws-name" class="mb-1 block text-xs font-medium text-slate-600">Workspace name</label>
          <input id="ws-name" type="text" bind:value={name} required
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30"
            placeholder="e.g. Engineering Team" />
        </div>
        <div>
          <label for="ws-desc" class="mb-1 block text-xs font-medium text-slate-600">Description (optional)</label>
          <textarea id="ws-desc" bind:value={description} rows="2"
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30"
            placeholder="What's this workspace for?" />
        </div>
        <div class="flex justify-end gap-2">
          <Button variant="ghost" on:click={close}>Cancel</Button>
          <Button type="submit" disabled={!name.trim() || loading}>
            {loading ? "Creating..." : "Create Workspace"}
          </Button>
        </div>
      </form>
    </div>
  </div>
{/if}
