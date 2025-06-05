# Writing Content

---

## Pages

### Adding New Pages

To add a new page, create a new Markdown (`.md`) or HTML (`.html`) file in the `./public` directory. The filename will determine the URL path of the page. For example, `hello-world.md` will be accessible at `/p/hello-word`.

For pages that mix Markdown and HTML, use the `.md` extension.

### Index and Footer

`./public/.index.md` and `./public/.footer.md` are reserved files that define the content of the landing page and footer. The content of these files will be rendered at the root of your site (`/`) and at the bottom of every page, respectively.

For help building pretty landings quickly, I'd check out the [hero component](https://daisyui.com/components/hero/) from DaisyUI.
For footers, their [footer component](https://daisyui.com/components/footer/) is a great starting point.

### Ignored Paths

Any files or directories that start with a dot (e.g., `.thing`) will be ignored by Intermark. This allows you to keep non-content files in the `public` directory without affecting your site.

### Meta data

Intermark keeps meta data and various runtime files in `./public/.meta/*` You don't need to edit any of these files directly, as they are managed by Intermark.

---

## Markdown

### Basic Syntax

`.md` pages can utilize markdown syntax for formatting. Here is a cheat sheet:

<div class="grid grid-cols-5">
  <div class="font-bold bg-base-300 text-center p-4 border">Element</div>
  <div class="font-bold bg-base-300 text-center p-4 border col-span-2">Markdown Syntax</div>
  <div class="font-bold bg-base-300 text-center p-4 border col-span-2">Rendered</div>
  <a href="https://markdownguide.offshoot.io/basic-syntax/#headings" class="flex items-center justify-center p-4 border">Heading</a>
  <div class="p-4 border col-span-2">

```markdown
# H1
## H2
### H3
```

  </div>
  <div class="p-4 border col-span-2" data-notoc data-nosearch>

# H1
## H2
### H3

  </div>
  <a href="https://markdownguide.offshoot.io/basic-syntax/#bold" class="flex items-center justify-center p-4 border">Bold</a>
  <div class="p-4 border col-span-2">

```markdown
**bold text**
```

  </div>
  <div class="p-4 border col-span-2">

**bold text**

  </div>
  <a href="https://markdownguide.offshoot.io/basic-syntax/#italic" class="flex items-center justify-center p-4 border">Italic</a>
  <div class="p-4 border col-span-2">

```markdown
*italicized text*
```

  </div>
  <div class="p-4 border col-span-2">

*italicized text*

  </div>
  <a href="https://markdownguide.offshoot.io/basic-syntax/#blockquotes" class="flex items-center justify-center p-4 border">Blockquote</a>
  <div class="p-4 border col-span-2">

```markdown
> blockquote
```

  </div>
  <div class="p-4 border col-span-2">

> blockquote

  </div>
  <a href="https://markdownguide.offshoot.io/basic-syntax/#ordered-lists" class="flex items-center justify-center p-4 border">Ordered List</a>
  <div class="p-4 border col-span-2">

```markdown
1. First item
2. Second item
3. Third item
```

  </div>
  <div class="p-4 border col-span-2">

1. First item
2. Second item
3. Third item

  </div>
  <a href="https://markdownguide.offshoot.io/basic-syntax/#unordered-lists" class="flex items-center justify-center p-4 border">Unordered List</a>
  <div class="p-4 border col-span-2">

```markdown
- First item
- Second item
- Third item
```

  </div>
  <div class="p-4 border col-span-2">

- First item
- Second item
- Third item

  </div>
  <a href="https://markdownguide.offshoot.io/basic-syntax/#code" class="flex items-center justify-center p-4 border">Code</a>
  <div class="p-4 border col-span-2">

```markdown
`code`
```

  </div>
  <div class="p-4 border col-span-2">

`code`

  </div>
  <a href="https://markdownguide.offshoot.io/basic-syntax/#horizontal-rules" class="flex items-center justify-center p-4 border">Horizontal Rule</a>
  <div class="p-4 border col-span-2">

```markdown
---
```

  </div>
  <div class="p-4 border col-span-2">

---

  </div>
  <a href="https://markdownguide.offshoot.io/basic-syntax/#links" class="flex items-center justify-center p-4 border">Link</a>
  <div class="p-4 border col-span-2">

```markdown
[title](https://www.example.com)
```

  </div>
  <div class="p-4 border col-span-2">

[title](https://www.example.com)

  </div>
  <a href="https://markdownguide.offshoot.io/basic-syntax/#images" class="flex items-center justify-center p-4 border">Image</a>
  <div class="p-4 border col-span-2">

```markdown
![alt text](image.jpg)
```

  </div>
  <div class="p-4 border col-span-2">

![alt text](image.jpg)

  </div>
  <a href="https://markdownguide.offshoot.io/extended-syntax/#tables" class="flex items-center justify-center p-4 border">Table</a>
  <div class="p-4 border col-span-2">

```markdown
| Syntax | Description |
| ----------- | ----------- |
| Header | Title |
| Paragraph | Text |
```

  </div>
  <div class="p-4 border col-span-2">

| Syntax | Description |
| ----------- | ----------- |
| Header | Title |
| Paragraph | Text |

  </div>
  <a href="https://markdownguide.offshoot.io/extended-syntax/#fenced-code-blocks" class="flex items-center justify-center p-4 border">Fenced Code Block</a>
  <div class="p-4 border col-span-2">

````markdown
```json
{
"firstName": "John",
"lastName": "Smith",
"age": 25
}
```
````

  </div>
  <div class="p-4 border col-span-2">

```
{
  "firstName": "John",
  "lastName": "Smith",
  "age": 25
}
```

  </div>
  <a href="https://markdownguide.offshoot.io/extended-syntax/#strikethrough" class="flex items-center justify-center p-4 border">Strikethrough</a>
  <div class="p-4 border col-span-2">

```markdown
~~The world is flat.~~
```

  </div>
  <div class="p-4 border col-span-2">

~~The world is flat.~~

  </div>
  <a href="https://markdownguide.offshoot.io/extended-syntax/#task-lists" class="flex items-center justify-center p-4 border">Task List</a>
  <div class="p-4 border col-span-2">

```markdown
- [x] Write the press release
- [ ] Update the website
- [ ] Contact the media
```

  </div>
  <div class="p-4 border col-span-2">

- [x] Write the press release
- [ ] Update the website
- [ ] Contact the media

  </div>
</div>

### Embedding In HTML

When embedding / nesting Markdown inside HTML (e.g., within a component), you need to be aware of how Markdown interprets indentation and blank lines.

**Key rules**:

1. **Blank Line Before Markdown**: Always insert an empty line before starting a Markdown block inside HTML.
2. **Unindented Markdown**: Markdown content must be left-aligned, even if it's visually nested within HTML tags.

#### Example

<div class="grid grid-cols-2">
  <div class="p-6 border bg-base-300">

<div id="embed_code"></div>

  </div>
  <div class="p-6 border bg-base-300">

<div>
 <div>
  <div data-notoc>

### This will render correctly

<div> ### This won't </div>
### This also won't

### This will
  </div>
 </div>
</div>

  </div>
</div>

### Fancy Code Blocks

Intermark includes [shiki](https://shiki.matsu.io/) via their CDN and exposes it via a window function you can use to render code.

First, create a `<div>` with an ID where you want the code block to appear.

<div id="shiki_1_code"></div>

Then in `<script></script>` tags, call the `codeBlock` function with the target element ID, code, and language:

<div id="shiki_2_code"></div>

For a list of supported languages, check the [shiki docs](https://shiki.matsu.io/languages).

---

## Templates

Intermark uses templates to define the layout of pages under the hood. The data passed into these templates is also available to you in your Markdown and HTML files.

The data available in templates includes:

- [{{< raw >}}{{ .Layout }}{{< /raw >}}]() - The layout of the site. Can be used to iterate over the pages in the sidebar.
  It's a little hacky, but see the [sidebar template]() for an example of how to use it.
- {{< raw >}}{{ .Themes }}{{< /raw >}} - A list of all available themes.
- {{< raw >}}{{ .EditMode }}{{< /raw >}} - A boolean indicating if the site is in edit mode.
- {{< raw >}}{{ .Debug }}{{< /raw >}} - A boolean indicating if debug level logging is enabled.

---

## Escaping

### TOC and Search

If you want to prevent a heading or content from being included in the table of contents or search results, you can add the `data-notoc` and `data-nosearch` attributes to the element or it's **direct** parent. For example:

<div id="escaping_code"></div>

### Prose styling

To prevent Intermark from applying prose styling to a section of HTML, you can use the `not-prose` class. Otherwise, tailwind prose styles will be applied to all content within the element. This is how the markdown content is styled by default.

### Everything

To prevent Intermark from processing a section of content in any way, you can use {{< raw >}}{{< raw >}} and {{< /raw >}}{{< /raw >}} tags. This will render the content exactly as it is, without any processing or formatting.

<div id="raw_code"></div>

---

## Assets

### Adding Files

To add files (images, documents, scripts, etc.) to your site, place them in the `./assets` directory (which is handled by **LFS**). You can then link to these files in your Markdown or HTML content like this:

<div id="asset_code"></div>

### Fingerprinting

Intermark automatically fingerprints assets for cache busting. When you link to an asset, at runtime, Intermark will replace the link with a version that includes a hash of the file content. This ensures that browsers always load the latest version of your assets and only re-download them when they change.

To escape the fingerprinting system, you can add a '/' before the link, like this:

<div id="asset_esc_code"></div>

---

<div class="flex flex-row justify-between mt-10" data-nosearch>
  <a href="/p/getting-started" class="btn btn-primary">Previous: Getting Started</a>
  <a href="/p/usage/customizing" class="btn btn-secondary">Next: Customizing</a>
</div>

<script>
  window.addEventListener('load', () => {
    const embed_code =
`<div>
 <div>
  <div>

### This will render correctly

<div> ### This won't </div>
### This also won't

### This will
  </div>
 </div>
</div>`;

    const escaping_code =
`<h2 data-notoc data-nosearch>This will not appear in TOC</h2>

<p data-nosearch>This will not appear in search results.</p>

<div data-notoc data-nosearch>
  <h3>This will not appear in TOC</h3>
  <p>This will not appear in search results.</p>
</div>`;

    const asset_esc_code =
`![logo](/assets/logo.jpg) -> ![logo](/a/<hash>.jpg) # default
![logo](//assets/logo.jpg) -> ![logo](/assets/logo.jpg) # escaped
![logo](///assets/logo.jpg) -> ![logo](//assets/logo.jpg) # escaped`;

    const shiki_1_code =
`# Example Code Block

<div id="example_code"></div>
`

    const shiki_2_code =
`window.addEventListener('load', () => {
  // args: target id, code, language
  codeBlock('example_code', 'console.log("Hello, World!");', 'javascript');
});`;

    const raw_code =
`{{< raw >}}{{< raw >}}

# This will not be processed

{{ .Layout.Title }} # Will be rendered as literal text

{{< /raw >}}{{< /raw >}}`;

    codeBlock('embed_code', embed_code, 'markdown');
    codeBlock('escaping_code', escaping_code, 'html');
    codeBlock('asset_code', '![logo](/assets/logo.jpg)', 'markdown');
    codeBlock('asset_esc_code', asset_esc_code, 'markdown');
    codeBlock('shiki_1_code', shiki_1_code, 'markdown');
    codeBlock('shiki_2_code', shiki_2_code, 'javascript');
    codeBlock('raw_code', raw_code, 'markdown');
  });
</script>
