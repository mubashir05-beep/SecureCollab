<script>
  import TaskCard from "./TaskCard.svelte";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  export let tasks = [
    { id: "SC-101", title: "Update social media banners for Q2 launch", priority: "high", status: "todo", dueDate: "Apr 24", assignee: { name: "Marcus" } },
    { id: "SC-105", title: "Review final marketing graphics and approve", priority: "medium", status: "todo", dueDate: "Apr 25", assignee: { name: "Sarah" } },
    { id: "SC-102", title: "Refactor encryption service for better performance", priority: "high", status: "in-progress", dueDate: "Apr 22", assignee: { name: "You" } },
    { id: "SC-108", title: "Implement new dashboard widgets", priority: "low", status: "done", dueDate: "Apr 20", assignee: { name: "James" } },
  ];

  $: todoTasks = tasks.filter(t => t.status === "todo");
  $: inProgressTasks = tasks.filter(t => t.status === "in-progress");
  $: doneTasks = tasks.filter(t => t.status === "done");

  $: progress = Math.round((doneTasks.length / tasks.length) * 100) || 0;
</script>

<aside class="w-[380px] h-full bg-sidebar border-l border-borderSoft/50 overflow-hidden flex flex-col animate-fade-in">
  <!-- Header -->
  <div class="h-[72px] flex items-center justify-between px-6 border-b border-borderSoft/50">
    <div class="flex items-center gap-2">
      <iconify-icon icon="lucide:layout-list" class="text-sage text-xl"></iconify-icon>
      <h2 class="text-[15px] font-bold text-charcoal tracking-tight">Productivity</h2>
    </div>
    <button on:click={() => dispatch("close")} class="w-8 h-8 flex items-center justify-center rounded-xl hover:bg-white text-muted hover:text-clay transition-all">
      <iconify-icon icon="lucide:x" class="text-lg"></iconify-icon>
    </button>
  </div>
  
  <div class="flex-1 overflow-y-auto custom-scrollbar p-6 space-y-8">
    <!-- Progress Overview -->
    <div class="p-6 rounded-[32px] bg-white border border-borderSoft shadow-sm">
      <div class="flex items-center justify-between mb-4">
        <div class="flex flex-col">
          <span class="text-[10px] font-bold uppercase tracking-widest text-muted/60 mb-0.5">Overall Progress</span>
          <span class="text-[20px] font-bold text-charcoal">{progress}%</span>
        </div>
        <div class="w-12 h-12 rounded-2xl bg-sage/10 flex items-center justify-center text-sage">
          <iconify-icon icon="lucide:trending-up" class="text-2xl"></iconify-icon>
        </div>
      </div>
      <div class="w-full h-2.5 bg-sidebar rounded-full overflow-hidden">
        <div class="h-full bg-sage rounded-full shadow-sm shadow-sage/40 transition-all duration-1000" style="width: {progress}%"></div>
      </div>
      <p class="mt-4 text-[11px] font-bold text-muted/40 uppercase tracking-widest text-center">
        {doneTasks.length} of {tasks.length} tasks completed
      </p>
    </div>

    <!-- Search / Filter -->
    <div class="relative">
      <iconify-icon icon="lucide:search" class="absolute left-4 top-1/2 -translate-y-1/2 text-muted/40 text-lg"></iconify-icon>
      <input 
        type="text" 
        placeholder="Search tasks..." 
        class="w-full pl-11 pr-4 py-3.5 rounded-2xl bg-white border border-borderSoft text-[13px] font-medium outline-none focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all"
      />
    </div>

    <!-- To Do -->
    <div class="space-y-4">
      <div class="flex items-center justify-between px-2">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-clay"></div>
          <h3 class="text-[12px] font-bold uppercase tracking-widest text-muted">To Do</h3>
        </div>
        <span class="px-2 py-0.5 rounded-lg bg-clay/10 text-clay text-[10px] font-bold">{todoTasks.length}</span>
      </div>
      
      <div class="space-y-4">
        {#each todoTasks as task}
          <TaskCard 
            taskId={task.id}
            title={task.title}
            priority={task.priority}
            assignee={task.assignee}
            dueDate={task.dueDate}
            status={task.status}
          />
        {/each}

        <button class="w-full py-4 rounded-[24px] border-2 border-dashed border-borderSoft text-muted/40 hover:border-sage hover:text-sage hover:bg-sage/5 transition-all flex items-center justify-center gap-2 text-[13px] font-bold">
          <iconify-icon icon="lucide:plus-circle" class="text-lg"></iconify-icon>
          <span>Create New Task</span>
        </button>
      </div>
    </div>

    <!-- In Progress -->
    <div class="space-y-4">
      <div class="flex items-center justify-between px-2">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-sage"></div>
          <h3 class="text-[12px] font-bold uppercase tracking-widest text-muted">In Progress</h3>
        </div>
        <span class="px-2 py-0.5 rounded-lg bg-sage/10 text-sage text-[10px] font-bold">{inProgressTasks.length}</span>
      </div>
      
      <div class="space-y-4">
        {#each inProgressTasks as task}
          <TaskCard 
            taskId={task.id}
            title={task.title}
            priority={task.priority}
            assignee={task.assignee}
            dueDate={task.dueDate}
            status={task.status}
          />
        {/each}
      </div>
    </div>
  </div>
</aside>
