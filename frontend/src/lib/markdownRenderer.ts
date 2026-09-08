import { createHighlighter } from 'shiki';
import { marked } from 'marked';

let highlighterPromise: Promise<any> | null = null;

function getHighlighter() {
  if (!highlighterPromise) {
    highlighterPromise = Promise.race([
      createHighlighter({
        themes: ['tokyo-night'],
        langs: [
          'javascript', 'typescript', 'python', 'java', 'cpp', 'c',
          'go', 'rust', 'sql', 'bash', 'json', 'yaml', 'html', 'css',
          'kotlin', 'swift', 'ruby', 'php', 'csharp', 'scala'
        ],
      }),
      new Promise((_, reject) => setTimeout(() => reject(new Error("Shiki highlighter load timeout (WASM block?)")), 3000))
    ]);
  }
  return highlighterPromise;
}

const renderer = new marked.Renderer();
renderer.code = async ({ text, lang }: { text: string; lang?: string | undefined }) => {
  const highlighter = await getHighlighter();
  const validLang = highlighter.getLoadedLanguages().includes(lang ?? '')
    ? lang!
    : 'text';

  const highlighted = highlighter.codeToHtml(text, {
    lang: validLang,
    theme: 'tokyo-night',
  });

  const label = validLang === 'text' ? 'code' : validLang;
  // Escape backticks and backslashes for safe inline onclick embedding
  const escaped = text.replace(/\\/g, '\\\\').replace(/`/g, '\\`').replace(/\$/g, '\\$');

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
  // marked.parseAsync was removed or changed in newer versions, use parse with async: true
  return await marked.parse(text, { async: true });
}
