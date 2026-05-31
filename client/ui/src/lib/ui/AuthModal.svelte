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
  <div class="fixed inset-0 z-[100] flex flex-col md:flex-row bg-ivory animate-fade-in overflow-y-auto">
    <!-- Left side: Branding/Visual (hidden on mobile) -->
    <div class="hidden md:flex md:w-[40%] bg-sidebar border-r border-borderSoft flex-col justify-between p-12 relative overflow-hidden">
      <div class="absolute top-[-10%] left-[-10%] w-[60%] h-[60%] bg-sage/5 rounded-full blur-3xl"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[60%] h-[60%] bg-clay/5 rounded-full blur-3xl"></div>

      <div class="relative z-10">
        <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-sage text-white shadow-lg shadow-sage/20 font-bold text-xl mb-8">
          SC
        </div>
        <h1 class="text-4xl font-bold text-charcoal leading-tight mb-4">The future of secure collaboration.</h1>
        <p class="text-muted text-lg font-medium leading-relaxed">
          Zero-knowledge workspace merging chat, tasks, and productivity in one beautiful place.
        </p>
      </div>

      <div class="relative z-10 flex items-center gap-4 text-muted text-sm font-bold uppercase tracking-widest">
        <span>&copy; 2024 SecureCollab</span>
        <span class="w-1 h-1 rounded-full bg-borderSoft"></span>
        <span>Privacy First</span>
      </div>
    </div>

    <!-- Right side: Form -->
    <div class="flex-1 flex flex-col items-center justify-center p-6 sm:p-12 relative">
      <button 
        on:click={close}
        class="absolute top-8 left-8 sm:left-12 flex items-center gap-2 text-muted hover:text-charcoal transition-colors font-bold text-sm"
      >
        <iconify-icon icon="lucide:arrow-left" class="text-lg"></iconify-icon>
        Back to Home
      </button>

      <div class="w-full max-w-[400px] animate-slide-up">
        <div class="mb-10 text-center md:text-left">
          <h2 class="text-3xl font-bold text-charcoal mb-2">
            {mode === "register" ? "Join the Workspace" : "Welcome Back"}
          </h2>
          <p class="text-muted font-medium">
            {mode === "register" ? "Create your secure account to get started." : "Sign in to continue your collaboration."}
          </p>
        </div>

        {#if error}
          <div class="mb-6 p-4 rounded-2xl bg-red-50 border border-red-100 flex items-start gap-3 animate-slide-up">
            <iconify-icon icon="lucide:alert-circle" class="text-red-500 text-xl flex-shrink-0 mt-0.5"></iconify-icon>
            <p class="text-sm font-bold text-red-600 leading-snug">{error}</p>
          </div>
        {/if}

        <form on:submit|preventDefault={handleSubmit} class="space-y-6">
          <div class="space-y-1.5">
            <label for="auth-username" class="text-[11px] font-bold text-muted uppercase tracking-widest ml-1">Username</label>
            <div class="relative">
              <span class="absolute left-4 top-1/2 -translate-y-1/2 text-muted/40">
                <iconify-icon icon="lucide:user" class="text-xl"></iconify-icon>
              </span>
              <input
                id="auth-username"
                type="text"
                bind:value={username}
                required
                class="w-full pl-12 pr-4 py-4 rounded-2xl border border-borderSoft bg-sidebar/20 text-charcoal font-medium focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all outline-none placeholder:text-muted/30"
                placeholder="yourname"
              />
            </div>
          </div>

          {#if mode === "register"}
            <div class="space-y-1.5">
              <label for="auth-email" class="text-[11px] font-bold text-muted uppercase tracking-widest ml-1">Email Address</label>
              <div class="relative">
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-muted/40">
                  <iconify-icon icon="lucide:mail" class="text-xl"></iconify-icon>
                </span>
                <input
                  id="auth-email"
                  type="email"
                  bind:value={email}
                  required
                  class="w-full pl-12 pr-4 py-4 rounded-2xl border border-borderSoft bg-sidebar/20 text-charcoal font-medium focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all outline-none placeholder:text-muted/30"
                  placeholder="name@company.com"
                />
              </div>
            </div>
          {/if}

          <div class="space-y-1.5">
            <label for="auth-password" class="text-[11px] font-bold text-muted uppercase tracking-widest ml-1">Password</label>
            <div class="relative">
              <span class="absolute left-4 top-1/2 -translate-y-1/2 text-muted/40">
                <iconify-icon icon="lucide:lock" class="text-xl"></iconify-icon>
              </span>
              <input
                id="auth-password"
                type="password"
                bind:value={password}
                required
                class="w-full pl-12 pr-4 py-4 rounded-2xl border border-borderSoft bg-sidebar/20 text-charcoal font-medium focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all outline-none placeholder:text-muted/30"
                placeholder="••••••••"
              />
            </div>
          </div>

          <Button type="submit" fullWidth={true} size="lg" variant={mode === 'register' ? 'clay' : 'sage'} {loading}>
            {mode === 'register' ? 'Create Secure Account' : 'Sign In to Workspace'}
          </Button>

          <p class="text-center text-sm font-medium text-muted">
            {#if mode === "login"}
              Don't have an account?
              <button
                type="button"
                class="font-bold text-sage hover:underline underline-offset-4"
                on:click={() => { mode = "register"; error = ""; }}
              >Sign up for free</button>
            {:else}
              Already a member?
              <button
                type="button"
                class="font-bold text-clay hover:underline underline-offset-4"
                on:click={() => { mode = "login"; error = ""; }}
              >Sign in here</button>
            {/if}
          </p>
        </form>
      </div>
    </div>
  </div>
{/if}
