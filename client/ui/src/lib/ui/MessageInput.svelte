<script>
  import { createEventDispatcher } from "svelte";

  export let placeholder = "Message #channel";
  export let disabled = false;
  export let members = []; // { user_id, username, role }

  let text = "";
  let mentionQuery = "";
  let showMentions = false;
  let mentionIndex = 0;
  let textarea;

  const dispatch = createEventDispatcher();

  const specialMentions = [
    { user_id: "@channel", username: "channel", role: "notify" },
    { user_id: "@here",    username: "here",    role: "notify" },
  ];

  $: mentionCandidates = (() => {
    if (!mentionQuery) return [];
    const q = mentionQuery.toLowerCase();
    return [...specialMentions, ...members]
      .filter(m => m.username?.toLowerCase().startsWith(q) || m.user_id?.toLowerCase().startsWith(q))
      .slice(0, 8);
  })();

  // Close mention popup when no candidates remain
  $: if (showMentions && mentionCandidates.length === 0) showMentions = false;

  function handleInput(e) {
    const val = e.target.value;
    const pos = e.target.selectionStart;
    const before = val.slice(0, pos);
    const atMatch = before.match(/(^|\s)@(\w*)$/);
    if (atMatch) {
      mentionQuery = atMatch[2];
      showMentions = true;
      mentionIndex = 0;
    } else {
      showMentions = false;
      mentionQuery = "";
    }
  }

  function insertMention(candidate) {
    const pos = textarea.selectionStart;
    const before = text.slice(0, pos);
    const after = text.slice(pos);
    const atPos = before.lastIndexOf("@");
    const name = candidate.user_id.startsWith("@") ? candidate.user_id : `@${candidate.username}`;
    text = before.slice(0, atPos) + name + " " + after;
    showMentions = false;
    mentionQuery = "";
    setTimeout(() => textarea?.focus(), 0);
  }

  function handleKeydown(e) {
    if (showMentions && mentionCandidates.length > 0) {
      if (e.key === "ArrowDown")  { e.preventDefault(); mentionIndex = (mentionIndex + 1) % mentionCandidates.length; return; }
      if (e.key === "ArrowUp")    { e.preventDefault(); mentionIndex = (mentionIndex - 1 + mentionCandidates.length) % mentionCandidates.length; return; }
      if (e.key === "Tab" || e.key === "Enter") { e.preventDefault(); insertMention(mentionCandidates[mentionIndex]); return; }
      if (e.key === "Escape")     { showMentions = false; return; }
    }
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  }

  function send() {
    if (!text.trim() || disabled) return;
    dispatch("send", text.trim());
    text = "";
    showMentions = false;
  }
</script>

<div class="relative border-t border-shell-borderSub bg-shell-bg px-4 py-3">

  <!-- Mention autocomplete -->
  {#if showMentions && mentionCandidates.length > 0}
    <div
      class="absolute bottom-full left-4 right-4 mb-1 overflow-hidden rounded-xl border border-shell-border bg-shell-elevated shadow-panel animate-slide-up z-50"
      role="listbox"
      aria-label="Mention suggestions"
    >
      {#each mentionCandidates as candidate, i}
        <button
          role="option"
          aria-selected={i === mentionIndex}
          class="flex w-full items-center gap-2.5 px-3 py-2 text-left text-sm transition-colors
            {i === mentionIndex ? 'bg-shell-accent/20 text-white' : 'text-shell-muted hover:bg-shell-surface hover:text-shell-ink'}"
          on:mousedown|preventDefault={() => insertMention(candidate)}
        >
          <div class="grid h-6 w-6 flex-shrink-0 place-content-center rounded-md text-xs font-bold
            {candidate.role === 'notify' ? 'bg-shell-warn/20 text-shell-warn' : 'bg-shell-surface text-shell-muted'}">
            {candidate.role === "notify" ? "@" : candidate.username?.charAt(0)?.toUpperCase() || "?"}
          </div>
          <span class="font-medium text-shell-ink">
            {candidate.user_id.startsWith("@") ? candidate.user_id : `@${candidate.username}`}
          </span>
          {#if candidate.role && candidate.role !== "notify"}
            <span class="ml-auto text-xs text-shell-subtle">{candidate.role}</span>
          {/if}
        </button>
      {/each}
    </div>
  {/if}

  <!-- Input box -->
  <div
    class="flex items-end gap-2 rounded-xl border border-shell-border bg-shell-elevated px-3 py-2 transition-colors duration-150 focus-within:border-shell-accent"
  >
    <!-- Attach button -->
    <button
      class="flex-shrink-0 rounded-md p-1 text-shell-subtle transition-colors hover:bg-shell-surface hover:text-shell-ink"
      title="Attach file (coming soon)"
      aria-label="Attach file"
      on:click={() => dispatch("attach")}
    >
      <svg class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
      </svg>
    </button>

    <!-- Textarea -->
    <textarea
      bind:this={textarea}
      bind:value={text}
      on:keydown={handleKeydown}
      on:input={handleInput}
      {placeholder}
      rows="1"
      {disabled}
      aria-label={placeholder}
      class="flex-1 resize-none bg-transparent text-sm text-shell-ink placeholder-shell-subtle outline-none leading-relaxed"
      style="max-height: 160px;"
    ></textarea>

    <!-- Emoji button -->
    <button
      class="flex-shrink-0 rounded-md p-1 text-shell-subtle transition-colors hover:bg-shell-surface hover:text-shell-ink"
      title="Insert emoji"
      aria-label="Insert emoji"
    >
      <svg class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    </button>

    <!-- Send button -->
    <button
      on:click={send}
      disabled={!text.trim() || disabled}
      title="Send message (Enter)"
      aria-label="Send message"
      class="flex-shrink-0 rounded-lg bg-shell-accent p-1.5 text-white transition-colors hover:bg-shell-accentHov disabled:opacity-30 disabled:cursor-not-allowed"
    >
      <svg class="h-4 w-4" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
        <path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z" />
      </svg>
    </button>
  </div>

  <!-- Hint text -->
  <p class="mt-1 pl-1 text-xs text-shell-subtle">
    Enter to send &middot; Shift+Enter for new line &middot; @ to mention
  </p>
</div>
