# SecureCollab Product Scope

Last updated: March 21, 2026

## Vision
A self-hosted, zero-knowledge team collaboration platform combining **Slack-like messaging** with **ClickUp-like project management**. All content is end-to-end encrypted.

## UI Direction
- **Inspiration**: Slack (messaging) + ClickUp (project management)
- **Stack**: Svelte 4 + Tailwind CSS (with Tauri for desktop)
- **Approach**: Build brick by brick — each component independently functional

---

## Feature Map

### 1. Authentication & Identity
- [x] User registration and login (backend)
- [x] JWT access/refresh token flow
- [x] Client auth UI (sign in / register modal)
- [ ] User profile (avatar, display name, status)
- [ ] Password reset flow
- [ ] OAuth/SSO integration (optional, Phase 5)

### 2. Workspaces
- [x] DB schema exists (`workspaces`, `workspace_members` tables)
- [ ] Create workspace API + UI
- [ ] Join workspace (invite link / code)
- [ ] Workspace settings (name, icon, description)
- [ ] Workspace member list with roles
- [ ] Role enforcement: `owner` > `admin` > `member` > `viewer`
- [ ] Workspace switcher in sidebar

### 3. Channels
- [x] DB schema exists (`channels` table)
- [ ] Create channel API + UI (public / private)
- [ ] Channel list in sidebar (grouped by workspace)
- [ ] Channel settings (name, topic, description)
- [ ] Channel member management (add/remove)
- [ ] Channel archive/delete
- [ ] Unread indicators and badges

### 4. Messaging
- [x] E2E encrypted message send/receive (backend)
- [x] WebSocket real-time delivery
- [x] Message UI with send/receive
- [ ] Message threads (reply to message)
- [ ] Message reactions (emoji)
- [ ] Pin messages in channel
- [ ] @mention users and @channel
- [ ] Edit and delete own messages
- [ ] Message search (metadata-only, zero-knowledge)
- [ ] Link previews (client-side unfurl)
- [ ] Code block / markdown rendering

### 5. File Attachments
- [ ] File upload with client-side encryption
- [ ] File download with decryption
- [ ] Image/file preview in chat
- [ ] File size limits and type validation
- [ ] File storage service (S3-compatible backend)

### 6. Project Management (Kanban)
- [ ] Project board CRUD (per workspace)
- [ ] Kanban columns (To Do, In Progress, Done, custom)
- [ ] Task cards (title, description, assignee, due date, priority)
- [ ] Drag-and-drop card movement
- [ ] Task labels / tags
- [ ] Task comments (linked to messaging)
- [ ] Board views: Kanban, List, Timeline (future)
- [ ] Task status tracking and filters

### 7. Notifications
- [ ] In-app notification center
- [ ] Notification for: mentions, DMs, task assignments, channel invites
- [ ] Read/unread state
- [ ] Desktop notifications (Tauri integration)
- [ ] Email notifications (optional)

### 8. User Presence & Status
- [ ] Online/offline/away indicators
- [ ] Custom status messages
- [ ] Typing indicators in channels

### 9. Admin & Settings
- [ ] Workspace admin panel
- [ ] User management (invite, remove, change role)
- [ ] Audit log (who did what, when)
- [ ] Data export
- [ ] Workspace-level settings (encryption policy, retention)

---

## UI Component Build Order (Brick by Brick)

Each component is independently buildable and testable:

### Brick 1: Layout Shell
- App layout: sidebar + main content area
- Sidebar: workspace switcher, channel list, DM list
- Top bar: channel name, search, user menu

### Brick 2: Workspace Management
- Create workspace form
- Workspace settings page
- Member invite / management

### Brick 3: Channel System
- Create channel modal
- Channel list with unread counts
- Channel header (topic, members, settings)

### Brick 4: Chat Experience
- Message list (virtualized for performance)
- Message input with markdown support
- Message actions (reply, react, pin, edit, delete)
- @mention autocomplete

### Brick 5: Thread View
- Thread sidebar / overlay
- Reply chain display
- Thread notification badges

### Brick 6: File Attachments
- Drag-and-drop upload
- File preview (images, PDFs)
- Download with decryption

### Brick 7: Kanban Board
- Board layout with columns
- Card CRUD
- Drag-and-drop
- Card detail modal (description, comments, assignee, due date)

### Brick 8: Notifications
- Notification bell with badge
- Notification dropdown/panel
- Desktop notifications

### Brick 9: User Presence
- Status indicators
- Typing indicators
- Custom status

### Brick 10: Search
- Global search bar
- Search results (messages, files, tasks)
- Filters (by channel, user, date)

---

## Backend Services Needed

| Service | Status | Purpose |
|---------|--------|---------|
| `gateway` | Done | API entry, rate limiting, auth middleware |
| `auth` | Done | Registration, login, JWT |
| `keydist` | Done | Public key distribution |
| `messaging` | Done | E2E encrypted messages, WebSocket |
| `analytics` | Done | Message volume metrics |
| `workspace` | **Needed** | Workspace + channel + member CRUD |
| `filestore` | **Needed** | Encrypted file upload/download |
| `projects` | **Needed** | Kanban boards, tasks, columns |
| `notifications` | **Needed** | Event-driven notification delivery |
| `presence` | **Needed** | Online status, typing indicators |
| `search` | **Needed** | Metadata search index |

---

## Tech Decisions
- **UI**: Svelte 4 + Tailwind CSS (custom components)
- **Desktop**: Tauri (Rust + Svelte)
- **Backend**: Go microservices
- **Database**: PostgreSQL (primary), ClickHouse (analytics)
- **Real-time**: WebSocket
- **Encryption**: X25519 key exchange + ChaCha20-Poly1305
- **File storage**: S3-compatible (MinIO for local dev)
