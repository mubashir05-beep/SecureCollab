<script>
  import { createEventDispatcher } from "svelte";
  import MessageBubble from "./MessageBubble.svelte";
  import MessageInput from "./MessageInput.svelte";

  export let visible = false;
  export let parentMessage = null;
  export let replies = [];
  export let getDecrypted = (m) => "[encrypted]";

  const dispatch = createEventDispatcher();
</script>

{#if visible && parentMessage}
  <aside
    class="flex w-80 flex-shrink-0 flex-col border-l border-shell-borderSub bg-shell-bg animate-fade-in"
    aria-label="Thread panel"
  >
    <!-- Header -->
    <div class="flex h-12 flex-shrink-0 items-center justify-between border-b border-shell-borderSub px-4">
      <div>
        <h3 class="text-sm font-bold text-shell-ink">Thread</h3>
        <p class="text-xs text-shell-subtle">
          {replies.length} {replies.length === 1 ? "reply" : "replies"}
        </p>
      </div>
      <button
        on:click={() => dispatch("close")}
        class="rounded-md p-1.5 text-shell-subtle transition-colors hover:bg-shell-surface hover:text-shell-ink"
        aria-label="Close thread panel"
      >
        <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Parent message (highlighted) -->
    <div class="border-b border-shell-borderSub bg-shell-surface/50">
      <MessageBubble
        sender={parentMessage.sender_user_id}
        content={getDecrypted(parentMessage)}
        timestamp={parentMessage.created_at}
      />
    </div>

    <!-- Replies list -->
    <div class="flex-1 overflow-y-auto" role="feed" aria-label="Thread replies">
      {#if replies.length === 0}
        <div class="flex flex-col items-center justify-center gap-2 p-8 text-center">
          <div class="grid h-10 w-10 place-content-center rounded-xl bg-shell-surface text-shell-subtle">
            <svg class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
            </svg>
          </div>
          <p class="text-sm font-medium text-shell-muted">No replies yet</p>
          <p class="text-xs text-shell-subtle">Be the first to reply in this thread.</p>
        </div>
      {:else}
        <div class="py-1">
          {#each replies as reply}
            <MessageBubble
              sender={reply.sender_user_id}
              content={getDecrypted(reply)}
              timestamp={reply.created_at}
            />
          {/each}
        </div>
      {/if}
    </div>

    <!-- Reply input -->
    <MessageInput
      placeholder="Reply in thread..."
      on:send={(e) => dispatch("reply", e.detail)}
    />
  </aside>
{/if}
