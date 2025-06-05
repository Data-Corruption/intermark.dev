# Customizing

---

## Title

This is the title of your site, displayed in the browser tab and navbar. You can set it from the `/edit` page.

---

## Icon

### Site Icon (Favicon)

Displayed in the browser tab. To update it:

1. Place your icon in the `./assets` directory.
2. Name it `icon.<ext>`, where `<ext>` is one of: `ico`, `svg`, `png`, `jpg`, `jpeg`, or `avif`.

### Inline Icon (Navbar)

To support theme reactive SVGs, the navbar icon is just raw HTML you set manually from the `/edit` page.

This gives you full control — for example, you can inline an SVG that reacts to the current theme using `currentColor` or DaisyUI’s [semantic colors](https://daisyui.com/docs/colors/) classes.

---

## Sidebar

The sidebar is customizable through the `/edit` page. You can drag and drop items to reorder them. It auto saves on changes. Each item also has a little edit button for more options.

### Non-Filesystem Sidebar Items

You can add non-filesystem items to the sidebar using a `+` button in each folder. Right now there are three of these types:

- Label: A simple text label.
- Link: A link to a URL.
- Divider: A horizontal line to separate items.

These can be dragged and dropped like regular items. They auto save on changes and have a little edit button for more options.

---

## Themes

Intermark comes with all the built-in themes from DaisyUI. You can select a theme from the dropdown on the right of the navbar.

### Default Themes

To set the default light and dark themes, edit the vars at the top of `.assets/js/utils.js`:

<div id="dfc"></div>

### Customizing Themes

You can customize the themes by editing `./public/.meta/app.css`. This file allows you to modify existing themes or create new ones. See the DaisyUI [theming docs](https://daisyui.com/docs/customization/) for details on how to customize themes. Also check out their [theme generator](https://daisyui.com/theme-generator).

**IMPORTANT** - When adding new themes, in order for them to show up in the theme selector, you need to add them to the `All` slice in `./go/themes/themes.go`. You can also hide themes by removing them from the slice:

<div id="atc"></div>

---

## Fonts

### Adding Custom Fonts

To add custom fonts, download them, add them to the `./assets` directory, and then add them to `./public/.meta/app.css`. Here is an example of how to add a custom font so it fully works with tailwind as expected:

<div id="fcss"></div>

Now you can use the `font-inter` class in your HTML to apply the Inter font:

<div id="fhtml"></div>

---

<div class="flex flex-row justify-between mt-10">
  <a href="/p/usage/writing-content" class="btn btn-primary">Previous: Writing Content</a>
  <a href="/p/usage/deployment" class="btn btn-secondary">Next: Deployment</a>
</div>

<script>
  window.addEventListener('load', () => {
    const dfc = `const DEFAULT_LIGHT_THEME = 'nord';
const DEFAULT_DARK_THEME = 'night';`;

    const atc = `package themes

// All the themes to include in the theme selector.
var All = []string{
  "nord", "night", "light", "dark", "cyberpunk", "cupcake",
  "bumblebee", "emerald", "corporate", "synthwave", "retro", "valentine",
  "halloween", "garden", "forest", "aqua", "lofi", "pastel", "fantasy",
  "wireframe", "luxury", "dracula", "cmyk", "autumn", "business", "acid",
  "lemonade", "black", "coffee", "winter", "dim", "sunset", "caramellatte",
  "abyss", "silk",
}`;

    const fcss = `@layer base {
  /* --- Inter Font --- */
  @font-face {
    font-family: 'InterVariable';
    font-style: normal;
    font-weight: 100 900;
    font-display: swap;
    src: url('/assets/fonts/InterVariable.woff2') format('woff2');
  }
  @font-face {
    font-family: 'InterVariable';
    font-style: italic;
    font-weight: 100 900;
    font-display: swap;
    src: url('/assets/fonts/InterVariable-Italic.woff2') format('woff2');
  }
}

@layer components {
  .font-inter {
    font-family: 'InterVariable', theme('fontFamily.sans');
  }
}`;

    const fhtml = `<h1 class="font-inter">Hello, Inter!</h1>
<h1 class="font-inter italic">Hello, Inter italic!</h1>
<h1 class="font-inter font-bold">Hello, Inter bold!</h1>
<h1 class="font-inter italic font-bold">Hello, Inter bold and italic!</h1>`;

    codeBlock('dfc', dfc, 'js');
    codeBlock('atc', atc, 'go');
    codeBlock('fcss', fcss, 'css');
    codeBlock('fhtml', fhtml, 'html');
  });
</script>
