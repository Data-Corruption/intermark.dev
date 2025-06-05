// buildToc dynamically builds a table of contents based on the headers in the container.
// Container should have #_content and #_toc elements. Uses h2-h3 headers in #_content.
// It ignores headers who themselves or parent has the data-notoc attribute.
// Also sets the page title to the first h1 + " | " + current page title.
function buildToc() {
  const content = document.getElementById('_content');
  const toc = document.getElementById('_toc');
  if (!content || !toc || toc.children.length > 0) return;

  // set the title
  const h1 = content.querySelector('h1');
  if (h1) {
    const title = document.title;
    const newTitle = `${h1.textContent.trim()} | ${title}`;
    document.title = newTitle;
  }

  const headers = [...content.querySelectorAll('h2, h3')]
  .filter(h =>
    !h.hasAttribute('data-notoc') &&
    !h.parentElement?.hasAttribute('data-notoc')
  );
  if (headers.length === 0) return;

  const seen = Object.create(null);
  const ul = document.createElement('ul');
  headers.forEach(header => {
    const li = document.createElement('li');
    li.className = [
      'border-l-1',
      'border-base-content/50',
      'hover:border-base-content',
      header.tagName === 'H3' ? 'pl-6' : 'pl-3',
    ].join(' ');

    const topBuffer = 100;
    const a = document.createElement('a');
    a.href = `#${header.id}`;
    a.textContent = header.textContent.trim();
    a.className = 'block py-1 text-base-content/75 hover:text-base-content';
    a.addEventListener('click', e => {
      e.preventDefault();
      const href = a.getAttribute('href');
      history.pushState(null, '', href);
      const el = document.getElementById(header.id);
      const y = el.getBoundingClientRect().top + window.pageYOffset - topBuffer;
      window.scrollTo({ top: y, behavior: 'smooth' });
    });

    li.append(a);
    ul.append(li);
  });

  toc.append(ul);

  // throttle helper
  function throttle(fn, wait = 100) {
    let last = 0;
    return (...args) => {
      const now = Date.now();
      if (now - last > wait) {
        last = now;
        fn.apply(this, args);
      }
    };
  }

  // simple highlight fn
  function setActive(id) {
    toc.querySelectorAll('li.border-base-content')
      .forEach(li => li.classList.replace('border-base-content', 'border-base-content/50'));
    const link = toc.querySelector(`a[href="#${id}"]`);
    if (link) link.parentElement.classList.replace('border-base-content/50', 'border-base-content');
  }

  // throttled scroll listener
  const onScroll = throttle(() => {
    const threshold = window.innerHeight * 0.5;

    // find the last header whose top ≤ threshold
    let current = headers[0];
    for (const h of headers) {
      if (h.getBoundingClientRect().top <= threshold) {
        current = h;
      } else {
        break;
      }
    }

    if (!current || !current.id) return;
    setActive(current.id);
  }, 150);

  // wire it up
  window.addEventListener('scroll', onScroll);
  onScroll(); // initialize on load
}
