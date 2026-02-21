// dashboard.js — Risk card rendering, score banner, weather alerts.
const Dashboard = {
  bannerEl: null,
  cardsEl: null,
  alertsEl: null,
  dashboardEl: null,

  init() {
    this.bannerEl = document.getElementById('score-banner');
    this.cardsEl = document.getElementById('hazard-cards');
    this.alertsEl = document.getElementById('weather-alerts');
    this.dashboardEl = document.getElementById('dashboard');
  },

  render(data) {
    this.renderBanner(data);
    this.renderAlerts(data.weather_alerts || []);
    this.renderCards(data.hazards || [], data.historical_events || []);
    this.dashboardEl.hidden = false;
  },

  renderBanner(data) {
    const levelLabels = {
      low: 'Lav risiko',
      medium: 'Moderat risiko',
      high: 'Høy risiko',
      very_high: 'Svært høy risiko',
    };

    this.bannerEl.className = `score-banner level-${this.safeLevel(data.overall_level)}`;
    this.bannerEl.innerHTML = `
      <div class="score-number">${Number(data.overall_score) || 0}</div>
      <div class="score-label">${levelLabels[data.overall_level] || ''}</div>
      <div class="score-summary">${this.esc(data.summary)}</div>
      <div class="score-address">${this.esc(data.address.text)}${data.elevation != null ? ` (${data.elevation.toFixed(1)} moh.)` : ''}</div>
    `;
  },

  renderAlerts(alerts) {
    this.alertsEl.innerHTML = '';
    if (alerts.length === 0) return;

    alerts.forEach(alert => {
      const div = document.createElement('div');
      const allowedSeverity = ['Extreme', 'Severe', 'Moderate', 'Minor'];
      const safeSeverity = allowedSeverity.includes(alert.severity) ? alert.severity : 'Minor';
      div.className = `alert-card severity-${safeSeverity}`;
      div.innerHTML = `
        <div class="alert-event">${this.esc(alert.event)} — ${this.esc(alert.severity)}</div>
        <div class="alert-desc">${this.esc(alert.description)}</div>
      `;
      this.alertsEl.appendChild(div);
    });
  },

  renderCards(hazards, historicalEvents) {
    // Sort: highest score first, errors last
    hazards.sort((a, b) => {
      if (a.error && !b.error) return 1;
      if (!a.error && b.error) return -1;
      return b.score - a.score;
    });

    this.cardsEl.innerHTML = '';
    hazards.forEach(h => {
      const div = document.createElement('div');
      div.className = `hazard-card level-${this.safeLevel(h.level)}`;
      div.innerHTML = `
        <div class="hazard-name">${this.esc(h.name)}</div>
        ${h.error
          ? `<div class="hazard-error">${this.esc(h.error)}</div>`
          : `
            <div class="hazard-score">${Number(h.score) || 0}</div>
            <div class="hazard-level">${this.levelText(h.level)}</div>
            <div class="hazard-details">${this.esc(h.details || h.description)}</div>
          `}
      `;

      // Add mini event list for historical landslides card
      if (h.id === 'historical_landslides' && !h.error && historicalEvents.length > 0) {
        const list = document.createElement('div');
        list.className = 'event-list';
        const shown = historicalEvents.slice(0, 5);
        shown.forEach(e => {
          const item = document.createElement('div');
          item.className = 'event-item';
          const tags = [];
          if (e.building_damage) tags.push('bygning');
          if (e.road_damage) tags.push('veg');
          if (e.fatalities > 0) tags.push(`${Number(e.fatalities)} omkommet`);
          item.innerHTML = `
            <span class="event-type">${this.esc(e.type)}</span>
            <span class="event-meta">${e.date ? this.esc(e.date) : 'ukjent dato'} &middot; ${Number(e.distance_m) || 0} m${tags.length ? ' &middot; ' + this.esc(tags.join(', ')) : ''}</span>
          `;
          list.appendChild(item);
        });
        if (historicalEvents.length > 5) {
          const more = document.createElement('div');
          more.className = 'event-item event-more';
          more.textContent = `+ ${historicalEvents.length - 5} flere hendelser`;
          list.appendChild(more);
        }
        div.appendChild(list);
      }

      this.cardsEl.appendChild(div);
    });
  },

  safeLevel(level) {
    const allowed = ['low', 'medium', 'high', 'very_high', 'unknown'];
    return allowed.includes(level) ? level : 'unknown';
  },

  levelText(level) {
    const m = { low: 'Lav', medium: 'Moderat', high: 'Høy', very_high: 'Svært høy', unknown: 'Ukjent' };
    return m[level] || level;
  },

  esc(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  },
};
