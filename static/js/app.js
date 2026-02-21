// app.js — Main controller wiring search, dashboard, and map.
document.addEventListener('DOMContentLoaded', () => {
  Dashboard.init();
  HazardMap.init();

  Search.init(async (address) => {
    const loading = document.getElementById('loading');
    const dashboard = document.getElementById('dashboard');

    dashboard.hidden = true;
    loading.hidden = false;

    try {
      const data = await Api.risk(address);
      loading.hidden = true;
      Dashboard.render(data);
      HazardMap.setLocation(address.latitude, address.longitude);
    } catch (err) {
      loading.hidden = true;
      console.error('Risk assessment error:', err);
      alert('Kunne ikke hente risikovurdering. Prøv igjen senere.');
    }
  });
});
