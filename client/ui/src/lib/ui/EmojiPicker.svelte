<script>
  import { createEventDispatcher } from "svelte";

  export let visible = false;

  const dispatch = createEventDispatcher();

  const commonEmojis = [
    { emoji: "\u{1F44D}", label: "thumbsup" },
    { emoji: "\u{1F44E}", label: "thumbsdown" },
    { emoji: "\u{2764}\u{FE0F}", label: "heart" },
    { emoji: "\u{1F602}", label: "joy" },
    { emoji: "\u{1F389}", label: "tada" },
    { emoji: "\u{1F440}", label: "eyes" },
    { emoji: "\u{1F525}", label: "fire" },
    { emoji: "\u{2705}", label: "check" },
    { emoji: "\u{274C}", label: "x" },
    { emoji: "\u{1F64F}", label: "pray" },
    { emoji: "\u{1F680}", label: "rocket" },
    { emoji: "\u{1F914}", label: "thinking" },
  ];

  function pick(label) {
    dispatch("pick", label);
    visible = false;
  }
</script>

{#if visible}
  <div class="absolute bottom-full right-0 z-40 mb-1 rounded-xl border border-slate-200 bg-white p-2 shadow-lg">
    <div class="grid grid-cols-6 gap-1">
      {#each commonEmojis as { emoji, label }}
        <button
          on:click={() => pick(label)}
          class="grid h-8 w-8 place-content-center rounded-md text-lg hover:bg-slate-100"
          title={label}
        >{emoji}</button>
      {/each}
    </div>
  </div>
{/if}
