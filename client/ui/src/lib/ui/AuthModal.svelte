<script>
  import { createEventDispatcher } from "svelte";
  import Button from "./Button.svelte";

  export let visible = false;

  let mode = "login";
  let username = "";
  let email = "";
  let password = "";
  let error = "";
  let loading = false;

  const dispatch = createEventDispatcher();

  function handleSubmit() {
    if (loading) return;
    error = "";
    loading = true;
    dispatch("auth", { mode, username, email, password });
  }

  export function setError(msg) {
    error = msg;
    loading = false;
  }

  export function setLoading(val) {
    loading = val;
  }

  function close() {
    dispatch("close");
    reset();
  }

  function reset() {
    username = "";
    email = "";
    password = "";
    error = "";
    mode = "login";
    loading = false;
  }
</script>

{#if visible}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-50 grid place-content-center bg-black/70 p-4 backdrop-blur-sm animate-fade-in"
    on:click={(e) => e.currentTarget === e.target && close()}
    on:keydown={(e) => e.key === "Escape" && close()}
    role="button"
    tabindex="0"
    aria-label="Close dialog"
  >
    <!-- Dialog -->
    <div
      class="w-[min(400px,90vw)] rounded-2xl border border-shell-border bg-shell-elevated p-6 shadow-modal animate-slide-up"
      role="dialog"
      tabindex="-1"
      aria-modal="true"
      aria-labelledby="auth-modal-title"
    >
      <!-- Logo + title -->
      <div class="mb-5 flex items-center gap-3">
        <div class="grid h-10 w-10 flex-shrink-0 place-content-center rounded-xl bg-shell-accent text-sm font-bold text-white" aria-hidden="true">
          SC
        </div>
        <div>
          <h3 id="auth-modal-title" class="text-lg font-bold text-shell-ink">
            {mode === "register" ? "Create Account" : "Welcome Back"}
          </h3>
          <p class="text-sm text-shell-muted">Continue to SecureCollab</p>
        </div>
      </div>

      <!-- Error banner -->
      {#if error}
        <div class="mb-4 flex items-start gap-2 rounded-lg bg-shell-dangerBg px-3 py-2.5 text-sm text-shell-danger" role="alert">
          <svg class="mt-0.5 h-4 w-4 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
          </svg>
          {error}
        </div>
      {/if}

      <form on:submit|preventDefault={handleSubmit} class="space-y-4">
        <!-- Username -->
        <div>
          <label for="auth-username" class="mb-1.5 block text-xs font-medium text-shell-muted">Username</label>
          <input
            id="auth-username"
            type="text"
            bind:value={username}
            required
            autocomplete="username"
            class="w-full rounded-lg border border-shell-border bg-shell-bg px-3 py-2.5 text-sm text-shell-ink placeholder-shell-subtle outline-none transition-colors focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30"
            placeholder="your-username"
          />
        </div>

        <!-- Email (register only) -->
        {#if mode === "register"}
          <div>
            <label for="auth-email" class="mb-1.5 block text-xs font-medium text-shell-muted">Email</label>
            <input
              id="auth-email"
              type="email"
              bind:value={email}
              required
              autocomplete="email"
              class="w-full rounded-lg border border-shell-border bg-shell-bg px-3 py-2.5 text-sm text-shell-ink placeholder-shell-subtle outline-none transition-colors focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30"
              placeholder="you@example.com"
            />
          </div>
        {/if}

        <!-- Password -->
        <div>
          <label for="auth-password" class="mb-1.5 block text-xs font-medium text-shell-muted">Password</label>
          <input
            id="auth-password"
            type="password"
            bind:value={password}
            required
            autocomplete={mode === "register" ? "new-password" : "current-password"}
            class="w-full rounded-lg border border-shell-border bg-shell-bg px-3 py-2.5 text-sm text-shell-ink placeholder-shell-subtle outline-none transition-colors focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30"
            placeholder="••••••••"
          />
        </div>

        <Button type="submit" fullWidth={true} {loading}>
          {loading ? "Signing in…" : mode === "register" ? "Create Account" : "Sign In"}
        </Button>

        <p class="text-center text-sm text-shell-subtle">
          {#if mode === "login"}
            No account?
            <button
              type="button"
              class="font-medium text-shell-accentText hover:underline"
              on:click={() => { mode = "register"; error = ""; }}
            >Sign up</button>
          {:else}
            Have an account?
            <button
              type="button"
              class="font-medium text-shell-accentText hover:underline"
              on:click={() => { mode = "login"; error = ""; }}
            >Sign in</button>
          {/if}
        </p>
      </form>
    </div>
  </div>
{/if}
