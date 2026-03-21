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
  import { auth, isAuthenticated } from "./lib/authStore.js";
  import * as api from "./lib/api.js";
  import { encryptMessage, decryptMessage } from "./lib/crypto.js";

  // --- State ---
  let showAuth = false;
  let showCreateWs = false;
  let showCreateCh = false;
  let showInvite = false;
  let onboardingInviteCode = "";
  let onboardingError = "";
  let authModal, wsModal, chModal, inviteModal;

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

  $: if ($isAuthenticated) {
    loadWorkspaces();
  }

  $: if (!$isAuthenticated) {
    cleanup();
  }

  // Load channels when workspace changes
  $: if (activeWorkspace && $isAuthenticated) {
    loadChannels(activeWorkspace.id);
    loadMembers(activeWorkspace.id);
  }

  // Load messages + WS when channel changes
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
      workspaces = (res.workspaces || []).map(ws => ({
        ...ws, unreadCount: 0,
      }));
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
      channels = (res.channels || []).map(ch => ({
        ...ch, unreadCount: 0,
      }));
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
      const payload = await encryptMessage("", "", text);
      await api.sendMessage($auth.token, $auth.userId, payload.ciphertext_b64, payload.nonce_b64, activeChannel?.id || "");
    } catch (err) {
      console.error("Send failed:", err);
    }
  }

  // --- Decrypt ---
  const decryptCache = {};
  function getDecrypted(msg) {
    const key = msg.id || `${msg.ciphertext_b64}:${msg.nonce_b64}`;
    if (decryptCache[key] !== undefined) return decryptCache[key];
    decryptCache[key] = "...";
    decryptMessage("", "", msg.ciphertext_b64, msg.nonce_b64).then((text) => {
      decryptCache[key] = text;
      messages = messages;
    });
    return decryptCache[key];
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
      const payload = await encryptMessage("", "", text);
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

{#if !$isAuthenticated}
  <div class="grid min-h-screen place-content-center bg-slate-900 text-white">
    <div class="text-center">
      <div class="mx-auto mb-6 grid h-20 w-20 place-content-center rounded-2xl bg-gradient-to-br from-shell-accent to-shell-success text-3xl font-bold">SC</div>
      <h1 class="mb-2 text-3xl font-bold">SecureCollab</h1>
      <p class="mb-8 text-slate-400">Zero-knowledge team messaging & project management</p>
      <button on:click={() => (showAuth = true)}
        class="rounded-xl bg-shell-accent px-8 py-3 font-medium text-white transition hover:opacity-90">
        Get Started
      </button>
    </div>
  </div>

  <AuthModal bind:this={authModal} visible={showAuth}
    on:auth={handleAuth} on:close={() => (showAuth = false)} />

{:else if workspacesLoaded && workspaces.length === 0}
  <!-- Onboarding: no workspaces -->
  <div class="grid min-h-screen place-content-center bg-slate-50">
    <div class="w-full max-w-md text-center">
      <div class="mx-auto mb-6 grid h-20 w-20 place-content-center rounded-2xl bg-gradient-to-br from-shell-accent to-shell-success text-3xl font-bold text-white">SC</div>
      <h1 class="mb-2 text-2xl font-bold text-slate-900">Welcome, {$auth.username}!</h1>
      <p class="mb-8 text-slate-500">Create a workspace for your team or join one with an invite code.</p>

      <div class="flex flex-col gap-3">
        <button on:click={() => (showCreateWs = true)}
          class="w-full rounded-xl bg-shell-accent px-6 py-3 font-medium text-white transition hover:opacity-90">
          Create a Workspace
        </button>

        <div class="flex items-center gap-3 text-sm text-slate-400">
          <hr class="flex-1 border-slate-200" />
          <span>or</span>
          <hr class="flex-1 border-slate-200" />
        </div>

        <div class="flex gap-2">
          <input
            type="text"
            bind:value={onboardingInviteCode}
            placeholder="Paste invite code..."
            class="flex-1 rounded-xl border border-slate-300 px-4 py-3 text-sm text-slate-700 placeholder-slate-400 focus:border-shell-accent focus:outline-none focus:ring-1 focus:ring-shell-accent/30"
            on:keydown={(e) => e.key === "Enter" && handleOnboardingJoin()}
          />
          <button on:click={handleOnboardingJoin}
            disabled={!onboardingInviteCode.trim()}
            class="rounded-xl bg-slate-800 px-5 py-3 text-sm font-medium text-white transition hover:bg-slate-700 disabled:opacity-40">
            Join
          </button>
        </div>

        {#if onboardingError}
          <p class="text-sm text-red-500">{onboardingError}</p>
        {/if}
      </div>

      <button on:click={handleLogout}
        class="mt-6 text-sm text-slate-400 hover:text-slate-600 transition">
        Logout
      </button>
    </div>
  </div>

  <CreateWorkspaceModal bind:this={wsModal} visible={showCreateWs}
    on:create={handleCreateWorkspace} on:close={() => (showCreateWs = false)} />

{:else}
  <div class="flex h-screen overflow-hidden bg-white">
    <Sidebar
      {workspaces} {channels} {directMessages} {activeWorkspace} {activeChannel} {currentUser}
      on:selectWorkspace={handleSelectWorkspace}
      on:selectChannel={handleSelectChannel}
      on:createWorkspace={() => (showCreateWs = true)}
      on:createChannel={() => (showCreateCh = true)}
      on:logout={handleLogout}
      on:invite={handleOpenInvite}
    />

    <div class="flex flex-1 flex-col overflow-hidden">
      <TopBar
        channelName={activeChannel?.name || ""}
        channelTopic={activeChannel?.topic || ""}
        memberCount={members.length}
      />

      {#if !activeChannel}
        <!-- No channels -->
        <div class="flex flex-1 flex-col items-center justify-center text-slate-400">
          <div class="mb-4 grid h-16 w-16 place-content-center rounded-2xl bg-slate-100 text-2xl">#</div>
          <p class="mb-2 text-lg font-semibold text-slate-700">No channels yet</p>
          <p class="mb-4 text-sm">Create a channel to start chatting.</p>
          <button on:click={() => (showCreateCh = true)}
            class="rounded-xl bg-shell-accent px-6 py-2.5 text-sm font-medium text-white hover:opacity-90">
            Create Channel
          </button>
        </div>

      {:else}
        <!-- Message area -->
        <div class="flex-1 overflow-y-auto">
          {#if messages.length === 0}
            <div class="flex h-full flex-col items-center justify-center text-slate-400">
              <div class="mb-3 grid h-16 w-16 place-content-center rounded-2xl bg-slate-100 text-2xl">#</div>
              <p class="text-lg font-semibold text-slate-700">Welcome to #{activeChannel.name}</p>
              <p class="text-sm">This is the start of the channel. Messages are end-to-end encrypted.</p>
            </div>
          {:else}
            <div class="py-2">
              {#each messages as msg}
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

        <MessageInput
          placeholder="Message #{activeChannel.name}"
          {members}
          on:send={handleSend}
        />
      {/if}
    </div>

    <ThreadPanel
      visible={showThread}
      parentMessage={threadParent}
      replies={threadReplies}
      {getDecrypted}
      on:reply={handleThreadReply}
      on:close={handleCloseThread}
    />
  </div>

  <CreateWorkspaceModal bind:this={wsModal} visible={showCreateWs}
    on:create={handleCreateWorkspace} on:close={() => (showCreateWs = false)} />

  <CreateChannelModal bind:this={chModal} visible={showCreateCh}
    on:create={handleCreateChannel} on:close={() => (showCreateCh = false)} />

  <InviteModal bind:this={inviteModal} visible={showInvite}
    inviteCode={activeWorkspace?.invite_code || ""}
    workspaceName={activeWorkspace?.name || ""}
    on:join={handleJoinByInvite} on:close={() => (showInvite = false)} />
{/if}
