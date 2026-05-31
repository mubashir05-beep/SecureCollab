<script>
  import { createEventDispatcher } from "svelte";

  export let placeholder = "Say something friendly...";
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

<div class="relative bg-white border-t border-borderSoft/30 px-6 py-4">

  <!-- Mention autocomplete -->
  {#if showMentions && mentionCandidates.length > 0}
    <div
      class="absolute bottom-full left-6 right-6 z-50 mb-3 overflow-hidden rounded-[24px] border border-borderSoft bg-white shadow-2xl animate-slide-up"
      role="listbox"
    >
      {#each mentionCandidates as candidate, i}
        <button
          role="option"
          aria-selected={i === mentionIndex}
          class="flex w-full items-center gap-3 px-4 py-3 text-left text-[13px] transition-all
            {i === mentionIndex ? 'bg-sage/10 text-sage' : 'text-muted hover:bg-sidebar/50 hover:text-charcoal'}"
          on:mousedown|preventDefault={() => insertMention(candidate)}
        >
          <div class="w-7 h-7 flex-shrink-0 flex items-center justify-center rounded-lg font-bold text-[11px]
            {candidate.role === 'notify' ? 'bg-clay/10 text-clay' : 'bg-sidebar text-muted'}">
            {candidate.role === "notify" ? "@" : candidate.username?.charAt(0)?.toUpperCase() || "?"}
          </div>
          <span class="font-bold">
            {candidate.user_id.startsWith("@") ? candidate.user_id : `@${candidate.username}`}
          </span>
          {#if candidate.role && candidate.role !== "notify"}
            <span class="ml-auto text-[10px] font-bold text-muted/40 uppercase tracking-tight">{candidate.role}</span>
          {/if}
        </button>
      {/each}
    </div>
  {/if}

  <!-- Main Input Container -->
  <div class="bg-white rounded-[24px] border border-borderSoft shadow-xl shadow-stone-200/40 overflow-hidden focus-within:ring-4 focus-within:ring-sage/10 transition-all">
    <!-- Toolbar -->
    <div class="flex items-center gap-1 px-4 py-2 border-b border-ivory bg-sidebar/30">
      <button class="w-8 h-8 rounded-lg text-muted/60 hover:text-charcoal hover:bg-white transition-all flex items-center justify-center">
        <iconify-icon icon="lucide:bold"></iconify-icon>
      </button>
      <button class="w-8 h-8 rounded-lg text-muted/60 hover:text-charcoal hover:bg-white transition-all flex items-center justify-center">
        <iconify-icon icon="lucide:italic"></iconify-icon>
      </button>
      <button class="w-8 h-8 rounded-lg text-muted/60 hover:text-charcoal hover:bg-white transition-all flex items-center justify-center">
        <iconify-icon icon="lucide:link"></iconify-icon>
      </button>
      <div class="w-px h-4 bg-borderSoft/60 mx-1"></div>
      <button class="w-8 h-8 rounded-lg text-muted/60 hover:text-charcoal hover:bg-white transition-all flex items-center justify-center">
        <iconify-icon icon="lucide:list"></iconify-icon>
      </button>
      <button class="w-8 h-8 rounded-lg text-muted/60 hover:text-charcoal hover:bg-white transition-all flex items-center justify-center ml-auto">
        <iconify-icon icon="lucide:smile"></iconify-icon>
      </button>
    </div>

    <div class="relative">
      <textarea
        bind:this={textarea}
        bind:value={text}
        on:keydown={handleKeydown}
        on:input={handleInput}
        {placeholder}
        {disabled}
        class="w-full px-5 py-4 bg-transparent border-none text-[14px] font-medium text-charcoal outline-none resize-none h-24 placeholder:text-muted/30 leading-relaxed custom-scrollbar"
      ></textarea>

      <!-- Bottom Actions -->
      <div class="absolute bottom-3 right-3 flex items-center gap-2">
        <button
          class="w-10 h-10 rounded-2xl bg-sidebar text-muted hover:text-charcoal hover:bg-white border border-borderSoft/50 transition-all flex items-center justify-center"
          on:click={() => dispatch("attach")}
        >
          <iconify-icon icon="lucide:plus" class="text-xl"></iconify-icon>
        </button>

        <button
          on:click={send}
          disabled={!text.trim() || disabled}
          class="px-6 py-2.5 bg-clay text-white rounded-2xl text-[13px] font-bold shadow-lg shadow-clay/20 hover:scale-[1.03] active:scale-95 transition-all flex items-center gap-2 disabled:opacity-30 disabled:scale-100"
        >
          <span>Send</span>
          <iconify-icon icon="lucide:send" class="text-sm"></iconify-icon>
        </button>
      </div>
    </div>
  </div>

  <div class="mt-3 flex items-center justify-between px-2">
    <p class="text-[11px] font-bold text-muted/40 uppercase tracking-widest">
      Shift + Enter for new line
    </p>
    <div class="flex items-center gap-3">
      <iconify-icon icon="lucide:video" class="text-muted/40 text-lg cursor-pointer hover:text-sage transition-colors"></iconify-icon>
      <iconify-icon icon="lucide:mic" class="text-muted/40 text-lg cursor-pointer hover:text-clay transition-colors"></iconify-icon>
    </div>
  </div>
</div>
