## Footnotes and cross-block reference links

The first paragraph cites the [Go documentation][go-docs], whose definition is in a later block. It also has a footnote[^note].

An ordinary paragraph separates the reference from its definition, ensuring they are parsed in different blocks. The same footnote is cited again[^note]. Another footnote contains multiple paragraphs[^detail].

Shortcut reference: [Go home]. Collapsed reference: [Concurrency patterns][].

[same-block]: https://go.dev/doc/
[Same-block link control][same-block]

Definitions follow. Once resolved, they should not appear as prose. Links should support both clicking and starting a text drag.

[go-docs]: https://go.dev/doc/
[Go home]: https://go.dev/
[Concurrency patterns]: https://go.dev/blog/pipelines

[^note]: This is the footnote body. Repeated citations should point to the same note and support returning to the citation.

[^detail]: First paragraph of the footnote.

    Second paragraph with **bold text** and `inline code`.
