// map.js — Leaflet map with marker and WMS hazard layer toggles.
const HazardMap = {
  map: null,
  marker: null,
  eventMarkers: null,
  wmsLayers: {},
  layerControlEl: null,

  wmsDefinitions: [
    {
      id: 'flood',
      label: 'Flomsoner',
      url: 'https://nve.geodataonline.no/arcgis/services/Flomsoner1/MapServer/WMSServer',
      layers: 'Flomsone_10arsflom,Flomsone_20arsflom,Flomsone_50arsflom,Flomsone_100arsflom,Flomsone_200arsflom',
    },
    {
      id: 'flood_awareness',
      label: 'Flomaktsomhet',
      url: 'https://nve.geodataonline.no/arcgis/services/FlomAktsomhet/MapServer/WMSServer',
      layers: 'Flom_aktsomhetsomrade',
    },
    {
      id: 'landslide',
      label: 'Skredaktsomhet',
      url: 'https://nve.geodataonline.no/arcgis/services/SkredSnoSteinAkt/MapServer/WMSServer',
      layers: 'Aktsomhetsomrade',
    },
    {
      id: 'quick_clay',
      label: 'Kvikkleire',
      url: 'https://nve.geodataonline.no/arcgis/services/KvikkleireskredAktsomhet/MapServer/WMSServer',
      layers: 'KvikkleireskredAktsomhet',
    },
    {
      id: 'avalanche',
      label: 'Sn\u00f8skred',
      url: 'https://nve.geodataonline.no/arcgis/services/SnoskredAktsomhet/MapServer/WMSServer',
      layers: 'S2_snoskred_u_skogeffekt_Aktsomhetsomrade,S3_snoskred_Aktsomhetsomrade',
    },
    {
      id: 'rock_fall',
      label: 'Steinsprang',
      url: 'https://nve.geodataonline.no/arcgis/services/SkredSteinAktR/MapServer/WMSServer',
      layers: 'Utlopsomrade,Utlosningsomrade,Steinsprang-AktsomhetOmrader',
    },
    {
      id: 'combined',
      label: 'Skredfaresoner',
      url: 'https://nve.geodataonline.no/arcgis/services/Skredfaresoner2/MapServer/WMSServer',
      layers: 'Skredsoner_100,Skredsoner_1000',
    },
  ],

  init() {
    this.layerControlEl = document.getElementById('map-layers');
    this.map = L.map('map').setView([65, 14], 5);

    // OSM as fallback base for areas outside Norway
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap',
      maxZoom: 19,
    }).addTo(this.map);

    // Kartverket topo on top — covers Norway with detailed terrain
    L.tileLayer('https://cache.kartverket.no/v1/wmts/1.0.0/topo/default/webmercator/{z}/{y}/{x}.png', {
      attribution: '&copy; <a href="https://www.kartverket.no/">Kartverket</a>',
      maxZoom: 18,
    }).addTo(this.map);

    this.eventMarkers = L.layerGroup().addTo(this.map);

    // Create WMS layers (not added to map until toggled)
    this.wmsDefinitions.forEach(def => {
      this.wmsLayers[def.id] = L.tileLayer.wms(def.url, {
        layers: def.layers,
        format: 'image/png',
        transparent: true,
        opacity: 0.5,
        version: '1.3.0',
      });
    });

    this.renderLayerToggles();
  },

  renderLayerToggles() {
    this.layerControlEl.innerHTML = '';
    this.wmsDefinitions.forEach(def => {
      const label = document.createElement('label');
      const checkbox = document.createElement('input');
      checkbox.type = 'checkbox';
      checkbox.addEventListener('change', () => {
        if (checkbox.checked) {
          this.wmsLayers[def.id].addTo(this.map);
        } else {
          this.map.removeLayer(this.wmsLayers[def.id]);
        }
      });
      label.appendChild(checkbox);
      label.appendChild(document.createTextNode(def.label));
      this.layerControlEl.appendChild(label);
    });
  },

  setLocation(lat, lon, historicalEvents) {
    // Recalculate size after container becomes visible
    this.map.invalidateSize();
    if (this.marker) {
      this.map.removeLayer(this.marker);
    }
    this.marker = L.marker([lat, lon]).addTo(this.map);
    this.map.setView([lat, lon], 14);

    // Clear old event markers and add new ones
    this.eventMarkers.clearLayers();
    if (historicalEvents && historicalEvents.length > 0) {
      historicalEvents.forEach(e => {
        const color = e.building_damage ? '#c0392b' : '#d96830';
        const marker = L.circleMarker([e.latitude, e.longitude], {
          radius: 7,
          fillColor: color,
          color: '#fff',
          weight: 2,
          fillOpacity: 0.85,
        });

        const parts = [`<b>${this.esc(e.type)}</b>`];
        if (e.date) parts.push(this.esc(e.date));
        if (e.location) parts.push(this.esc(e.location));
        parts.push(`${e.distance_m} m fra adressen`);
        if (e.building_damage) parts.push('Bygningsskade');
        if (e.road_damage) parts.push('Vegskade');
        if (e.fatalities > 0) parts.push(`${e.fatalities} omkommet`);
        if (e.description) parts.push(`<i>${this.esc(e.description.substring(0, 200))}</i>`);

        marker.bindPopup(parts.join('<br>'));
        this.eventMarkers.addLayer(marker);
      });
    }
  },

  esc(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  },
};
