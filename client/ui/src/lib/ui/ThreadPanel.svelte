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
    class="w-96 h-full flex flex-col border-l border-borderSoft/50 bg-sidebar overflow-hidden animate-fade-in"
    aria-label="Thread panel"
  >
    <!-- Header -->
    <div class="h-[72px] flex items-center justify-between px-6 border-b border-borderSoft/50">
      <div class="flex flex-col">
        <h3 class="text-[15px] font-bold text-charcoal tracking-tight">Message Thread</h3>
        <span class="text-[11px] font-bold text-muted/40 uppercase tracking-widest">
          {replies.length} {replies.length === 1 ? "Reply" : "Replies"}
        </span>
      </div>
      <button
        on:click={() => dispatch("close")}
        class="w-10 h-10 flex items-center justify-center rounded-xl hover:bg-white text-muted hover:text-clay transition-all"
      >
        <iconify-icon icon="lucide:x" class="text-xl"></iconify-icon>
      </button>
    </div>

    <!-- Parent message (highlighted) -->
    <div class="border-b border-borderSoft/30 bg-white/40 p-1">
      <MessageBubble
        sender={parentMessage.sender_user_id}
        content={getDecrypted(parentMessage)}
        timestamp={parentMessage.created_at}
        isPinned={true} 
      />
    </div>

    <!-- Replies list -->
    <div class="flex-1 overflow-y-auto custom-scrollbar p-2" role="feed">
      {#if replies.length === 0}
        <div class="flex flex-col items-center justify-center gap-4 py-20 text-center px-8">
          <div class="w-16 h-16 rounded-3xl bg-sidebar flex items-center justify-center text-muted/30 border border-borderSoft/50">
            <iconify-icon icon="lucide:message-square" class="text-3xl"></iconify-icon>
          </div>
          <div class="space-y-1">
            <p class="text-[14px] font-bold text-charcoal">No replies yet</p>
            <p class="text-[12px] font-medium text-muted">Be the first to share your thoughts in this thread.</p>
          </div>
        </div>
      {:else}
        <div class="space-y-1">
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
    <div class="p-2 bg-white">
      <MessageInput
        placeholder="Reply to thread..."
        on:send={(e) => dispatch("reply", e.detail)}
      />
    </div>
  </aside>
{/if}
