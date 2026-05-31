<script>
  import { onDestroy } from "svelte";
  import Sidebar from "./lib/ui/Sidebar.svelte";
  import TaskPanel from "./lib/ui/TaskPanel.svelte";
  import TopBar from "./lib/ui/TopBar.svelte";
  import MessageBubble from "./lib/ui/MessageBubble.svelte";
  import MessageInput from "./lib/ui/MessageInput.svelte";
  import AuthModal from "./lib/ui/AuthModal.svelte";
  import CreateWorkspaceModal from "./lib/ui/CreateWorkspaceModal.svelte";
  import CreateChannelModal from "./lib/ui/CreateChannelModal.svelte";
  import ThreadPanel from "./lib/ui/ThreadPanel.svelte";
  import InviteModal from "./lib/ui/InviteModal.svelte";
  import MembersPanel from "./lib/ui/MembersPanel.svelte";
  import { auth, isAuthenticated } from "./lib/authStore.js";
  import { keyStore } from "./lib/keyStore.js";
  import * as api from "./lib/api.js";
  import { encryptMessage, decryptMessage } from "./lib/crypto.js";

  // --- State ---
  let showAuth = false;
  let showCreateWs = false;
  let showCreateCh = false;
  let showInvite = false;
  let showMembers = false;
  let showTasks = true; // Default to true on desktop
  let onboardingInviteCode = "";
  let onboardingError = "";
  let authModal, wsModal, chModal, inviteModal, membersPanel;

  let workspaces = [];
  let channels = [];
  let directMessages = [];
  let members = [];
  let activeWorkspace = null;
  let activeChannel = null;
  let messages = [];
  let wsConn = null;
  let wsReconnectTimer = null;
  let wsGeneration = 0;
  let connectionState = "disconnected";
  let lastLiveEventAt = null;
  let workspacesLoaded = false;
  let messageViewport;

  let unreadByWorkspace = {};
  let unreadByChannel = {};

  // Rich messaging state
  let threadParent = null;
  let threadReplies = [];
  let showThread = false;
  let messageReactions = {}; // messageId -> ReactionSummary[]

  // --- Auth ---
  $: currentUser = $isAuthenticated ? { username: $auth.username } : null;
  $: isAdmin = members.some(m => m.user_id === $auth.userId && (m.role === "owner" || m.role === "admin"));

  $: if ($isAuthenticated) {
    bootstrapAndLoad();
  }

  $: if (!$isAuthenticated) {
    cleanup();
  }

  async function bootstrapAndLoad() {
    await keyStore.bootstrap($auth.token);
    loadWorkspaces();
  }

  $: if (activeWorkspace && $isAuthenticated) {
    loadChannels(activeWorkspace.id);
    loadMembers(activeWorkspace.id);
  }

  $: if (activeChannel && $isAuthenticated) {
    loadInbox();
    connectWs();
  }

  function cleanup() {
    if (wsReconnectTimer) {
      clearTimeout(wsReconnectTimer);
      wsReconnectTimer = null;
    }
    wsGeneration += 1;
    if (wsConn) { wsConn.close(); wsConn = null; }
    workspaces = [];
    channels = [];
    messages = [];
    activeWorkspace = null;
    activeChannel = null;
    connectionState = "disconnected";
    workspacesLoaded = false;
  }

  onDestroy(() => {
    cleanup();
  });

  async function handleAuth(e) {
    const { mode, username, email, password } = e.detail;
    try {
      let result;
      if (mode === "register") {
        result = await api.register(username, email, password);
      } else {
        result = await api.login(username, password);
      }
      auth.login(result.access_token, result.username || username, result.user_id);
      showAuth = false;
    } catch (err) {
      authModal?.setError(err.message);
    }
  }

  // --- Workspaces ---
  async function loadWorkspaces() {
    try {
      const res = await api.listWorkspaces($auth.token);
      workspaces = (res.workspaces || []).map((ws) => ({
        ...ws,
        unreadCount: unreadByWorkspace[ws.id] || 0,
      }));
      if (activeWorkspace) {
        activeWorkspace = workspaces.find((ws) => ws.id === activeWorkspace.id) || workspaces[0] || null;
      } else if (workspaces.length > 0) {
        activeWorkspace = workspaces[0];
      }
    } catch { /* silent */ }
    workspacesLoaded = true;
  }

  async function handleCreateWorkspace(e) {
    const { name, description } = e.detail;
    try {
      const ws = await api.createWorkspace($auth.token, name, description);
      showCreateWs = false;
      wsModal?.reset();
      await loadWorkspaces();
      activeWorkspace = workspaces.find(w => w.id === ws.id) || workspaces[0];
    } catch (err) {
      wsModal?.setError(err.message);
    }
  }

  // --- Channels ---
  async function loadChannels(workspaceId) {
    try {
      const res = await api.listChannels($auth.token, workspaceId);
      channels = (res.channels || []).map((ch) => ({
        ...ch,
        unreadCount: unreadByChannel[ch.id] || 0,
      }));
      if (activeChannel) {
        activeChannel = channels.find((ch) => ch.id === activeChannel.id) || channels[0] || null;
      } else if (channels.length > 0) {
        activeChannel = channels[0];
      }
    } catch { /* silent */ }
  }

  async function handleCreateChannel(e) {
    const { name, topic, isPrivate } = e.detail;
    if (!activeWorkspace) return;
    try {
      const ch = await api.createChannel($auth.token, activeWorkspace.id, name, "", topic, isPrivate);
      showCreateCh = false;
      chModal?.reset();
      await loadChannels(activeWorkspace.id);
      activeChannel = channels.find(c => c.id === ch.id) || channels[0];
    } catch (err) {
      chModal?.setError(err.message);
    }
  }

  // --- Members ---
  async function loadMembers(workspaceId) {
    try {
      const res = await api.listWorkspaceMembers($auth.token, workspaceId);
      members = res.members || [];
    } catch { /* silent */ }
  }

  // --- Messages ---
  async function loadInbox() {
    try {
      const res = await api.getInbox($auth.token);
      messages = res.messages || [];
      scrollMessagesToBottom();
    } catch { /* silent */ }
  }

  function scrollMessagesToBottom() {
    requestAnimationFrame(() => {
      if (messageViewport) {
        messageViewport.scrollTop = messageViewport.scrollHeight;
      }
    });
  }

  function updateUnreadCounts(channelId, nextCount) {
    if (!channelId) return;
    unreadByChannel = { ...unreadByChannel, [channelId]: nextCount };
    channels = channels.map((channel) => (
      channel.id === channelId ? { ...channel, unreadCount: nextCount } : channel
    ));
    if (activeWorkspace) {
      const workspaceUnread = channels.reduce((total, channel) => total + (unreadByChannel[channel.id] || 0), 0);
      unreadByWorkspace = { ...unreadByWorkspace, [activeWorkspace.id]: workspaceUnread };
      workspaces = workspaces.map((workspace) => (
        workspace.id === activeWorkspace.id ? { ...workspace, unreadCount: workspaceUnread } : workspace
      ));
    }
  }

  function clearUnreadForChannel(channelId) {
    if (!channelId) return;
    updateUnreadCounts(channelId, 0);
  }

  function markLiveActivity() {
    lastLiveEventAt = new Date().toISOString();
  }

  function handleIncomingMessage(envelope) {
    if (!envelope || (envelope.id && messages.some((msg) => msg.id === envelope.id))) {
      return;
    }

    messages = [...messages, envelope];
    markLiveActivity();

    if (envelope.channel_id) {
      const nextUnread = activeChannel?.id === envelope.channel_id ? 0 : (unreadByChannel[envelope.channel_id] || 0) + 1;
      updateUnreadCounts(envelope.channel_id, nextUnread);
      if (activeChannel?.id === envelope.channel_id) {
        clearUnreadForChannel(envelope.channel_id);
      }
    }

    scrollMessagesToBottom();
  }

  function connectWs() {
    const generation = ++wsGeneration;
    if (wsReconnectTimer) {
      clearTimeout(wsReconnectTimer);
      wsReconnectTimer = null;
    }
    if (wsConn) { wsConn.close(); wsConn = null; }

    connectionState = "connecting";
    wsConn = api.connectInboxWs($auth.token, handleIncomingMessage);
    wsConn.onopen = () => {
      if (generation !== wsGeneration) return;
      connectionState = "connected";
      markLiveActivity();
    };
    wsConn.onerror = () => {
      if (generation !== wsGeneration) return;
      connectionState = "reconnecting";
    };
    wsConn.onclose = () => {
      if (generation !== wsGeneration) return;
      wsConn = null;
      if (!$isAuthenticated) {
        connectionState = "disconnected";
        return;
      }
      connectionState = "reconnecting";
      wsReconnectTimer = setTimeout(() => {
        if (generation === wsGeneration && $isAuthenticated && activeChannel) {
          connectWs();
        }
      }, 2500);
    };
  }

  async function handleSend(e) {
    const text = e.detail;
    try {
      const keys = $keyStore;
      const payload = await encryptMessage(keys.privateKey, keys.publicKey, text);
      const envelope = await api.sendMessage($auth.token, $auth.userId, payload.ciphertext_b64, payload.nonce_b64, activeChannel?.id || "");
      handleIncomingMessage(envelope);
    } catch (err) {
      console.error("Send failed:", err);
    }
  }

  // --- Decrypt ---
  const decryptCache = {};
  function getDecrypted(msg) {
    const cacheKey = msg.id || `${msg.ciphertext_b64}:${msg.nonce_b64}`;
    if (decryptCache[cacheKey] !== undefined) return decryptCache[cacheKey];
    decryptCache[cacheKey] = "...";
    const keys = $keyStore;
    decryptMessage(keys.privateKey, keys.publicKey, msg.ciphertext_b64, msg.nonce_b64).then((text) => {
      decryptCache[cacheKey] = text;
      messages = [...messages];
    });
    return decryptCache[cacheKey];
  }

  // --- Reactions ---
  async function handleReaction(e) {
    const { messageId, emoji } = e.detail;
    try {
      await api.addReaction($auth.token, messageId, emoji);
      const res = await api.getReactions($auth.token, messageId);
      messageReactions[messageId] = res.reactions || [];
      messageReactions = messageReactions;
      markLiveActivity();
    } catch { /* silent */ }
  }

  // --- Threads ---
  async function handleOpenThread(e) {
    const { messageId } = e.detail;
    threadParent = messages.find(m => m.id === messageId) || null;
    if (!threadParent) return;
    showThread = true;
    try {
      const res = await api.getThreadReplies($auth.token, messageId);
      threadReplies = res.replies || [];
    } catch { threadReplies = []; }
  }

  async function handleThreadReply(e) {
    const text = e.detail;
    if (!threadParent) return;
    try {
      const keys = $keyStore;
      const payload = await encryptMessage(keys.privateKey, keys.publicKey, text);
      await api.postThreadReply($auth.token, threadParent.id, $auth.username, payload.ciphertext_b64, payload.nonce_b64);
      const res = await api.getThreadReplies($auth.token, threadParent.id);
      threadReplies = res.replies || [];
      markLiveActivity();
    } catch (err) {
      console.error("Thread reply failed:", err);
    }
  }

  function handleCloseThread() {
    showThread = false;
    threadParent = null;
    threadReplies = [];
  }

  // --- Pin ---
  async function handlePin(e) {
    const { messageId } = e.detail;
    try {
      await api.pinMessage($auth.token, messageId, activeChannel?.id || "");
      markLiveActivity();
    } catch { /* silent */ }
  }

  // --- Delete ---
  async function handleDelete(e) {
    const { messageId } = e.detail;
    try {
      await api.deleteMessageApi($auth.token, messageId);
      messages = messages.filter(m => m.id !== messageId);
      markLiveActivity();
    } catch { /* silent */ }
  }

  // --- Member management ---
  async function handleAddMember(e) {
    const { userId, role } = e.detail;
    if (!activeWorkspace) return;
    try {
      await api.addWorkspaceMember($auth.token, activeWorkspace.id, userId, role);
      await loadMembers(activeWorkspace.id);
    } catch (err) {
      membersPanel?.setError(err.message);
    }
  }

  async function handleRemoveMember(e) {
    const { userId } = e.detail;
    if (!activeWorkspace) return;
    try {
      await api.removeWorkspaceMember($auth.token, activeWorkspace.id, userId);
      await loadMembers(activeWorkspace.id);
    } catch (err) {
      membersPanel?.setError(err.message);
    }
  }

  // --- Onboarding join ---
  async function handleOnboardingJoin() {
    if (!onboardingInviteCode.trim()) return;
    onboardingError = "";
    try {
      await api.joinWorkspaceByInvite($auth.token, onboardingInviteCode.trim());
      onboardingInviteCode = "";
      await loadWorkspaces();
    } catch (err) {
      onboardingError = err.message;
    }
  }

  // --- Logout ---
  function handleLogout() {
    keyStore.clear();
    auth.logout();
  }

  // --- Invite ---
  function handleOpenInvite() {
    showInvite = true;
    inviteModal?.reset();
  }

  async function handleJoinByInvite(e) {
    const code = e.detail;
    try {
      await api.joinWorkspaceByInvite($auth.token, code);
      showInvite = false;
      await loadWorkspaces();
    } catch (err) {
      inviteModal?.setError(err.message);
    }
  }

  // --- Nav ---
  function handleSelectWorkspace(e) {
    activeWorkspace = e.detail;
    activeChannel = null;
  }

  function handleSelectChannel(e) {
    activeChannel = e.detail;
    clearUnreadForChannel(activeChannel?.id);
  }

  function formatLiveActivity(ts) {
    if (!ts) return "Connecting";
    const delta = Date.now() - new Date(ts).getTime();
    if (Number.isNaN(delta) || delta < 1000) return "Live now";
    if (delta < 60_000) return `${Math.max(1, Math.floor(delta / 1000))}s ago`;
    if (delta < 3_600_000) return `${Math.floor(delta / 60_000)}m ago`;
    return `${Math.floor(delta / 3_600_000)}h ago`;
  }
</script>

<!-- ════════════════════════════════════════════
     LANDING — not authenticated
═══════════════════════════════════════════════ -->
{#if !$isAuthenticated}
  <main class="flex min-h-screen flex-col items-center justify-center bg-sidebar px-4 relative overflow-hidden">
    <!-- Decorative background elements -->
    <div class="absolute top-[-10%] left-[-5%] w-[40%] h-[40%] bg-sage/5 rounded-full blur-3xl"></div>
    <div class="absolute bottom-[-10%] right-[-5%] w-[40%] h-[40%] bg-clay/5 rounded-full blur-3xl"></div>

    <div class="w-full max-w-[480px] text-center animate-slide-up relative z-10">
      <!-- Logo mark -->
      <div class="mx-auto mb-8 flex h-24 w-24 items-center justify-center rounded-[32px] bg-sage shadow-2xl shadow-sage/30" aria-hidden="true">
        <span class="text-3xl font-bold text-white">SC</span>
      </div>

      <h1 class="mb-4 text-4xl font-bold text-charcoal tracking-tight">SecureCollab</h1>
      <p class="mb-8 text-lg text-muted font-medium max-w-sm mx-auto leading-relaxed">
        A premium-but-friendly workspace merging chat with project management.
      </p>

      <!-- Feature cards -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-10 text-left">
        <div class="p-4 rounded-2xl bg-white border border-borderSoft shadow-sm">
          <iconify-icon icon="lucide:shield-check" class="text-2xl text-sage mb-2"></iconify-icon>
          <h3 class="text-[13px] font-bold text-charcoal mb-1">Secure</h3>
          <p class="text-[11px] text-muted leading-snug">E2E encryption for all messages.</p>
        </div>
        <div class="p-4 rounded-2xl bg-white border border-borderSoft shadow-sm">
          <iconify-icon icon="lucide:layout" class="text-2xl text-clay mb-2"></iconify-icon>
          <h3 class="text-[13px] font-bold text-charcoal mb-1">Hybrid</h3>
          <p class="text-[11px] text-muted leading-snug">Slack ease with Jira power.</p>
        </div>
        <div class="p-4 rounded-2xl bg-white border border-borderSoft shadow-sm">
          <iconify-icon icon="lucide:zap" class="text-2xl text-sage mb-2"></iconify-icon>
          <h3 class="text-[13px] font-bold text-charcoal mb-1">Productive</h3>
          <p class="text-[11px] text-muted leading-snug">Task tracking built right in.</p>
        </div>
      </div>

      <button
        on:click={() => (showAuth = true)}
        class="w-full sm:w-64 rounded-2xl bg-sage py-4 text-base font-bold text-white shadow-lg shadow-sage/20 transition-all hover:scale-[1.03] active:scale-95"
      >
        Open Workspace
      </button>

      <p class="mt-8 text-[11px] font-bold text-muted uppercase tracking-[0.2em]">
        Free &middot; Open Source &middot; Privacy First
      </p>
    </div>
  </main>

  <AuthModal bind:this={authModal} visible={showAuth}
    on:auth={handleAuth} on:close={() => (showAuth = false)} />

<!-- ════════════════════════════════════════════
     ONBOARDING — authenticated, no workspaces
═══════════════════════════════════════════════ -->
{:else if workspacesLoaded && workspaces.length === 0}
  <main class="flex min-h-screen flex-col items-center justify-center bg-sidebar px-4">
    <div class="w-full max-w-[420px] animate-slide-up">
      <div class="p-10 rounded-[40px] bg-white border border-borderSoft shadow-xl text-center mb-6">
        <div class="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-sage shadow-lg shadow-sage/10" aria-hidden="true">
          <span class="text-xl font-bold text-white">SC</span>
        </div>
        <h1 class="mb-2 text-2xl font-bold text-charcoal">
          Welcome, {$auth.username}!
        </h1>
        <p class="mb-8 text-sm text-muted font-medium leading-relaxed">
          Your workspace is empty. Create a new one or join an existing team.
        </p>

        <div class="space-y-4">
          <button
            on:click={() => (showCreateWs = true)}
            class="w-full rounded-2xl bg-sage py-3.5 text-[14px] font-bold text-white shadow-lg shadow-sage/10 transition-all hover:scale-[1.02] active:scale-95"
          >
            Create Workspace
          </button>

          <div class="flex items-center gap-4 py-2">
            <div class="flex-1 h-px bg-borderSoft"></div>
            <span class="text-[10px] font-bold text-muted/60 uppercase tracking-widest">or</span>
            <div class="flex-1 h-px bg-borderSoft"></div>
          </div>

          <div class="flex gap-2">
            <input
              type="text"
              bind:value={onboardingInviteCode}
              placeholder="Invite Code..."
              class="flex-1 rounded-xl border border-borderSoft bg-sidebar/30 px-4 py-3 text-sm text-charcoal placeholder-muted/50 outline-none focus:border-sage focus:ring-4 focus:ring-sage/5 transition-all"
              on:keydown={(e) => e.key === "Enter" && handleOnboardingJoin()}
            />
            <button
              on:click={handleOnboardingJoin}
              disabled={!onboardingInviteCode.trim()}
              class="rounded-xl bg-clay px-5 py-3 text-[14px] font-bold text-white shadow-lg shadow-clay/10 transition-all hover:scale-[1.02] active:scale-95 disabled:opacity-30 disabled:scale-100"
            >
              Join
            </button>
          </div>

          {#if onboardingError}
            <p class="text-xs font-bold text-clay mt-2">{onboardingError}</p>
          {/if}
        </div>
      </div>

      <div class="text-center">
        <button
          on:click={handleLogout}
          class="text-[11px] font-bold text-muted uppercase tracking-widest hover:text-clay transition-colors"
        >
          Sign out
        </button>
      </div>
    </div>
  </main>

  <CreateWorkspaceModal bind:this={wsModal} visible={showCreateWs}
    on:create={handleCreateWorkspace} on:close={() => (showCreateWs = false)} />

<!-- ════════════════════════════════════════════
     MAIN SHELL — authenticated with workspaces
═══════════════════════════════════════════════ -->
{:else}
  <div class="flex h-screen overflow-hidden bg-ivory" role="application">
    <!-- Sidebar (Left) -->
    <Sidebar
      {workspaces} {channels} {directMessages} {activeWorkspace} {activeChannel} {currentUser}
      on:selectWorkspace={handleSelectWorkspace}
      on:selectChannel={handleSelectChannel}
      on:createWorkspace={() => (showCreateWs = true)}
      on:createChannel={() => (showCreateCh = true)}
      on:logout={handleLogout}
      on:invite={handleOpenInvite}
    />

    <!-- Main Content Area (Middle) -->
    <div class="flex flex-1 flex-col overflow-hidden bg-white border-r border-borderSoft/50 relative">
      <TopBar
        channelName={activeChannel?.name || ""}
        channelTopic={activeChannel?.topic || ""}
        memberCount={members.length}
        connectionState={connectionState}
        liveActivity={formatLiveActivity(lastLiveEventAt)}
        on:showMembers={() => (showMembers = !showMembers)}
        on:toggleTasks={() => (showTasks = !showTasks)}
      />

      <!-- Channel content -->
      {#if !activeChannel}
        <div class="flex flex-1 flex-col items-center justify-center p-12 text-center">
          <div class="w-20 h-20 rounded-3xl bg-sidebar flex items-center justify-center mb-6 border border-borderSoft">
            <iconify-icon icon="lucide:message-square-plus" class="text-4xl text-muted/40"></iconify-icon>
          </div>
          <h2 class="text-2xl font-bold text-charcoal mb-2">No Channel Selected</h2>
          <p class="text-muted text-[15px] max-w-sm mb-8 leading-relaxed">
            Select a channel from the sidebar or create a new one to start collaborating.
          </p>
          <button
            on:click={() => (showCreateCh = true)}
            class="px-8 py-3 rounded-2xl bg-sage text-white font-bold shadow-lg shadow-sage/10 transition-all hover:scale-[1.03] active:scale-95"
          >
            Create New Channel
          </button>
        </div>
      {:else}
        <!-- Message list -->
        <div bind:this={messageViewport} class="flex-1 overflow-y-auto custom-scrollbar px-6" role="log">
          {#if messages.length === 0}
            <div class="flex h-full flex-col items-center justify-center text-center opacity-40">
              <iconify-icon icon="lucide:messages-square" class="text-6xl mb-4"></iconify-icon>
              <p class="text-lg font-bold">Start the conversation</p>
              <p class="text-sm">Messages are secured with E2E encryption</p>
            </div>
          {:else}
            <div class="py-8 space-y-1">
              {#each messages as msg (msg.id || `${msg.ciphertext_b64}:${msg.nonce_b64}`)}
                <MessageBubble
                  sender={msg.sender_user_id}
                  content={getDecrypted(msg)}
                  timestamp={msg.created_at}
                  isOwn={msg.sender_user_id === $auth.userId}
                  messageId={msg.id || ""}
                  reactions={messageReactions[msg.id] || []}
                  on:react={handleReaction}
                  on:thread={handleOpenThread}
                  on:pin={handlePin}
                  on:delete={handleDelete}
                />
              {/each}
            </div>
          {/if}
        </div>

        <!-- Message composer -->
        <div class="p-6 pt-0">
          <MessageInput
            placeholder="Say something friendly in #{activeChannel.name}..."
            {members}
            on:send={handleSend}
          />
        </div>
      {/if}
    </div>

    <!-- Tasks/Productivity Panel (Right) -->
    {#if showTasks}
      <TaskPanel on:close={() => (showTasks = false)} />
    {/if}

    <!-- Other panels (Thread + Members) - still as overlays for now, or could be integrated -->
    <ThreadPanel
      visible={showThread}
      parentMessage={threadParent}
      replies={threadReplies}
      {getDecrypted}
      on:reply={handleThreadReply}
      on:close={handleCloseThread}
    />

    <MembersPanel
      bind:this={membersPanel}
      visible={showMembers}
      {members}
      currentUserId={$auth.userId}
      {isAdmin}
      on:addMember={handleAddMember}
      on:removeMember={handleRemoveMember}
      on:close={() => (showMembers = false)}
    />
  </div>

  <!-- Modals -->
  <CreateWorkspaceModal bind:this={wsModal} visible={showCreateWs}
    on:create={handleCreateWorkspace} on:close={() => (showCreateWs = false)} />

  <CreateChannelModal bind:this={chModal} visible={showCreateCh}
    on:create={handleCreateChannel} on:close={() => (showCreateCh = false)} />

  <InviteModal bind:this={inviteModal} visible={showInvite}
    inviteCode={activeWorkspace?.invite_code || ""}
    workspaceName={activeWorkspace?.name || ""}
    on:join={handleJoinByInvite} on:close={() => (showInvite = false)} />
{/if}
