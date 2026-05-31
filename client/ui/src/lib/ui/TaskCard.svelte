<script>
  /** @type {"todo" | "in-progress" | "review" | "done"} */
  export let status = "todo";
  export let title = "";
  export let priority = "medium"; // low, medium, high
  export let assignee = null; // { name, avatar }
  export let dueDate = "";
  export let taskId = "";

  const priorityColors = {
    low: "text-sage bg-sage/10",
    medium: "text-clay bg-clay/10",
    high: "text-red-500 bg-red-500/10",
  };

  const statusLabels = {
    todo: "To Do",
    "in-progress": "Doing",
    review: "Review",
    done: "Done",
  };
</script>

<div class="group p-5 rounded-[24px] bg-white border border-borderSoft shadow-sm hover:shadow-xl hover:shadow-stone-200/40 hover:-translate-y-1 transition-all cursor-pointer">
  <div class="flex justify-between items-start mb-3">
    <span class="text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-lg {priorityColors[priority]}">
      {priority} Priority
    </span>
    <span class="text-[10px] font-bold text-muted/30 uppercase tracking-widest">{taskId}</span>
  </div>

  <h4 class="text-[14px] font-bold text-charcoal leading-snug mb-4 group-hover:text-sage transition-colors">
    {title}
  </h4>

  <div class="flex items-center justify-between pt-4 border-t border-borderSoft/50">
    <div class="flex items-center gap-2">
      {#if assignee}
        <div class="w-6 h-6 rounded-lg bg-sidebar flex items-center justify-center text-[10px] font-bold text-muted border border-borderSoft/50 overflow-hidden">
          {#if assignee.avatar}
            <img src={assignee.avatar} alt={assignee.name} class="w-full h-full object-cover" />
          {:else}
            {assignee.name.charAt(0)}
          {/if}
        </div>
        <span class="text-[11px] font-bold text-muted/60">{assignee.name}</span>
      {:else}
        <div class="w-6 h-6 rounded-lg border-2 border-dashed border-borderSoft flex items-center justify-center text-muted/30">
          <iconify-icon icon="lucide:user-plus" class="text-[10px]"></iconify-icon>
        </div>
        <span class="text-[11px] font-bold text-muted/30 italic">Unassigned</span>
      {/if}
    </div>

    {#if dueDate}
      <div class="flex items-center gap-1.5 text-muted/40 group-hover:text-clay transition-colors">
        <iconify-icon icon="lucide:calendar" class="text-[12px]"></iconify-icon>
        <span class="text-[11px] font-bold">{dueDate}</span>
      </div>
    {/if}
  </div>
</div>
