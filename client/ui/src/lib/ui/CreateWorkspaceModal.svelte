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
    class="fixed inset-0 z-[110] grid place-content-center bg-charcoal/40 p-4 backdrop-blur-md animate-fade-in"
    on:click={(e) => e.currentTarget === e.target && close()}
    on:keydown={(e) => e.key === "Escape" && close()}
    role="button"
    tabindex="0"
  >
    <div
      class="w-[min(480px,90vw)] rounded-[40px] border border-borderSoft bg-white p-10 shadow-2xl animate-slide-up"
      role="dialog"
      tabindex="-1"
      aria-modal="true"
    >
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-sage shadow-lg shadow-sage/10 text-white" aria-hidden="true">
          <iconify-icon icon="lucide:layout-grid" class="text-3xl"></iconify-icon>
        </div>
        <h3 class="text-2xl font-bold text-charcoal mb-2">Create a Workspace</h3>
        <p class="text-sm text-muted font-medium">Define your new hub for collaboration.</p>
      </div>

      {#if error}
        <div class="mb-6 p-4 rounded-2xl bg-red-50 border border-red-100 flex items-start gap-3 animate-slide-up">
          <iconify-icon icon="lucide:alert-circle" class="text-red-500 text-xl flex-shrink-0"></iconify-icon>
          <p class="text-sm font-bold text-red-600">{error}</p>
        </div>
      {/if}

      <form on:submit|preventDefault={submit} class="space-y-6">
        <div class="space-y-2">
          <label for="ws-name" class="text-[11px] font-bold text-muted uppercase tracking-widest ml-1">Workspace Name</label>
          <input
            id="ws-name"
            type="text"
            bind:value={name}
            required
            placeholder="e.g. Creative Labs"
            class="w-full px-5 py-4 rounded-2xl border border-borderSoft bg-sidebar/20 text-charcoal font-medium focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all outline-none"
          />
        </div>
        
        <div class="space-y-2">
          <label for="ws-desc" class="text-[11px] font-bold text-muted uppercase tracking-widest ml-1">Description <span class="text-muted/40 font-normal">(optional)</span></label>
          <textarea
            id="ws-desc"
            bind:value={description}
            rows="3"
            placeholder="What's this workspace for?"
            class="w-full px-5 py-4 rounded-2xl border border-borderSoft bg-sidebar/20 text-charcoal font-medium focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all outline-none resize-none"
          ></textarea>
        </div>

        <div class="flex flex-col sm:flex-row gap-3 pt-4">
          <Button variant="ghost" fullWidth={true} on:click={close}>Cancel</Button>
          <Button type="submit" variant="sage" fullWidth={true} disabled={!name.trim() || loading} {loading}>
            {loading ? "Creating…" : "Establish Workspace"}
          </Button>
        </div>
      </form>
    </div>
  </div>
{/if}
