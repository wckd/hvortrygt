// search.js — Address autocomplete with debounce and keyboard navigation.
const Search = {
  input: null,
  resultsEl: null,
  activeIndex: -1,
  results: [],
  debounceTimer: null,
  onSelect: null,

  init(onSelect) {
    this.input = document.getElementById('search-input');
    this.resultsEl = document.getElementById('search-results');
    this.onSelect = onSelect;

    this.input.addEventListener('input', () => this.onInput());
    this.input.addEventListener('keydown', (e) => this.onKeydown(e));
    document.addEventListener('click', (e) => {
      if (!this.input.contains(e.target) && !this.resultsEl.contains(e.target)) {
        this.hideResults();
      }
    });
  },

  onInput() {
    clearTimeout(this.debounceTimer);
    const q = this.input.value.trim();
    if (q.length < 2) {
      this.hideResults();
      return;
    }
    this.debounceTimer = setTimeout(() => this.doSearch(q), 250);
  },

  async doSearch(query) {
    try {
      this.results = await Api.search(query);
      this.activeIndex = -1;
      this.renderResults();
    } catch (err) {
      console.error('Search error:', err);
      this.hideResults();
    }
  },

  renderResults() {
    if (this.results.length === 0) {
      this.hideResults();
      return;
    }

    this.resultsEl.innerHTML = '';
    this.results.forEach((addr, i) => {
      const div = document.createElement('div');
      div.className = 'result-item';
      div.setAttribute('role', 'option');
      div.innerHTML = `
        <div class="result-main">${this.escapeHtml(addr.text)}</div>
        <div class="result-sub">${this.escapeHtml(addr.postnummer || '')} ${this.escapeHtml(addr.poststed || '')}, ${this.escapeHtml(addr.kommunenavn || '')}</div>
      `;
      div.addEventListener('click', () => this.selectResult(i));
      this.resultsEl.appendChild(div);
    });

    this.resultsEl.hidden = false;
    this.input.setAttribute('aria-expanded', 'true');
  },

  onKeydown(e) {
    if (this.resultsEl.hidden) return;

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      this.activeIndex = Math.min(this.activeIndex + 1, this.results.length - 1);
      this.highlightActive();
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      this.activeIndex = Math.max(this.activeIndex - 1, 0);
      this.highlightActive();
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (this.activeIndex >= 0) {
        this.selectResult(this.activeIndex);
      }
    } else if (e.key === 'Escape') {
      this.hideResults();
    }
  },

  highlightActive() {
    const items = this.resultsEl.querySelectorAll('.result-item');
    items.forEach((el, i) => {
      el.classList.toggle('active', i === this.activeIndex);
    });
  },

  selectResult(index) {
    const addr = this.results[index];
    if (!addr) return;
    this.input.value = addr.text;
    this.hideResults();
    if (this.onSelect) this.onSelect(addr);
  },

  hideResults() {
    this.resultsEl.hidden = true;
    this.resultsEl.innerHTML = '';
    this.input.setAttribute('aria-expanded', 'false');
    this.activeIndex = -1;
  },

  escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  },
};
