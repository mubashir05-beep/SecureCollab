<script>
  import { createEventDispatcher } from "svelte";

  export let visible = false;

  const dispatch = createEventDispatcher();

  const commonEmojis = [
    { emoji: "👍", label: "thumbsup"  },
    { emoji: "👎", label: "thumbsdown" },
    { emoji: "❤️", label: "heart"    },
    { emoji: "😂", label: "joy"       },
    { emoji: "🎉", label: "tada"      },
    { emoji: "👀", label: "eyes"      },
    { emoji: "🔥", label: "fire"      },
    { emoji: "✅", label: "check"     },
    { emoji: "❌", label: "x"         },
    { emoji: "🙏", label: "pray"      },
    { emoji: "🚀", label: "rocket"    },
    { emoji: "🤔", label: "thinking"  },
  ];

  function pick(label) {
    dispatch("pick", label);
    visible = false;
  }
</script>

{#if visible}
  <div
    class="absolute bottom-full right-0 z-50 mb-3 rounded-2xl border border-borderSoft bg-white p-2 shadow-2xl animate-slide-up overflow-hidden"
    role="tooltip"
  >
    <div class="grid grid-cols-6 gap-1">
      {#each commonEmojis as { emoji, label }}
        <button
          on:click={() => pick(label)}
          class="w-9 h-9 flex items-center justify-center rounded-xl text-lg transition-all hover:bg-sidebar hover:scale-110 active:scale-90"
          title={label}
        >
          {emoji}
        </button>
      {/each}
    </div>
  </div>
{/if}
