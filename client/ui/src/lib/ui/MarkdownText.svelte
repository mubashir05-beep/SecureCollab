<script>
  export let text = "";

  function escapeHtml(str) {
    return str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  // Lightweight markdown → HTML renderer (no external dependencies)
  function renderMarkdown(src) {
    if (!src) return "";
    let html = escapeHtml(src);

    // Fenced code blocks: ```lang\n...\n```
    html = html.replace(/```(\w*)\n?([\s\S]*?)```/g, (_m, lang, code) =>
      `<pre class="md-codeblock"><code${lang ? ` data-lang="${lang}"` : ""}>${code.trim()}</code></pre>`
    );

    // Inline code: `code`
    html = html.replace(/`([^`\n]+)`/g, '<code class="md-inline-code">$1</code>');

    // Bold
    html = html.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");
    html = html.replace(/__(.+?)__/g, "<strong>$1</strong>");

    // Italic
    html = html.replace(/(?<!\*)\*(?!\*)(.+?)(?<!\*)\*(?!\*)/g, "<em>$1</em>");
    html = html.replace(/(?<!_)_(?!_)(.+?)(?<!_)_(?!_)/g, "<em>$1</em>");

    // Strikethrough
    html = html.replace(/~~(.+?)~~/g, "<del>$1</del>");

    // Markdown links: [text](url)
    html = html.replace(
      /\[([^\]]+)\]\(([^)]+)\)/g,
      '<a href="$2" target="_blank" rel="noopener noreferrer" class="md-link">$1</a>'
    );

    // Auto-link bare URLs (not already inside href)
    html = html.replace(
      /(?<!href="|">)(https?:\/\/[^\s<]+)/g,
      '<a href="$1" target="_blank" rel="noopener noreferrer" class="md-link">$1</a>'
    );

    // @mentions
    html = html.replace(/(^|\s)(@\w+)/g, '$1<span class="md-mention">$2</span>');

    // Blockquotes: > text (after &gt; escape)
    html = html.replace(/^&gt; (.+)$/gm, '<blockquote class="md-blockquote">$1</blockquote>');

    // Line breaks
    html = html.replace(/\n/g, "<br>");

    return html;
  }

  $: rendered = renderMarkdown(text);
</script>

<!-- Styles are in app.css as global .md-* classes -->
<span class="md-text">{@html rendered}</span>

<style>
  /* Scoped fallback in case app.css global classes aren't loaded */
  .md-text :global(strong) { font-weight: 700; }
  .md-text :global(em)     { font-style: italic; }
  .md-text :global(del)    { text-decoration: line-through; opacity: 0.6; }
</style>
