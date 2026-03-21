const AUTH_BASE = import.meta.env.VITE_AUTH_URL || "http://localhost:8081";
const MESSAGING_BASE = import.meta.env.VITE_MESSAGING_URL || "http://localhost:8082";
const KEYDIST_BASE = import.meta.env.VITE_KEYDIST_URL || "http://localhost:8083";
const WORKSPACE_BASE = import.meta.env.VITE_WORKSPACE_URL || "http://localhost:8086";

function authHeaders(token) {
  return {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

export async function register(username, email, password) {
  const res = await fetch(`${AUTH_BASE}/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, email, password }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Registration failed (${res.status})`);
  }
  return res.json();
}

export async function login(username, password) {
  const res = await fetch(`${AUTH_BASE}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Login failed (${res.status})`);
  }
  return res.json();
}

export async function refreshToken(token) {
  const res = await fetch(`${AUTH_BASE}/refresh`, {
    method: "POST",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Token refresh failed");
  return res.json();
}

export async function uploadIdentityKey(token, publicKeyB64) {
  const res = await fetch(`${KEYDIST_BASE}/v1/keys/identity`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ public_key_b64: publicKeyB64, key_type: "identity" }),
  });
  if (!res.ok) throw new Error("Key upload failed");
  return res.json();
}

export async function getIdentityKey(token, userId) {
  const res = await fetch(`${KEYDIST_BASE}/v1/keys/identity/${userId}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Key fetch failed");
  return res.json();
}

export async function sendMessage(token, recipientUserId, ciphertextB64, nonceB64, channelId) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({
      recipient_user_id: recipientUserId,
      ciphertext_b64: ciphertextB64,
      nonce_b64: nonceB64,
      channel_id: channelId || "",
      content_type: "text",
    }),
  });
  if (!res.ok) throw new Error("Send message failed");
  return res.json();
}

export async function getInbox(token, limit = 50) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages/inbox?limit=${limit}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Inbox fetch failed");
  return res.json();
}

export function connectInboxWs(token, onMessage) {
  const wsBase = MESSAGING_BASE.replace(/^http/, "ws");
  const ws = new WebSocket(`${wsBase}/v1/ws?access_token=${token}`);
  ws.onmessage = (event) => {
    try {
      onMessage(JSON.parse(event.data));
    } catch { /* ignore parse errors */ }
  };
  return ws;
}

// --- Workspace APIs ---

export async function createWorkspace(token, name, description = "") {
  const res = await fetch(`${WORKSPACE_BASE}/v1/workspaces`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ name, description }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || "Failed to create workspace");
  }
  return res.json();
}

export async function listWorkspaces(token) {
  const res = await fetch(`${WORKSPACE_BASE}/v1/workspaces`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to list workspaces");
  return res.json();
}

export async function getWorkspace(token, id) {
  const res = await fetch(`${WORKSPACE_BASE}/v1/workspaces/${id}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to get workspace");
  return res.json();
}

export async function joinWorkspaceByInvite(token, inviteCode) {
  const res = await fetch(`${WORKSPACE_BASE}/v1/workspaces/join`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ invite_code: inviteCode }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || "Failed to join workspace");
  }
  return res.json();
}

export async function listWorkspaceMembers(token, workspaceId) {
  const res = await fetch(`${WORKSPACE_BASE}/v1/workspaces/${workspaceId}/members`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to list members");
  return res.json();
}

export async function addWorkspaceMember(token, workspaceId, userId, role = "member") {
  const res = await fetch(`${WORKSPACE_BASE}/v1/workspaces/${workspaceId}/members`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ user_id: userId, role }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || "Failed to add member");
  }
  return res.json();
}

export async function removeWorkspaceMember(token, workspaceId, userId) {
  const res = await fetch(`${WORKSPACE_BASE}/v1/workspaces/${workspaceId}/members/${userId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to remove member");
  return res.json();
}

// --- Channel APIs ---

export async function createChannel(token, workspaceId, name, description = "", topic = "", isPrivate = false) {
  const res = await fetch(`${WORKSPACE_BASE}/v1/workspaces/${workspaceId}/channels`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ name, description, topic, is_private: isPrivate }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || "Failed to create channel");
  }
  return res.json();
}

export async function listChannels(token, workspaceId) {
  const res = await fetch(`${WORKSPACE_BASE}/v1/workspaces/${workspaceId}/channels`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to list channels");
  return res.json();
}

export async function updateChannelTopic(token, channelId, topic) {
  const res = await fetch(`${WORKSPACE_BASE}/v1/channels/${channelId}/topic`, {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify({ topic }),
  });
  if (!res.ok) throw new Error("Failed to update topic");
  return res.json();
}

export async function archiveChannel(token, channelId) {
  const res = await fetch(`${WORKSPACE_BASE}/v1/channels/${channelId}/archive`, {
    method: "POST",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to archive channel");
  return res.json();
}

// --- Rich Messaging APIs ---

export async function postThreadReply(token, parentId, recipientUserId, ciphertextB64, nonceB64) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages/${parentId}/replies`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ recipient_user_id: recipientUserId, ciphertext_b64: ciphertextB64, nonce_b64: nonceB64 }),
  });
  if (!res.ok) throw new Error("Failed to post reply");
  return res.json();
}

export async function getThreadReplies(token, parentId, limit = 100) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages/${parentId}/replies?limit=${limit}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to get replies");
  return res.json();
}

export async function addReaction(token, messageId, emoji) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages/${messageId}/reactions`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ emoji }),
  });
  if (!res.ok) throw new Error("Failed to add reaction");
  return res.json();
}

export async function removeReaction(token, messageId, emoji) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages/${messageId}/reactions/${emoji}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to remove reaction");
  return res.json();
}

export async function getReactions(token, messageId) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages/${messageId}/reactions`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to get reactions");
  return res.json();
}

export async function pinMessage(token, messageId, channelId) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages/${messageId}/pin`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ channel_id: channelId }),
  });
  if (!res.ok) throw new Error("Failed to pin message");
  return res.json();
}

export async function unpinMessage(token, messageId) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages/${messageId}/pin`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to unpin message");
  return res.json();
}

export async function listPins(token, channelId) {
  const res = await fetch(`${MESSAGING_BASE}/v1/channels/${channelId}/pins`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to list pins");
  return res.json();
}

export async function editMessage(token, messageId, ciphertextB64, nonceB64) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages/${messageId}`, {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify({ ciphertext_b64: ciphertextB64, nonce_b64: nonceB64 }),
  });
  if (!res.ok) throw new Error("Failed to edit message");
  return res.json();
}

export async function deleteMessageApi(token, messageId) {
  const res = await fetch(`${MESSAGING_BASE}/v1/messages/${messageId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error("Failed to delete message");
  return res.json();
}
