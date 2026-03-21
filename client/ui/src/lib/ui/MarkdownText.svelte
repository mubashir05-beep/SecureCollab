<script>
  export let text = "";

  // Lightweight markdown → HTML (no external dep)
  function renderMarkdown(src) {
    if (!src) return "";
    let html = escapeHtml(src);

    // Code blocks: ```lang\n...\n```
    html = html.replace(/```(\w*)\n([\s\S]*?)```/g, (_m, lang, code) =>
      `<pre class="md-codeblock"><code${lang ? ` class="lang-${lang}"` : ""}>${code.trim()}</code></pre>`
    );

    // Inline code: `code`
    html = html.replace(/`([^`\n]+)`/g, '<code class="md-inline-code">$1</code>');

    // Bold: **text** or __text__
    html = html.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");
    html = html.replace(/__(.+?)__/g, "<strong>$1</strong>");

    // Italic: *text* or _text_
    html = html.replace(/(?<!\*)\*(?!\*)(.+?)(?<!\*)\*(?!\*)/g, "<em>$1</em>");
    html = html.replace(/(?<!_)_(?!_)(.+?)(?<!_)_(?!_)/g, "<em>$1</em>");

    // Strikethrough: ~~text~~
    html = html.replace(/~~(.+?)~~/g, "<del>$1</del>");

    // Links: [text](url)
    html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer" class="md-link">$1</a>');

    // Auto-link bare URLs (not already inside an href)
    html = html.replace(/(?<!href="|">)(https?:\/\/[^\s<]+)/g, '<a href="$1" target="_blank" rel="noopener noreferrer" class="md-link">$1</a>');

    // @mentions highlight
    html = html.replace(/(^|\s)(@\w+)/g, '$1<span class="md-mention">$2</span>');

    // Blockquotes: > text
    html = html.replace(/^&gt; (.+)$/gm, '<blockquote class="md-blockquote">$1</blockquote>');

    // Line breaks
    html = html.replace(/\n/g, "<br>");

    return html;
  }

  function escapeHtml(str) {
    return str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  $: rendered = renderMarkdown(text);
</script>

<span class="md-text">{@html rendered}</span>

<style>
  .md-text :global(.md-codeblock) {
    background: #f1f5f9;
    border-radius: 0.5rem;
    padding: 0.75rem 1rem;
    margin: 0.25rem 0;
    overflow-x: auto;
    font-size: 0.8125rem;
    font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
    white-space: pre;
    display: block;
  }
  .md-text :global(.md-inline-code) {
    background: #f1f5f9;
    border-radius: 0.25rem;
    padding: 0.125rem 0.375rem;
    font-size: 0.8125rem;
    font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
    color: #be185d;
  }
  .md-text :global(.md-link) {
    color: #2563eb;
    text-decoration: underline;
    text-underline-offset: 2px;
  }
  .md-text :global(.md-link:hover) {
    color: #1d4ed8;
  }
  .md-text :global(.md-mention) {
    background: #dbeafe;
    color: #1e40af;
    border-radius: 0.25rem;
    padding: 0.0625rem 0.25rem;
    font-weight: 500;
  }
  .md-text :global(.md-blockquote) {
    border-left: 3px solid #cbd5e1;
    padding-left: 0.75rem;
    color: #64748b;
    margin: 0.25rem 0;
    display: block;
  }
  .md-text :global(strong) {
    font-weight: 700;
  }
  .md-text :global(em) {
    font-style: italic;
  }
  .md-text :global(del) {
    text-decoration: line-through;
    color: #94a3b8;
  }
</style>
