## Double-click words and triple-click paragraphs

Double-click these words: streaming renderer, hello_world, context.Context, and can't. A word beside punctuation should not select the whole paragraph.

This sentence includes continuous English words, punctuation, and mixed 中文 text. Double-click different positions to verify English word and Chinese character boundaries.

A word across styles: render**ing**. Double-clicking it should select the complete word rendering. Standalone **bold**, *italic*, and `code_word` text should also select as words.

Triple-click this paragraph to select all of it, including wrapped lines. Narrow the window until it wraps, then click the first and last lines and confirm that adjacent paragraphs are excluded.

This adjacent paragraph checks the triple-click boundary.

```go
hello_world := "double click a word"
second_line := "triple click selection boundary"
```

Inside the code block, a triple click should select only the current source line. After a double or triple click, keep holding and drag to extend by word or paragraph. Pause at the top or bottom edge of the message area to auto-scroll; releasing the pointer should stop it.

Keep a selection and use the wheel. The page should scroll without changing the selection.

After each double or triple click, press Cmd/Ctrl+C and paste to verify the selected text.
