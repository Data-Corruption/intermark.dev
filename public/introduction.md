# Introduction

Most sites want the same few things:
Markdown, a sidebar, search, some control over layout, and a way to drop down when that’s not enough.

Intermark gives you exactly that. No plugins. No magic.

* Markdown or HTML, mixed freely
* TailwindCSS and DaisyUI prewired
* Search, themes, and sidebar editing built-in
* Go backend, ~4k LOC, minimal framework

---

## Philosophy: The middle is the problem

There are three levels of tooling:

* **High-level** - Simple, easy, and rigid. Think Notion, Wix, or even some SSGs with opinionated themes. Great until you want to do something they didn’t plan for.
* **Middle-level** - Complex templating systems, plugin frameworks, configs. Claims to offer the best of both worlds.
* **Low-level** - HTML, CSS, JS, Go. Flexible, powerful, and brutally honest. You can build anything once you’ve past the learning curve.

The middle-level is seductive. It promises ease and flexibility, but bridging those two is extraordinarily difficult. Much harder than most realize or admit. The result is often leaky abstractions juuust big enough to resist attempts to modify or fully understand. This is where most modern software rots.

Intermark’s philosophy is simple: cut out the middle as much as possible. Tooling should have very well made high and low levels with as little friction as possible jumping between the two. It should also stay small enough to fit in your head all at once.

---

## Dependencies

Dependencies are liabilities:

* Security surface
* Performance cost
* Maintenance drag
* Cognitive load

Most dependencies can be replaced with few hundred lines of std-lib, that's especially true with golang. Intermark keeps dependencies rare. Each one solves a non-trivial problem and earns its import:

* `goldmark` — fast robust Markdown converter
* `tailwindcss` — utility-first CSS that scales without a DSL
* `daisyui` — themeable components

No giant transitive dependency trees. No abstraction tax. The rest of the imports are tiny, optional, and replaceable with std-lib if needed, e.g. sha256 speedup lib.

---

## Who it's for

* You like tools that fail early and loudly, prefer HTML over templating syntax, and view "use a plugin” as a red flag.
* You want to be able to read and understand the entire codebase in an afternoon.
* You're not afraid of tinkering a bit.
* You want something small enough to hack, complete enough to ship™

If any of that resonates, Intermark is for you. <3

---

<div class="relative w-full h-10">
  <a href="/p/getting-started" class="btn btn-secondary absolute inset-y-0 right-0">Next: Getting Started</a>
</div>
