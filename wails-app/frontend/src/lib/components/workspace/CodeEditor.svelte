<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { EditorView, basicSetup } from 'codemirror';
  import { markdown } from '@codemirror/lang-markdown';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { EditorState, StateEffect, StateField, RangeSetBuilder } from '@codemirror/state';
  import { Decoration } from '@codemirror/view';
  import type { DecorationSet } from '@codemirror/view';

  let {
    content = '',
    readOnly = false,
    fontSize = 1,
    lineHeight = 1.5,
    activeLine = 0,
    onChange
  } = $props<{
    content?: string;
    readOnly?: boolean;
    fontSize?: number;
    lineHeight?: number;
    activeLine?: number;
    onChange?: (text: string) => void;
  }>();

  let editorContainer: HTMLDivElement;
  let view: EditorView | null = null;
  let currentContent = $state(content);
  $effect(() => {
    if (content !== currentContent) {
      currentContent = content;
    }
  });

  const customTheme = EditorView.theme({
    "&": {
      backgroundColor: "transparent !important",
      fontSize: "var(--editor-font-size, 1rem)",
      lineHeight: "var(--editor-line-height, 1.5)"
    },
    ".cm-content": {
      fontFamily: "system-ui, -apple-system, sans-serif"
    },
    ".active-speech-line": {
      backgroundColor: "rgba(74, 222, 128, 0.2)",
      borderLeft: "4px solid #4ade80",
    },
    "&.cm-focused": {
      outline: "none"
    }
  });

  const activeLineEffect = StateEffect.define<number>();

  const activeLineField = StateField.define<DecorationSet>({
    create() {
      return Decoration.none;
    },
    update(decorations, tr) {
      decorations = decorations.map(tr.changes);
      for (let e of tr.effects) {
        if (e.is(activeLineEffect)) {
          let lineIdx = e.value;
          if (lineIdx < 0) lineIdx = 0;
          let docLineCount = tr.state.doc.lines;
          if (lineIdx >= docLineCount) lineIdx = docLineCount - 1;

          let line = tr.state.doc.line(lineIdx + 1); // 1-based indexing
          const builder = new RangeSetBuilder<Decoration>();
          builder.add(line.from, line.from, Decoration.line({ class: "active-speech-line" }));
          return builder.finish();
        }
      }
      return decorations;
    },
    provide: f => EditorView.decorations.from(f)
  });

  onMount(() => {
    const updateListener = EditorView.updateListener.of((update) => {
      if (update.docChanged && onChange) {
        currentContent = update.state.doc.toString();
        onChange(currentContent);
      }
    });

    const extensions = [
      basicSetup,
      markdown(),
      oneDark,
      customTheme,
      updateListener,
      activeLineField,
      EditorView.lineWrapping
    ];

    if (readOnly) {
      extensions.push(EditorState.readOnly.of(true));
    }

    view = new EditorView({
      doc: content,
      extensions,
      parent: editorContainer
    });
  });

  onDestroy(() => {
    if (view) {
      view.destroy();
    }
  });

  $effect(() => {
    if (view && content !== currentContent) {
      currentContent = content;
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: content }
      });
    }
  });

  $effect(() => {
    if (view) {
      view.dispatch({
        effects: activeLineEffect.of(activeLine)
      });

      try {
        let lineIdx = activeLine;
        let docLineCount = view.state.doc.lines;
        if (lineIdx < 0) lineIdx = 0;
        if (lineIdx >= docLineCount) lineIdx = docLineCount - 1;

        const line = view.state.doc.line(lineIdx + 1);
        view.dispatch({
          effects: EditorView.scrollIntoView(line.from, { y: 'center' })
        });
      } catch (err) {
        console.error("Failed to scroll to line", err);
      }
    }
  });
</script>

<div
  class="code-editor-wrapper"
  style="--editor-font-size: {fontSize}rem; --editor-line-height: {lineHeight};"
>
  <div bind:this={editorContainer} class="editor-container"></div>
</div>

<style>
  .code-editor-wrapper {
    width: 100%;
    height: 100%;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .editor-container {
    flex-grow: 1;
    overflow: auto;
    width: 100%;
    height: 100%;
  }

  :global(.cm-editor) {
    height: 100%;
  }
</style>
