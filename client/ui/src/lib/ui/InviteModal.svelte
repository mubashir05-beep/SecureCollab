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
    // Use Tauri-compatible clipboard approach: navigator.clipboard works in Tauri WebView
    navigator.clipboard.writeText(inviteCode).then(() => {
      copied = true;
      setTimeout(() => (copied = false), 2000);
    }).catch(() => {
      // Fallback for restricted contexts
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
      aria-labelledby="invite-modal-title"
    >
      <!-- Header -->
      <div class="mb-5 flex items-start justify-between">
        <div>
          <h2 id="invite-modal-title" class="text-lg font-bold text-shell-ink">Invite & Join</h2>
          <p class="mt-1 text-sm text-shell-muted">Share a code or join a workspace.</p>
        </div>
        <button
          on:click={close}
          class="rounded-md p-1.5 text-shell-subtle transition-colors hover:bg-shell-surface hover:text-shell-ink"
          aria-label="Close invite dialog"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      {#if error}
        <div class="mb-4 rounded-lg bg-shell-dangerBg px-3 py-2.5 text-sm text-shell-danger" role="alert">{error}</div>
      {/if}

      <!-- Share invite code section -->
      {#if inviteCode}
        <div class="mb-5">
          <p class="mb-2 text-xs font-semibold uppercase tracking-wider text-shell-subtle">
            Invite to <span class="text-shell-muted">{workspaceName}</span>
          </p>
          <div class="flex gap-2">
            <input
              type="text"
              value={inviteCode}
              readonly
              aria-label="Invite code"
              class="flex-1 rounded-lg border border-shell-border bg-shell-bg px-3 py-2.5 font-mono text-sm text-shell-ink outline-none select-all"
            />
            <button
              on:click={copyCode}
              class="flex-shrink-0 rounded-lg bg-shell-accent px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-shell-accentHov"
              aria-label="{copied ? 'Copied' : 'Copy invite code'}"
            >
              {#if copied}
                <span class="flex items-center gap-1">
                  <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  Copied
                </span>
              {:else}
                Copy
              {/if}
            </button>
          </div>
        </div>

        <div class="mb-5 flex items-center gap-3">
          <hr class="flex-1 border-shell-borderSub" />
          <span class="text-xs text-shell-subtle">or</span>
          <hr class="flex-1 border-shell-borderSub" />
        </div>
      {/if}

      <!-- Join with code -->
      <div>
        <p class="mb-2 text-xs font-semibold uppercase tracking-wider text-shell-subtle">
          Join a workspace
        </p>
        <div class="flex gap-2">
          <input
            type="text"
            bind:value={joinCode}
            placeholder="Paste invite code…"
            aria-label="Enter invite code to join"
            class="flex-1 rounded-lg border border-shell-border bg-shell-bg px-3 py-2.5 text-sm text-shell-ink placeholder-shell-subtle outline-none transition-colors focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30"
            on:keydown={(e) => e.key === "Enter" && handleJoin()}
          />
          <button
            on:click={handleJoin}
            disabled={!joinCode.trim()}
            class="flex-shrink-0 rounded-lg bg-shell-surface px-4 py-2 text-sm font-medium text-shell-ink transition-colors hover:bg-shell-elevated disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Join
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
