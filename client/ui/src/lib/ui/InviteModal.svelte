<script>
  import { createEventDispatcher } from "svelte";
  import Button from "./Button.svelte";

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
    }).catch(() => {
      const el = document.createElement("textarea");
      el.value = inviteCode;
      document.body.appendChild(el);
      el.select();
      document.execCommand("copy");
      document.body.removeChild(el);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    });
  }

  function handleJoin() {
    if (!joinCode.trim()) return;
    error = "";
    dispatch("join", joinCode.trim());
  }

  function close() { dispatch("close"); }
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
      class="w-[min(480px,90vw)] rounded-[40px] border border-borderSoft bg-white p-10 shadow-2xl animate-slide-up relative"
      role="dialog"
      tabindex="-1"
      aria-modal="true"
    >
      <button 
        on:click={close}
        class="absolute top-6 right-6 w-10 h-10 flex items-center justify-center rounded-xl hover:bg-sidebar text-muted transition-all"
      >
        <iconify-icon icon="lucide:x" class="text-xl"></iconify-icon>
      </button>

      <!-- Header -->
      <div class="text-center mb-8">
        <div class="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-sage shadow-lg shadow-sage/10 text-white" aria-hidden="true">
          <iconify-icon icon="lucide:user-plus" class="text-3xl"></iconify-icon>
        </div>
        <h3 class="text-2xl font-bold text-charcoal mb-2">Invite & Join</h3>
        <p class="text-sm text-muted font-medium">Expand your team or enter a new workspace.</p>
      </div>

      {#if error}
        <div class="mb-6 p-4 rounded-2xl bg-red-50 border border-red-100 flex items-start gap-3 animate-slide-up">
          <iconify-icon icon="lucide:alert-circle" class="text-red-500 text-xl flex-shrink-0"></iconify-icon>
          <p class="text-sm font-bold text-red-600">{error}</p>
        </div>
      {/if}

      <div class="space-y-8">
        <!-- Share invite code section -->
        {#if inviteCode}
          <div class="space-y-4">
            <div class="flex items-center justify-between px-1">
              <label class="text-[11px] font-bold text-muted uppercase tracking-widest">Invite to {workspaceName}</label>
            </div>
            <div class="flex gap-3">
              <div class="flex-1 relative">
                <input
                  type="text"
                  value={inviteCode}
                  readonly
                  class="w-full px-5 py-4 rounded-2xl border border-borderSoft bg-sidebar/20 text-charcoal font-mono text-sm outline-none select-all"
                />
              </div>
              <Button variant={copied ? "primary" : "sage"} on:click={copyCode}>
                {#if copied}
                  <iconify-icon icon="lucide:check" class="text-lg"></iconify-icon>
                {:else}
                  <iconify-icon icon="lucide:copy" class="text-lg"></iconify-icon>
                {/if}
              </Button>
            </div>
          </div>

          <div class="flex items-center gap-4 py-2">
            <div class="flex-1 h-px bg-borderSoft"></div>
            <span class="text-[10px] font-bold text-muted/40 uppercase tracking-widest">or</span>
            <div class="flex-1 h-px bg-borderSoft"></div>
          </div>
        {/if}

        <!-- Join with code -->
        <div class="space-y-4">
          <div class="flex items-center justify-between px-1">
            <label class="text-[11px] font-bold text-muted uppercase tracking-widest">Join with Code</label>
          </div>
          <div class="flex gap-3">
            <input
              type="text"
              bind:value={joinCode}
              placeholder="Paste invite code..."
              class="flex-1 px-5 py-4 rounded-2xl border border-borderSoft bg-sidebar/20 text-charcoal font-medium focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all outline-none"
              on:keydown={(e) => e.key === "Enter" && handleJoin()}
            />
            <Button variant="clay" on:click={handleJoin} disabled={!joinCode.trim()}>
              Join
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}
