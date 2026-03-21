<script>
  import { createEventDispatcher } from "svelte";
  import MessageBubble from "./MessageBubble.svelte";
  import MessageInput from "./MessageInput.svelte";

  export let visible = false;
  export let parentMessage = null;
  export let replies = [];
  export let getDecrypted = (m) => "[encrypted]";

  const dispatch = createEventDispatcher();

  function close() {
    dispatch("close");
  }
</script>

{#if visible && parentMessage}
  <div class="flex w-80 flex-shrink-0 flex-col border-l border-slate-200 bg-white">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-slate-200 px-4 py-3">
      <div>
        <h3 class="text-sm font-bold text-slate-900">Thread</h3>
        <p class="text-xs text-slate-400">{replies.length} {replies.length === 1 ? "reply" : "replies"}</p>
      </div>
      <button on:click={close} class="rounded p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600">
        <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Parent message -->
    <div class="border-b border-slate-100 bg-slate-50">
      <MessageBubble
        sender={parentMessage.sender_user_id}
        content={getDecrypted(parentMessage)}
        timestamp={parentMessage.created_at}
      />
    </div>

    <!-- Replies -->
    <div class="flex-1 overflow-y-auto">
      {#if replies.length === 0}
        <p class="p-4 text-center text-sm text-slate-400">No replies yet.</p>
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
  </div>
{/if}
