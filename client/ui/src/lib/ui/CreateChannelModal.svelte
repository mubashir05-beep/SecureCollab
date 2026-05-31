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
    class="fixed inset-0 z-[110] grid place-content-center bg-charcoal/40 p-4 backdrop-blur-md animate-fade-in"
    on:click={(e) => e.currentTarget === e.target && close()}
    on:keydown={(e) => e.key === "Escape" && close()}
    role="button"
    tabindex="0"
  >
    <div
      class="w-[min(500px,90vw)] rounded-[40px] border border-borderSoft bg-white p-10 shadow-2xl animate-slide-up"
      role="dialog"
      tabindex="-1"
      aria-modal="true"
    >
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-clay shadow-lg shadow-clay/10 text-white" aria-hidden="true">
          <iconify-icon icon="lucide:hash" class="text-3xl"></iconify-icon>
        </div>
        <h3 class="text-2xl font-bold text-charcoal mb-2">Create a Channel</h3>
        <p class="text-sm text-muted font-medium">Focused spaces for specific topics or teams.</p>
      </div>

      {#if error}
        <div class="mb-6 p-4 rounded-2xl bg-red-50 border border-red-100 flex items-start gap-3 animate-slide-up">
          <iconify-icon icon="lucide:alert-circle" class="text-red-500 text-xl flex-shrink-0"></iconify-icon>
          <p class="text-sm font-bold text-red-600">{error}</p>
        </div>
      {/if}

      <form on:submit|preventDefault={submit} class="space-y-6">
        <div class="space-y-2">
          <label for="ch-name" class="text-[11px] font-bold text-muted uppercase tracking-widest ml-1">Channel Name</label>
          <div class="relative">
            <span class="absolute left-4 top-1/2 -translate-y-1/2 text-muted/40 font-bold text-lg">#</span>
            <input
              id="ch-name"
              type="text"
              bind:value={name}
              required
              placeholder="e.g. general-strategy"
              class="w-full pl-10 pr-5 py-4 rounded-2xl border border-borderSoft bg-sidebar/20 text-charcoal font-medium focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all outline-none"
            />
          </div>
          <p class="text-[10px] font-bold text-muted/30 uppercase tracking-tight ml-1">No spaces or uppercase letters.</p>
        </div>
        
        <div class="space-y-2">
          <label for="ch-topic" class="text-[11px] font-bold text-muted uppercase tracking-widest ml-1">Topic <span class="text-muted/40 font-normal">(optional)</span></label>
          <input
            id="ch-topic"
            type="text"
            bind:value={topic}
            placeholder="What's this channel about?"
            class="w-full px-5 py-4 rounded-2xl border border-borderSoft bg-sidebar/20 text-charcoal font-medium focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all outline-none"
          />
        </div>

        <!-- Private toggle -->
        <label class="group flex cursor-pointer items-center justify-between p-4 rounded-2xl border border-borderSoft bg-sidebar/10 hover:bg-sidebar/30 transition-all">
          <div class="flex items-center gap-4">
            <div class="w-10 h-10 rounded-xl flex items-center justify-center {isPrivate ? 'bg-sage text-white' : 'bg-white text-muted'} transition-all shadow-sm">
              <iconify-icon icon={isPrivate ? "lucide:lock" : "lucide:unlock"} class="text-xl"></iconify-icon>
            </div>
            <div>
              <span class="text-[14px] font-bold text-charcoal">Private Channel</span>
              <p class="text-[11px] font-medium text-muted">Only invited members can access.</p>
            </div>
          </div>
          <input
            type="checkbox"
            bind:checked={isPrivate}
            class="sr-only"
          />
          <div class="w-12 h-6 rounded-full bg-borderSoft p-1 transition-all {isPrivate ? 'bg-sage' : ''}">
            <div class="w-4 h-4 bg-white rounded-full transition-all {isPrivate ? 'translate-x-6' : 'translate-x-0'}"></div>
          </div>
        </label>

        <div class="flex flex-col sm:flex-row gap-3 pt-4">
          <Button variant="ghost" fullWidth={true} on:click={close}>Cancel</Button>
          <Button type="submit" variant="clay" fullWidth={true} disabled={!name.trim() || loading} {loading}>
            {loading ? "Creating…" : "Initialize Channel"}
          </Button>
        </div>
      </form>
    </div>
  </div>
{/if}
