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
  const urlRegex = /https?:\/\/[^\s<]+/g;
  $: urls = (content || "").match(urlRegex) || [];

  const emojiMap = {
    thumbsup: "\u{1F44D}", thumbsdown: "\u{1F44E}", heart: "\u{2764}\u{FE0F}",
    joy: "\u{1F602}", tada: "\u{1F389}", eyes: "\u{1F440}", fire: "\u{1F525}",
    check: "\u{2705}", x: "\u{274C}", pray: "\u{1F64F}", rocket: "\u{1F680}",
    thinking: "\u{1F914}",
  };

  function formatTime(ts) {
    if (!ts) return "";
    try { return new Date(ts).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }); }
    catch { return ts; }
  }

  function handleEmojiPick(e) {
    dispatch("react", { messageId, emoji: e.detail });
    showEmojiPicker = false;
  }
</script>

<div class="group relative flex gap-3 px-4 py-1.5 hover:bg-slate-50 transition {isPinned ? 'border-l-2 border-yellow-400 bg-yellow-50/50' : ''}">
  <div class="mt-0.5 flex-shrink-0">
    <div class="grid h-8 w-8 place-content-center rounded-lg {isOwn ? 'bg-teal-100 text-teal-700' : 'bg-slate-200 text-slate-500'} text-xs font-bold">
      {sender?.charAt(0)?.toUpperCase() || "?"}
    </div>
  </div>

  <div class="flex-1 min-w-0">
    <div class="flex items-baseline gap-2">
      <span class="text-sm font-bold text-slate-900">{sender}</span>
      <span class="text-xs text-slate-400">{formatTime(timestamp)}</span>
      {#if isPinned}<span class="text-xs text-yellow-600">pinned</span>{/if}
      {#if isEdited}<span class="text-xs text-slate-400 italic">(edited)</span>{/if}
    </div>
    <div class="text-sm text-slate-700 break-words"><MarkdownText text={content} /></div>

    {#if urls.length > 0}
      {#each urls.slice(0, 3) as url}
        <LinkPreview {url} />
      {/each}
    {/if}

    {#if reactions.length > 0}
      <div class="mt-1 flex flex-wrap gap-1">
        {#each reactions as r}
          <button
            class="flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-xs hover:bg-slate-100 transition"
            on:click={() => dispatch("react", { messageId, emoji: r.emoji })}
          >
            <span>{emojiMap[r.emoji] || r.emoji}</span>
            <span class="text-slate-500">{r.count}</span>
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Hover actions -->
  <div class="absolute right-2 -top-3 flex items-center gap-0.5 rounded-lg border border-slate-200 bg-white px-1 py-0.5 shadow-sm opacity-0 group-hover:opacity-100 transition">
    <div class="relative">
      <button on:click={() => (showEmojiPicker = !showEmojiPicker)}
        class="rounded p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600" title="React">
        <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </button>
      <EmojiPicker bind:visible={showEmojiPicker} on:pick={handleEmojiPick} />
    </div>
    <button on:click={() => dispatch("thread", { messageId })}
      class="rounded p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600" title="Reply in thread">
      <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
      </svg>
    </button>
    <button on:click={() => dispatch("pin", { messageId })}
      class="rounded p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600" title="Pin">
      <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
      </svg>
    </button>
    {#if isOwn}
      <button on:click={() => dispatch("delete", { messageId })}
        class="rounded p-1 text-slate-400 hover:bg-red-100 hover:text-red-600" title="Delete">
        <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
        </svg>
      </button>
    {/if}
  </div>
</div>
