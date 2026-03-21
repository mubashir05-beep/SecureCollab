<script>
  import { createEventDispatcher } from "svelte";
  import Button from "./Button.svelte";

  export let visible = false;

  let name = "";
  let topic = "";
  let isPrivate = false;
  let error = "";
  let loading = false;

  const dispatch = createEventDispatcher();

  function submit() {
    if (!name.trim() || loading) return;
    error = "";
    loading = true;
    dispatch("create", { name: name.trim(), topic: topic.trim(), isPrivate });
  }

  export function setError(msg) { error = msg; loading = false; }
  export function reset() { name = ""; topic = ""; isPrivate = false; error = ""; loading = false; }

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
      <h3 class="mb-1 text-lg font-bold text-slate-900">Create a Channel</h3>
      <p class="mb-4 text-sm text-slate-500">Channels are where conversations happen.</p>

      {#if error}
        <div class="mb-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{error}</div>
      {/if}

      <form on:submit|preventDefault={submit} class="space-y-3">
        <div>
          <label for="ch-name" class="mb-1 block text-xs font-medium text-slate-600">Channel name</label>
          <div class="flex items-center rounded-lg border border-slate-300 focus-within:border-shell-accent focus-within:ring-1 focus-within:ring-shell-accent/30">
            <span class="pl-3 text-slate-400">#</span>
            <input id="ch-name" type="text" bind:value={name} required
              class="flex-1 bg-transparent px-2 py-2 text-sm outline-none"
              placeholder="e.g. design-reviews" />
          </div>
        </div>
        <div>
          <label for="ch-topic" class="mb-1 block text-xs font-medium text-slate-600">Topic (optional)</label>
          <input id="ch-topic" type="text" bind:value={topic}
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30"
            placeholder="What's this channel about?" />
        </div>
        <label class="flex items-center gap-2 text-sm text-slate-700">
          <input type="checkbox" bind:checked={isPrivate} class="rounded border-slate-300" />
          Make private — only invited members can see this channel
        </label>
        <div class="flex justify-end gap-2">
          <Button variant="ghost" on:click={close}>Cancel</Button>
          <Button type="submit" disabled={!name.trim() || loading}>
            {loading ? "Creating..." : "Create Channel"}
          </Button>
        </div>
      </form>
    </div>
  </div>
{/if}
