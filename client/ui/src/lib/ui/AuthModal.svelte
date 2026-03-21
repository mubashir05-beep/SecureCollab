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
  <div
    class="fixed inset-0 z-50 grid place-content-center bg-slate-900/60 p-4 backdrop-blur-sm"
    on:click={(e) => e.currentTarget === e.target && close()}
    on:keydown={(e) => e.key === "Escape" && close()}
    role="button"
    tabindex="0"
    aria-label="Close dialog"
  >
    <div
      class="w-[min(400px,90vw)] rounded-2xl bg-white p-6 shadow-2xl"
      role="dialog"
      tabindex="-1"
      aria-modal="true"
    >
      <!-- Logo -->
      <div class="mb-4 flex items-center gap-3">
        <div class="grid h-10 w-10 place-content-center rounded-xl bg-gradient-to-br from-shell-accent to-shell-success text-white font-bold text-sm">SC</div>
        <div>
          <h3 class="text-lg font-bold text-slate-900">{mode === "register" ? "Create Account" : "Welcome Back"}</h3>
          <p class="text-sm text-slate-500">Continue to SecureCollab</p>
        </div>
      </div>

      {#if error}
        <div class="mb-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{error}</div>
      {/if}

      <form on:submit|preventDefault={handleSubmit} class="space-y-3">
        <div>
          <label for="auth-username" class="mb-1 block text-xs font-medium text-slate-600">Username</label>
          <input
            id="auth-username"
            type="text"
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30 outline-none"
            bind:value={username}
            required
          />
        </div>

        {#if mode === "register"}
          <div>
            <label for="auth-email" class="mb-1 block text-xs font-medium text-slate-600">Email</label>
            <input
              id="auth-email"
              type="email"
              class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30 outline-none"
              bind:value={email}
              required
            />
          </div>
        {/if}

        <div>
          <label for="auth-password" class="mb-1 block text-xs font-medium text-slate-600">Password</label>
          <input
            id="auth-password"
            type="password"
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30 outline-none"
            bind:value={password}
            required
          />
        </div>

        <Button type="submit">
          {#if loading}Signing in...{:else}{mode === "register" ? "Create Account" : "Sign In"}{/if}
        </Button>

        <p class="text-center text-sm text-slate-500">
          {#if mode === "login"}
            Don't have an account?
            <button type="button" class="text-shell-accent hover:underline" on:click={() => { mode = "register"; error = ""; }}>Sign up</button>
          {:else}
            Already have an account?
            <button type="button" class="text-shell-accent hover:underline" on:click={() => { mode = "login"; error = ""; }}>Sign in</button>
          {/if}
        </p>
      </form>
    </div>
  </div>
{/if}
