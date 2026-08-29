import { createHighlighter } from 'shiki';
import { marked } from 'marked';

let highlighterPromise: Promise<any> | null = null;

function getHighlighter() {
  if (!highlighterPromise) {
    highlighterPromise = createHighlighter({
      themes: ['tokyo-night'],
      langs: [
        'javascript', 'typescript', 'python', 'java', 'cpp', 'c',
        'go', 'rust', 'sql', 'bash', 'json', 'yaml', 'html', 'css',
        'kotlin', 'swift', 'ruby', 'php', 'csharp', 'scala'
      ],
    });
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
  return await marked.parseAsync(text);
}
