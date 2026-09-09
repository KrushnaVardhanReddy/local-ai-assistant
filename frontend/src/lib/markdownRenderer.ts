import { createHighlighter } from 'shiki';
import { marked } from 'marked';

let highlighter: any = null;

// Initialize Shiki asynchronously in the background
createHighlighter({
  themes: ['tokyo-night'],
  langs: [
    'javascript', 'typescript', 'python', 'java', 'cpp', 'c',
    'go', 'rust', 'sql', 'bash', 'json', 'yaml', 'html', 'css',
    'kotlin', 'swift', 'ruby', 'php', 'csharp', 'scala'
  ],
}).then(h => {
  highlighter = h;
}).catch(console.error);

const renderer = new marked.Renderer();
(renderer as any).code = ({ text, lang }: { text: string; lang?: string | undefined }) => {
  const validLang = lang || 'text';
  const label = validLang === 'text' ? 'code' : validLang;
  
  // Escape backticks and backslashes for safe inline onclick embedding
  const escaped = text.replace(/\\/g, '\\\\').replace(/`/g, '\\`').replace(/\$/g, '\\$');

  let highlighted = '';
  if (highlighter && highlighter.getLoadedLanguages().includes(validLang)) {
    highlighted = highlighter.codeToHtml(text, {
      lang: validLang,
      theme: 'tokyo-night',
    });
  } else {
    // Fallback unstyled code block while Shiki is loading
    const escapedHtml = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    highlighted = `<pre><code class="language-${validLang}">${escapedHtml}</code></pre>`;
  }

  return `<div class="code-block-wrapper">
  <div class="code-block-header">
    <span>${label}</span>
    <button class="code-copy-btn" onclick="(function(btn){
      navigator.clipboard.writeText(\`${escaped}\`).then(()=>{
        btn.textContent='Copied!';
        setTimeout(()=>btn.textContent='Copy',1500);
      });
    })(this)">Copy</button>
  </div>
  ${highlighted}
</div>`;
};
marked.use({ renderer });

export async function renderMarkdown(text: string): Promise<string> {
  return await marked.parse(text);
}
