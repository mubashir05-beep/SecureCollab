<script>
  import Sidebar from "./lib/ui/Sidebar.svelte";
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
  let workspacesLoaded = false;

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
    if (wsConn) { wsConn.close(); wsConn = null; }
    workspaces = [];
    channels = [];
    messages = [];
    activeWorkspace = null;
    activeChannel = null;
    workspacesLoaded = false;
  }

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
      workspaces = (res.workspaces || []).map(ws => ({ ...ws, unreadCount: 0 }));
      if (workspaces.length > 0 && !activeWorkspace) {
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
      channels = (res.channels || []).map(ch => ({ ...ch, unreadCount: 0 }));
      if (channels.length > 0 && (!activeChannel || activeChannel.workspace_id !== workspaceId)) {
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
    } catch { /* silent */ }
  }

  function connectWs() {
    if (wsConn) { wsConn.close(); wsConn = null; }
    wsConn = api.connectInboxWs($auth.token, (envelope) => {
      messages = [...messages, envelope];
    });
    wsConn.onclose = () => { wsConn = null; };
  }

  async function handleSend(e) {
    const text = e.detail;
    try {
      const keys = $keyStore;
      const payload = await encryptMessage(keys.privateKey, keys.publicKey, text);
      await api.sendMessage($auth.token, $auth.userId, payload.ciphertext_b64, payload.nonce_b64, activeChannel?.id || "");
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
      messages = messages; // trigger reactivity
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
    } catch { /* silent */ }
  }

  // --- Delete ---
  async function handleDelete(e) {
    const { messageId } = e.detail;
    try {
      await api.deleteMessageApi($auth.token, messageId);
      messages = messages.filter(m => m.id !== messageId);
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
  }
</script>

<!-- ════════════════════════════════════════════
     LANDING — not authenticated
═══════════════════════════════════════════════ -->
{#if !$isAuthenticated}
  <main class="flex min-h-screen flex-col items-center justify-center bg-shell-sidebar px-4">
    <div class="w-full max-w-sm text-center animate-slide-up">

      <!-- Logo mark -->
      <div class="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-2xl bg-shell-accent shadow-panel" aria-hidden="true">
        <span class="text-2xl font-bold text-white">SC</span>
      </div>

      <h1 class="mb-2 text-3xl font-bold text-shell-ink">SecureCollab</h1>
      <p class="mb-2 text-base text-shell-muted">Zero-knowledge team messaging</p>

      <!-- Feature pills -->
      <div class="mb-8 flex flex-wrap justify-center gap-2">
        <span class="flex items-center gap-1.5 rounded-full border border-shell-border bg-shell-elevated px-3 py-1 text-xs text-shell-muted">
          <svg class="h-3 w-3 text-shell-success" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
            <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
          </svg>
          E2E Encrypted
        </span>
        <span class="flex items-center gap-1.5 rounded-full border border-shell-border bg-shell-elevated px-3 py-1 text-xs text-shell-muted">
          <svg class="h-3 w-3 text-shell-accentText" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
            <path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v1h8v-1zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-1a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v1h-3zM4.75 12.094A5.973 5.973 0 004 15v1H1v-1a3 3 0 013.75-2.906z" />
          </svg>
          Team Workspaces
        </span>
        <span class="flex items-center gap-1.5 rounded-full border border-shell-border bg-shell-elevated px-3 py-1 text-xs text-shell-muted">
          <svg class="h-3 w-3 text-shell-warn" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
            <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
          </svg>
          Self-Hosted
        </span>
      </div>

      <button
        on:click={() => (showAuth = true)}
        class="w-full rounded-xl bg-shell-accent py-3 text-base font-semibold text-white transition-colors hover:bg-shell-accentHov focus-visible:ring-2 focus-visible:ring-shell-accent focus-visible:ring-offset-2 focus-visible:ring-offset-shell-sidebar"
      >
        Get Started
      </button>

      <p class="mt-4 text-xs text-shell-subtle">
        Open source &middot; Zero knowledge &middot; No tracking
      </p>
    </div>
  </main>

  <AuthModal bind:this={authModal} visible={showAuth}
    on:auth={handleAuth} on:close={() => (showAuth = false)} />

<!-- ════════════════════════════════════════════
     ONBOARDING — authenticated, no workspaces
═══════════════════════════════════════════════ -->
{:else if workspacesLoaded && workspaces.length === 0}
  <main class="flex min-h-screen flex-col items-center justify-center bg-shell-sidebar px-4">
    <div class="w-full max-w-md animate-slide-up">

      <!-- Welcome card -->
      <div class="rounded-2xl border border-shell-border bg-shell-elevated p-8 shadow-modal text-center mb-4">
        <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-shell-accent" aria-hidden="true">
          <span class="text-xl font-bold text-white">SC</span>
        </div>
        <h1 class="mb-1 text-2xl font-bold text-shell-ink">
          Welcome, {$auth.username}!
        </h1>
        <p class="mb-6 text-sm text-shell-muted leading-relaxed">
          Create a workspace for your team or join one with an invite code to get started.
        </p>

        <div class="flex flex-col gap-3">
          <!-- Create workspace CTA -->
          <button
            on:click={() => (showCreateWs = true)}
            class="w-full rounded-xl bg-shell-accent py-3 text-sm font-semibold text-white transition-colors hover:bg-shell-accentHov"
          >
            Create a Workspace
          </button>

          <div class="flex items-center gap-3">
            <hr class="flex-1 border-shell-borderSub" />
            <span class="text-xs text-shell-subtle">or join with an invite code</span>
            <hr class="flex-1 border-shell-borderSub" />
          </div>

          <!-- Join workspace -->
          <div class="flex gap-2">
            <input
              type="text"
              bind:value={onboardingInviteCode}
              placeholder="Paste invite code…"
              aria-label="Invite code to join a workspace"
              class="flex-1 rounded-xl border border-shell-border bg-shell-bg px-4 py-2.5 text-sm text-shell-ink placeholder-shell-subtle outline-none transition-colors focus:border-shell-accent focus:ring-1 focus:ring-shell-accent/30"
              on:keydown={(e) => e.key === "Enter" && handleOnboardingJoin()}
            />
            <button
              on:click={handleOnboardingJoin}
              disabled={!onboardingInviteCode.trim()}
              class="rounded-xl bg-shell-surface px-5 py-2.5 text-sm font-medium text-shell-ink transition-colors hover:bg-shell-elevated disabled:opacity-40 disabled:cursor-not-allowed"
            >
              Join
            </button>
          </div>

          {#if onboardingError}
            <p class="text-sm text-shell-danger" role="alert">{onboardingError}</p>
          {/if}
        </div>
      </div>

      <div class="text-center">
        <button
          on:click={handleLogout}
          class="text-sm text-shell-subtle transition-colors hover:text-shell-muted"
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
  <div class="flex h-screen overflow-hidden bg-shell-bg" role="application" aria-label="SecureCollab">

    <!-- Sidebar -->
    <Sidebar
      {workspaces} {channels} {directMessages} {activeWorkspace} {activeChannel} {currentUser}
      on:selectWorkspace={handleSelectWorkspace}
      on:selectChannel={handleSelectChannel}
      on:createWorkspace={() => (showCreateWs = true)}
      on:createChannel={() => (showCreateCh = true)}
      on:logout={handleLogout}
      on:invite={handleOpenInvite}
    />

    <!-- Main content area -->
    <div class="flex flex-1 flex-col overflow-hidden">
      <TopBar
        channelName={activeChannel?.name || ""}
        channelTopic={activeChannel?.topic || ""}
        memberCount={members.length}
        on:showMembers={() => (showMembers = !showMembers)}
      />

      <!-- Channel content -->
      {#if !activeChannel}
        <!-- Empty state: no channels -->
        <div class="flex flex-1 flex-col items-center justify-center gap-4 px-8 text-center">
          <div class="grid h-16 w-16 place-content-center rounded-2xl bg-shell-surface text-2xl text-shell-subtle" aria-hidden="true">#</div>
          <div>
            <p class="mb-1 text-lg font-semibold text-shell-ink">No channels yet</p>
            <p class="text-sm text-shell-muted">Create a channel to start collaborating with your team.</p>
          </div>
          <button
            on:click={() => (showCreateCh = true)}
            class="rounded-xl bg-shell-accent px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-shell-accentHov"
          >
            Create Channel
          </button>
        </div>

      {:else}
        <!-- Message list -->
        <div class="flex-1 overflow-y-auto" role="log" aria-label="Message history" aria-live="polite">
          {#if messages.length === 0}
            <!-- Empty channel state -->
            <div class="flex h-full flex-col items-center justify-center gap-3 px-8 text-center">
              <div class="grid h-14 w-14 place-content-center rounded-2xl bg-shell-surface text-2xl text-shell-subtle" aria-hidden="true">#</div>
              <div>
                <p class="text-lg font-bold text-shell-ink">Welcome to #{activeChannel.name}</p>
                <p class="mt-1 text-sm text-shell-muted leading-relaxed">
                  This is the start of <span class="font-medium text-shell-ink">#{activeChannel.name}</span>.
                  All messages are end-to-end encrypted.
                </p>
              </div>
              {#if activeChannel.topic}
                <div class="max-w-sm rounded-lg border border-shell-borderSub bg-shell-surface px-4 py-2.5 text-sm text-shell-muted">
                  <span class="font-medium text-shell-ink">Topic:</span> {activeChannel.topic}
                </div>
              {/if}
            </div>
          {:else}
            <div class="py-2">
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
        <MessageInput
          placeholder="Message #{activeChannel.name}"
          {members}
          on:send={handleSend}
        />
      {/if}
    </div>

    <!-- Right panels (Thread + Members) -->
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

  <!-- Modals (rendered outside layout flow) -->
  <CreateWorkspaceModal bind:this={wsModal} visible={showCreateWs}
    on:create={handleCreateWorkspace} on:close={() => (showCreateWs = false)} />

  <CreateChannelModal bind:this={chModal} visible={showCreateCh}
    on:create={handleCreateChannel} on:close={() => (showCreateCh = false)} />

  <InviteModal bind:this={inviteModal} visible={showInvite}
    inviteCode={activeWorkspace?.invite_code || ""}
    workspaceName={activeWorkspace?.name || ""}
    on:join={handleJoinByInvite} on:close={() => (showInvite = false)} />
{/if}
