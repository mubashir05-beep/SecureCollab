<script>
  import Button from "./lib/ui/Button.svelte";
  import Panel from "./lib/ui/Panel.svelte";

  const workspaces = ["Atlas Security", "Audit Team", "Infra Ops"];
  const channels = ["# incident-room", "# architecture", "# release-readiness", "# key-management"];

  let selectedWorkspace = workspaces[0];
  let selectedChannel = channels[0];
  let showAuth = false;
</script>

<div class="mx-auto grid min-h-screen w-[min(1140px,95vw)] grid-cols-1 gap-4 py-6 lg:grid-cols-[300px_1fr]">
  <Panel>
    <div class="mb-6 flex items-center gap-3">
      <div class="grid h-12 w-12 place-content-center rounded-xl bg-gradient-to-br from-shell-accent to-shell-success text-white font-bold">SC</div>
      <div>
        <h1 class="text-lg font-bold">SecureCollab</h1>
        <p class="text-sm text-shell-muted">Workspace Console</p>
      </div>
    </div>

    <div class="mb-6">
      <p class="mb-2 font-mono text-xs uppercase tracking-[0.15em] text-shell-muted">Workspaces</p>
      <div class="space-y-2">
        {#each workspaces as name}
          <button on:click={() => (selectedWorkspace = name)} class={`w-full rounded-xl border px-3 py-2 text-left text-sm transition ${selectedWorkspace === name ? "border-teal-300 bg-teal-50" : "border-transparent bg-white hover:bg-shell-bg"}`}>
            {name}
          </button>
        {/each}
      </div>
    </div>

    <div>
      <p class="mb-2 font-mono text-xs uppercase tracking-[0.15em] text-shell-muted">Channels</p>
      <div class="space-y-2">
        {#each channels as name}
          <button on:click={() => (selectedChannel = name)} class={`w-full rounded-xl border px-3 py-2 text-left text-sm transition ${selectedChannel === name ? "border-teal-300 bg-teal-50" : "border-transparent bg-white hover:bg-shell-bg"}`}>
            {name}
          </button>
        {/each}
      </div>
    </div>
  </Panel>

  <Panel padded={false}>
    <header class="flex items-center justify-between border-b border-shell-line px-4 py-4">
      <div>
        <h2 class="text-base font-semibold">{selectedChannel}</h2>
        <p class="text-sm text-shell-muted">Encrypted | 14 members online</p>
      </div>
      <Button variant="ghost" on:click={() => (showAuth = true)}>Sign In</Button>
    </header>

    <div class="space-y-3 p-4">
      <article class="rounded-xl border border-cyan-200 bg-cyan-50 p-3 text-sm text-cyan-900">Secure session established. Message content remains end-to-end encrypted.</article>
      <article class="rounded-xl border border-shell-line bg-white p-3">
        <p class="mb-1 text-xs font-semibold uppercase text-teal-700">Amna</p>
        <p class="text-sm">Key rotation completed for workspace device group.</p>
      </article>
      <article class="rounded-xl border border-shell-line bg-white p-3">
        <p class="mb-1 text-xs font-semibold uppercase text-teal-700">You</p>
        <p class="text-sm">Gateway limiter dashboards look stable after Redis rollout.</p>
      </article>
    </div>

    <div class="grid grid-cols-[1fr_auto] gap-2 border-t border-shell-line p-4">
      <textarea rows="2" class="w-full rounded-xl border border-shell-line px-3 py-2 text-sm" placeholder="Write encrypted message..."></textarea>
      <Button>Send</Button>
    </div>
  </Panel>
</div>

{#if showAuth}
  <div class="fixed inset-0 z-50 grid place-content-center bg-slate-900/40 p-4" on:click={() => (showAuth = false)}>
    <Panel>
      <div class="w-[min(420px,90vw)]" on:click|stopPropagation>
        <h3 class="text-lg font-bold">Sign In</h3>
        <p class="mb-4 text-sm text-shell-muted">Continue to your encrypted workspace.</p>
        <div class="space-y-3">
          <input type="email" class="w-full rounded-xl border border-shell-line px-3 py-2 text-sm" placeholder="name@company.com" />
          <input type="password" class="w-full rounded-xl border border-shell-line px-3 py-2 text-sm" placeholder="Enter password" />
          <div class="flex justify-end gap-2">
            <Button variant="ghost" on:click={() => (showAuth = false)}>Cancel</Button>
            <Button on:click={() => (showAuth = false)}>Continue</Button>
          </div>
        </div>
      </div>
    </Panel>
  </div>
{/if}
