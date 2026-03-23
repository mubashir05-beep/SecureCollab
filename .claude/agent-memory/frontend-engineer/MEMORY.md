# Frontend Engineer — SecureCollab Agent Memory

## Design System (confirmed in tailwind.config.cjs)
- **Accent**: `shell-accent` = `#5865f2` (indigo), `shell-accentHov` = `#4752c4`
- **Sidebar bg**: `shell-sidebar` = `#1a1d21`
- **Channel panel bg**: `shell-panel` = `#19171d`
- **Main content bg**: `shell-bg` = `#1d2026`
- **Surface (hover/elevated)**: `shell-surface` = `#222529`, `shell-elevated` = `#2b2d31`
- **Borders**: `shell-border` = `#3f4147`, `shell-borderSub` = `#2b2d31`
- **Text**: `shell-ink` = `#e3e5e8`, `shell-muted` = `#949ba4`, `shell-subtle` = `#6d7379`
- **Semantic**: `shell-success` = `#23a55a`, `shell-warn` = `#f0b232`, `shell-danger` = `#f23f42`
- **Mention**: `shell-mention` = `#444271`, `shell-mentionTxt` = `#c9ccff`

## File Locations
- `client/ui/src/App.svelte` — root: landing, onboarding, main shell
- `client/ui/src/lib/ui/` — all UI components (Sidebar, TopBar, MessageBubble, MessageInput, etc.)
- `client/ui/src/lib/authStore.js` — Svelte writable store, persists to localStorage
- `client/ui/src/lib/keyStore.js` — crypto key bootstrap + storage
- `client/ui/src/lib/api.js` — all API calls (do NOT change endpoints/formats)
- `client/ui/src/lib/crypto.js` — X25519+ChaCha20-Poly1305 encrypt/decrypt
- `client/ui/tailwind.config.cjs` — design tokens (shell-* colors, animations, fonts)
- `client/ui/src/app.css` — global base styles, scrollbar, .md-* markdown classes

## Component API Reference
- **Button**: props `variant` (primary|secondary|ghost|danger), `size` (sm|md|lg), `loading`, `disabled`, `fullWidth`, `type`
- **Sidebar**: events `selectWorkspace`, `selectChannel`, `createWorkspace`, `createChannel`, `logout`, `invite`
- **TopBar**: props `channelName`, `channelTopic`, `memberCount`, `isEncrypted`; events `showMembers`, `search`
- **MessageBubble**: props `sender`, `content`, `timestamp`, `isOwn`, `isPinned`, `isEdited`, `reactions`, `messageId`; events `react`, `thread`, `pin`, `delete`
- **MessageInput**: props `placeholder`, `disabled`, `members`; events `send`, `attach`
- **MembersPanel**: props `visible`, `members`, `currentUserId`, `isAdmin`; events `addMember`, `removeMember`, `close`; export `setError()`
- **ThreadPanel**: props `visible`, `parentMessage`, `replies`, `getDecrypted`; events `reply`, `close`
- **AuthModal**: events `auth` (detail: {mode, username, email, password}), `close`; exports `setError()`, `setLoading()`
- **CreateWorkspaceModal / CreateChannelModal**: events `create`, `close`; exports `setError()`, `reset()`
- **InviteModal**: props `inviteCode`, `workspaceName`; events `join`, `close`; exports `setError()`, `reset()`
- **EmojiPicker**: prop `visible` (bind:); event `pick` (detail: emoji label string)

## Key Patterns
- Dark theme throughout: no `bg-white`, no `text-gray-*` — use `shell-*` tokens exclusively
- Avatar colors: deterministic from `name.charCodeAt(0) % palette.length` (8-color palette)
- Reactive `$:` statements with `showMentions` and `mentionCandidates` — avoid circular deps (don't read showMentions inside mentionCandidates derivation)
- Modals use `backdrop-blur-sm` + `animate-fade-in` + `animate-slide-up` for entry
- `aria-label`, `role`, `aria-current`, `aria-modal` on all interactive/structural elements
- Markdown rendered in `MarkdownText.svelte` via custom regex — `.md-*` global classes live in `app.css`
- All API endpoints preserved in `api.js` — never modify URLs or request shapes

## Gotchas
- Svelte 4 cyclical reactive dep: `$: derived = () => { if (!query) return [] }` followed by `$: if (derived.length === 0) flag = false` creates a cycle if `flag` is also read in `derived`. Fix: remove `flag` from derived statement.
- `shell-panel` color `#19171d` is intentionally slightly purple-tinted (Slack-like).
- Tauri clipboard: use `navigator.clipboard` with execCommand fallback (see InviteModal).
- No SvelteKit, no browser router — single-page conditional rendering in App.svelte.
