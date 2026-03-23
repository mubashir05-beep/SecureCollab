<script>
  import { createEventDispatcher } from "svelte";
  import EmojiPicker from "./EmojiPicker.svelte";
  import MarkdownText from "./MarkdownText.svelte";
  import LinkPreview from "./LinkPreview.svelte";

  export let sender = "";
  export let content = "";
  export let timestamp = "";
  export let isPinned = false;
  export let isOwn = false;
  export let isEdited = false;
  export let reactions = [];
  export let messageId = "";

  let showEmojiPicker = false;

  const dispatch = createEventDispatcher();

  // Pull URLs for link preview
  const urlRegex = /https?:\/\/[^\s<"]+/g;
  $: urls = (typeof content === "string" ? content : "").match(urlRegex) || [];

  // Avatar color deterministic from sender string
  const avatarPalette = [
    "bg-indigo-500", "bg-violet-500", "bg-sky-500",
    "bg-teal-500",   "bg-rose-500",   "bg-amber-500",
    "bg-green-600",  "bg-pink-500",
  ];
  $: avatarBg = avatarPalette[(sender?.charCodeAt(0) || 0) % avatarPalette.length];

  function formatTime(ts) {
    if (!ts) return "";
    try {
      return new Date(ts).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    } catch {
      return ts;
    }
  }

  function formatDate(ts) {
    if (!ts) return "";
    try {
      return new Date(ts).toLocaleDateString([], { month: "short", day: "numeric" });
    } catch {
      return "";
    }
  }

  const emojiMap = {
    thumbsup: "👍", thumbsdown: "👎", heart: "❤️", joy: "😂",
    tada: "🎉",    eyes: "👀",       fire: "🔥",  check: "✅",
    x: "❌",       pray: "🙏",       rocket: "🚀", thinking: "🤔",
  };

  function handleEmojiPick(e) {
    dispatch("react", { messageId, emoji: e.detail });
    showEmojiPicker = false;
  }
</script>

<!-- Message row -->
<div
  class="group relative flex gap-3 px-4 py-1 transition-colors duration-100 hover:bg-shell-surface
    {isPinned ? 'border-l-2 border-shell-warn bg-shell-warn/5' : ''}"
  role="article"
  aria-label="Message from {sender}"
>
  <!-- Avatar -->
  <div class="mt-0.5 flex-shrink-0">
    <div
      class="grid h-9 w-9 place-content-center rounded-lg text-xs font-bold text-white {avatarBg}"
      aria-hidden="true"
    >
      {sender?.charAt(0)?.toUpperCase() || "?"}
    </div>
  </div>

  <!-- Content -->
  <div class="min-w-0 flex-1">
    <!-- Sender + timestamp row -->
    <div class="mb-0.5 flex items-baseline gap-2">
      <span class="text-sm font-semibold {isOwn ? 'text-shell-accentText' : 'text-shell-ink'}">
        {sender}
        {#if isOwn}<span class="ml-0.5 text-xs font-normal text-shell-subtle">(you)</span>{/if}
      </span>
      <time class="text-xs text-shell-subtle" datetime={timestamp} title={formatDate(timestamp)}>
        {formatTime(timestamp)}
      </time>
      {#if isPinned}
        <span class="flex items-center gap-1 text-xs font-medium text-shell-warn">
          <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
            <path d="M5 5a2 2 0 012-2h6a2 2 0 012 2v2.5l1.5 1.5V14h-3v5l-2-1-2 1v-5H5V9l1.5-1.5V5z"/>
          </svg>
          pinned
        </span>
      {/if}
      {#if isEdited}
        <span class="text-xs italic text-shell-subtle">(edited)</span>
      {/if}
    </div>

    <!-- Message body -->
    <div class="text-sm leading-relaxed text-shell-ink break-words">
      {#if content === "..."}
        <span class="italic text-shell-subtle">Decrypting...</span>
      {:else}
        <MarkdownText text={content} />
      {/if}
    </div>

    <!-- Link previews (max 2) -->
    {#if urls.length > 0 && content !== "..."}
      {#each urls.slice(0, 2) as url}
        <LinkPreview {url} />
      {/each}
    {/if}

    <!-- Reactions -->
    {#if reactions.length > 0}
      <div class="mt-1.5 flex flex-wrap gap-1" role="group" aria-label="Reactions">
        {#each reactions as r}
          <button
            class="flex items-center gap-1 rounded-full border border-shell-border bg-shell-elevated px-2 py-0.5 text-xs transition-colors hover:border-shell-accent hover:bg-shell-mention"
            on:click={() => dispatch("react", { messageId, emoji: r.emoji })}
            title="React with {r.emoji}"
          >
            <span>{emojiMap[r.emoji] || r.emoji}</span>
            <span class="text-shell-muted">{r.count}</span>
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Hover action toolbar -->
  <div
    class="absolute -top-3.5 right-3 flex items-center gap-0.5 rounded-lg border border-shell-border bg-shell-elevated px-1 py-0.5 shadow-panel opacity-0 transition-opacity duration-100 group-hover:opacity-100"
    role="toolbar"
    aria-label="Message actions"
  >
    <!-- React -->
    <div class="relative">
      <button
        on:click={() => (showEmojiPicker = !showEmojiPicker)}
        class="rounded p-1.5 text-shell-subtle transition-colors hover:bg-shell-surface hover:text-shell-ink"
        title="Add reaction"
        aria-label="Add reaction"
        aria-expanded={showEmojiPicker}
      >
        <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </button>
      <EmojiPicker bind:visible={showEmojiPicker} on:pick={handleEmojiPick} />
    </div>

    <!-- Reply in thread -->
    <button
      on:click={() => dispatch("thread", { messageId })}
      class="rounded p-1.5 text-shell-subtle transition-colors hover:bg-shell-surface hover:text-shell-ink"
      title="Reply in thread"
      aria-label="Reply in thread"
    >
      <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
      </svg>
    </button>

    <!-- Pin -->
    <button
      on:click={() => dispatch("pin", { messageId })}
      class="rounded p-1.5 text-shell-subtle transition-colors hover:bg-shell-surface hover:text-shell-ink"
      title="{isPinned ? 'Unpin' : 'Pin'} message"
      aria-label="{isPinned ? 'Unpin' : 'Pin'} message"
    >
      <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
      </svg>
    </button>

    <!-- Delete (own messages only) -->
    {#if isOwn}
      <button
        on:click={() => dispatch("delete", { messageId })}
        class="rounded p-1.5 text-shell-subtle transition-colors hover:bg-shell-dangerBg hover:text-shell-danger"
        title="Delete message"
        aria-label="Delete message"
      >
        <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
        </svg>
      </button>
    {/if}
  </div>
</div>
