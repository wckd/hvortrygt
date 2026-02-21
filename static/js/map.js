// map.js — Leaflet map with marker and WMS hazard layer toggles.
const HazardMap = {
  map: null,
  marker: null,
  wmsLayers: {},
  layerControlEl: null,

  wmsDefinitions: [
    {
      id: 'flood',
      label: 'Flomsoner',
      url: 'https://nve.geodataonline.no/arcgis/services/Flomsoner1/MapServer/WMSServer',
      layers: '11,12,13,14,15',
    },
    {
      id: 'flood_awareness',
      label: 'Flomaktsomhet',
      url: 'https://nve.geodataonline.no/arcgis/services/FlomAktsomhet/MapServer/WMSServer',
      layers: '1',
    },
    {
      id: 'landslide',
      label: 'Skredaktsomhet',
      url: 'https://nve.geodataonline.no/arcgis/services/SkredSnoSteinAkt/MapServer/WMSServer',
      layers: '0',
    },
    {
      id: 'quick_clay',
      label: 'Kvikkleire',
      url: 'https://nve.geodataonline.no/arcgis/services/KvikkleireskredAktsomhet/MapServer/WMSServer',
      layers: '0',
    },
    {
      id: 'avalanche',
      label: 'Snoskred',
      url: 'https://nve.geodataonline.no/arcgis/services/SnoskredAktsomhet/MapServer/WMSServer',
      layers: '1',
    },
    {
      id: 'rock_fall',
      label: 'Steinsprang',
      url: 'https://nve.geodataonline.no/arcgis/services/SkredSteinAktR/MapServer/WMSServer',
      layers: '0,1,2',
    },
    {
      id: 'combined',
      label: 'Skredfaresoner',
      url: 'https://nve.geodataonline.no/arcgis/services/Skredfaresoner2/MapServer/WMSServer',
      layers: '2,3',
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

  setLocation(lat, lon) {
    // Recalculate size after container becomes visible
    this.map.invalidateSize();
    if (this.marker) {
      this.map.removeLayer(this.marker);
    }
    this.marker = L.marker([lat, lon]).addTo(this.map);
    this.map.setView([lat, lon], 14);
  },
};
