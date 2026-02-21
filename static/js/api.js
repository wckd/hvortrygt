// api.js — Fetch wrapper for backend endpoints.
const Api = {
  async search(query) {
    const resp = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
    if (!resp.ok) throw new Error(`Search failed: ${resp.status}`);
    return resp.json();
  },

  async risk(address) {
    const params = new URLSearchParams({
      lat: address.latitude,
      lon: address.longitude,
      knr: address.kommunenummer,
      text: address.text || '',
      kommune: address.kommunenavn || '',
    });
    const resp = await fetch(`/api/risk?${params}`);
    if (!resp.ok) throw new Error(`Risk assessment failed: ${resp.status}`);
    return resp.json();
  },
};
