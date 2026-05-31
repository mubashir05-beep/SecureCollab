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

  // Premium Avatar Palette
  const avatarPalette = [
    "bg-sage", "bg-clay", "bg-charcoal",
    "bg-stone-400", "bg-[#A8A29E]", "bg-[#78716C]",
    "bg-[#44403C]", "bg-[#1C1917]",
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
  class="group relative flex gap-4 px-6 py-3 transition-all duration-200 hover:bg-sidebar/40 rounded-2xl mx-1
    {isPinned ? 'bg-sage/5 border-l-4 border-sage' : ''}"
  role="article"
>
  <!-- Avatar -->
  <div class="mt-1 flex-shrink-0">
    <div
      class="w-10 h-10 flex items-center justify-center rounded-2xl text-xs font-bold text-white shadow-md {avatarBg} transition-transform group-hover:scale-110"
      aria-hidden="true"
    >
      {sender?.charAt(0)?.toUpperCase() || "?"}
    </div>
  </div>

  <!-- Content -->
  <div class="min-w-0 flex-1">
    <!-- Sender + timestamp row -->
    <div class="mb-1.5 flex flex-wrap items-center gap-x-2 gap-y-1">
      <span class="text-[14px] font-bold tracking-tight {isOwn ? 'text-sage' : 'text-charcoal'}">
        {sender}
        {#if isOwn}<span class="ml-1 text-[10px] font-bold text-muted/40 uppercase tracking-widest">(you)</span>{/if}
      </span>
      <time class="text-[11px] font-bold text-muted/40 uppercase tracking-widest" datetime={timestamp} title={formatDate(timestamp)}>
        {formatTime(timestamp)}
      </time>
      
      {#if isPinned}
        <div class="flex items-center gap-1 px-2 py-0.5 rounded-lg bg-sage/10 text-sage text-[10px] font-bold uppercase tracking-tight">
          <iconify-icon icon="lucide:pin" class="text-[11px]"></iconify-icon>
          <span>Pinned</span>
        </div>
      {/if}
      
      {#if isEdited}
        <span class="text-[10px] font-bold text-muted/30 uppercase tracking-widest">(edited)</span>
      {/if}
    </div>

    <!-- Message body -->
    <div class="max-w-3xl text-[14px] font-medium leading-relaxed text-charcoal/90 break-words">
      {#if content === "..."}
        <span class="inline-flex items-center gap-2 px-3 py-1.5 rounded-xl bg-sidebar/50 text-[12px] font-bold text-muted animate-pulse border border-borderSoft/30">
          <iconify-icon icon="lucide:shield-check" class="text-sage"></iconify-icon>
          Secured Message...
        </span>
      {:else}
        <MarkdownText text={content} />
      {/if}
    </div>

    <!-- Link previews -->
    {#if urls.length > 0 && content !== "..."}
      <div class="mt-3 space-y-2">
        {#each urls.slice(0, 2) as url}
          <LinkPreview {url} />
        {/each}
      </div>
    {/if}

    <!-- Reactions -->
    {#if reactions.length > 0}
      <div class="mt-3 flex flex-wrap gap-2" role="group">
        {#each reactions as r}
          <button
            class="flex items-center gap-1.5 px-3 py-1 rounded-xl border border-borderSoft bg-white text-[12px] font-bold transition-all hover:border-sage hover:text-sage hover:bg-sage/5 shadow-sm"
            on:click={() => dispatch("react", { messageId, emoji: r.emoji })}
          >
            <span>{emojiMap[r.emoji] || r.emoji}</span>
            <span class="text-muted/60">{r.count}</span>
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Hover action toolbar -->
  <div
    class="absolute -top-4 right-6 flex items-center gap-1 p-1 rounded-2xl bg-white border border-borderSoft shadow-xl opacity-0 translate-y-2 transition-all duration-200 group-hover:opacity-100 group-hover:translate-y-0"
    role="toolbar"
  >
    <!-- React -->
    <div class="relative">
      <button
        on:click={() => (showEmojiPicker = !showEmojiPicker)}
        class="w-8 h-8 flex items-center justify-center rounded-xl text-muted hover:bg-sidebar hover:text-charcoal transition-colors"
        title="Add reaction"
      >
        <iconify-icon icon="lucide:smile" class="text-lg"></iconify-icon>
      </button>
      <EmojiPicker bind:visible={showEmojiPicker} on:pick={handleEmojiPick} />
    </div>

    <button
      on:click={() => dispatch("thread", { messageId })}
      class="w-8 h-8 flex items-center justify-center rounded-xl text-muted hover:bg-sidebar hover:text-charcoal transition-colors"
      title="Reply in thread"
    >
      <iconify-icon icon="lucide:message-square" class="text-lg"></iconify-icon>
    </button>

    <button
      on:click={() => dispatch("pin", { messageId })}
      class="w-8 h-8 flex items-center justify-center rounded-xl text-muted hover:bg-sidebar hover:text-sage transition-colors"
      title={isPinned ? 'Unpin' : 'Pin'}
    >
      <iconify-icon icon={isPinned ? "lucide:pin-off" : "lucide:pin"} class="text-lg"></iconify-icon>
    </button>

    {#if isOwn}
      <div class="w-px h-4 bg-borderSoft/60 mx-1"></div>
      <button
        on:click={() => dispatch("delete", { messageId })}
        class="w-8 h-8 flex items-center justify-center rounded-xl text-muted hover:bg-red-50 hover:text-red-500 transition-colors"
        title="Delete"
      >
        <iconify-icon icon="lucide:trash-2" class="text-lg"></iconify-icon>
      </button>
    {/if}
  </div>
</div>
