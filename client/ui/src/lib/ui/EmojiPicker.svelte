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
    class="absolute bottom-full right-0 z-50 mb-1.5 rounded-xl border border-shell-border bg-shell-elevated p-2 shadow-modal animate-slide-up"
    role="tooltip"
    aria-label="Emoji picker"
  >
    <div class="grid grid-cols-6 gap-0.5">
      {#each commonEmojis as { emoji, label }}
        <button
          on:click={() => pick(label)}
          class="grid h-8 w-8 place-content-center rounded-lg text-base transition-colors hover:bg-shell-surface"
          title={label}
          aria-label="React with {label}"
        >{emoji}</button>
      {/each}
    </div>
  </div>
{/if}
