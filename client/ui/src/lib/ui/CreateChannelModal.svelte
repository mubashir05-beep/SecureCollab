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
      aria-labelledby="create-ch-title"
    >
      <!-- Header -->
      <div class="mb-5">
        <div class="mb-3 grid h-10 w-10 place-content-center rounded-xl bg-shell-accent/20 text-xl text-shell-subtle" aria-hidden="true">#</div>
        <h3 id="create-ch-title" class="text-lg font-bold text-shell-ink">Create a Channel</h3>
        <p class="mt-1 text-sm text-shell-muted">Channels are where conversations happen around a topic.</p>
      </div>

      {#if error}
        <div class="mb-4 rounded-lg bg-shell-dangerBg px-3 py-2.5 text-sm text-shell-danger" role="alert">{error}</div>
      {/if}

      <form on:submit|preventDefault={submit} class="space-y-4">
        <!-- Channel name with # prefix -->
        <div>
          <label for="ch-name" class="mb-1.5 block text-xs font-medium text-shell-muted">Channel name</label>
          <div class="flex items-center overflow-hidden rounded-lg border border-shell-border bg-shell-bg transition-colors focus-within:border-shell-accent focus-within:ring-1 focus-within:ring-shell-accent/30">
            <span class="flex-shrink-0 px-3 text-sm text-shell-subtle select-none" aria-hidden="true">#</span>
            <input
              id="ch-name"
              type="text"
              bind:value={name}
              required
              placeholder="e.g. design-reviews"
              class="flex-1 bg-transparent py-2.5 pr-3 text-sm text-shell-ink placeholder-shell-subtle outline-none"
            />
          </div>
          <p class="mt-1 text-xs text-shell-subtle">Lowercase letters, numbers, and hyphens only.</p>
        </div>

        <!-- Topic -->
        <div>
          <label for="ch-topic" class="mb-1.5 block text-xs font-medium text-shell-muted">
            Topic <span class="font-normal text-shell-subtle">(optional)</span>
          </label>
          <input
            id="ch-topic"
            type="text"
            bind:value={topic}
            placeholder="What's this channel about?"
            class="w-full rounded-lg border border-shell-border bg-shell-bg px-3 py-2.5 text-sm text-shell-ink placeholder-shell-subtle outline-none transition-colors focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30"
          />
        </div>

        <!-- Private toggle -->
        <label class="flex cursor-pointer items-start gap-3">
          <input
            type="checkbox"
            bind:checked={isPrivate}
            class="mt-0.5 h-4 w-4 rounded border-shell-border bg-shell-bg text-shell-accent focus:ring-shell-accent focus:ring-offset-shell-elevated"
          />
          <div>
            <span class="text-sm font-medium text-shell-ink">Make private</span>
            <p class="text-xs text-shell-muted">Only invited members can view this channel.</p>
          </div>
        </label>

        <div class="flex justify-end gap-2 pt-1">
          <Button variant="ghost" on:click={close}>Cancel</Button>
          <Button type="submit" disabled={!name.trim() || loading} {loading}>
            {loading ? "Creating…" : "Create Channel"}
          </Button>
        </div>
      </form>
    </div>
  </div>
{/if}
