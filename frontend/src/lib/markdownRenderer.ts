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
  const html = highlighter.codeToHtml(text, {
    lang: validLang,
    theme: 'tokyo-night',
  });
  return html;
};
marked.use({ renderer });

export async function renderMarkdown(text: string): Promise<string> {
  // marked.parseAsync was removed or changed in newer versions, use parse with async: true
  return await marked.parse(text, { async: true });
}
