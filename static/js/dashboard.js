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
    this.renderCards(data.hazards || []);
    this.dashboardEl.hidden = false;
  },

  renderBanner(data) {
    const levelLabels = {
      low: 'Lav risiko',
      medium: 'Moderat risiko',
      high: 'Høy risiko',
      very_high: 'Svært høy risiko',
    };

    this.bannerEl.className = `score-banner level-${data.overall_level}`;
    this.bannerEl.innerHTML = `
      <div class="score-number">${data.overall_score}</div>
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
      div.className = `alert-card severity-${alert.severity}`;
      div.innerHTML = `
        <div class="alert-event">${this.esc(alert.event)} — ${this.esc(alert.severity)}</div>
        <div class="alert-desc">${this.esc(alert.description)}</div>
      `;
      this.alertsEl.appendChild(div);
    });
  },

  renderCards(hazards) {
    // Sort: highest score first, errors last
    hazards.sort((a, b) => {
      if (a.error && !b.error) return 1;
      if (!a.error && b.error) return -1;
      return b.score - a.score;
    });

    this.cardsEl.innerHTML = '';
    hazards.forEach(h => {
      const div = document.createElement('div');
      div.className = `hazard-card level-${h.level || 'unknown'}`;
      div.innerHTML = `
        <div class="hazard-name">${this.esc(h.name)}</div>
        ${h.error
          ? `<div class="hazard-error">${this.esc(h.error)}</div>`
          : `
            <div class="hazard-score">${h.score}</div>
            <div class="hazard-level">${this.levelText(h.level)}</div>
            <div class="hazard-details">${this.esc(h.details || h.description)}</div>
          `}
      `;
      this.cardsEl.appendChild(div);
    });
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
