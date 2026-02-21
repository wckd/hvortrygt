# Hvor trygt bor du?

Sjekk naturfare for enhver norsk adresse. Skriv inn en adresse og få en risikovurdering basert på åpne data fra NVE, Kartverket og MET.

## Hva sjekkes

|Fare | Kilde | Beskrivelse |
|------|-------|-------------|
| Flomsoner | NVE | 10-, 20-, 50-, 100- og 200-årsflom |
| Flomaktsomhet | NVE | Generelle aktsomhetsområder for flom |
| Jord- og flomskred | NVE | Aktsomhetsområder for jord- og flomskred |
| Kvikkleire | NVE | Detaljert faregrad + aktsomhetsområder |
| Snøskred | NVE | Aktsomhetsområder for snøskred |
| Steinsprang | NVE | Aktsomhetsområder for steinsprang |
| Skredfaresoner | NVE | Kartlagte faresoner (100- og 1000-år) |
| Stormflo | Kartverket | Konsekvensdata for kystkommuner |
| Værvarsler | MET | Aktive farevarsler (MetAlerts) |
| Høyde | Kartverket | Høyde over havet for risikojustering |

## Risikoscore

Hver fare gir en score fra 0 til 100. Den høyeste enkeltverdien bestemmer totalscoren.

- **0–15** Lav risiko (grønn)
- **16–40** Moderat risiko (gul)
- **41–70** Høy risiko (oransje)
- **71–100** Svært høy risiko (rød)

Adresser under 5 moh. i kystkommuner får +10 poeng.

## Kjør lokalt

```
go run .
```

Åpne http://localhost:8080. Ingen database, ingen API-nøkler, ingen konfigurasjon.

Annen port:

```
go run . -port 3000
# eller
PORT=3000 go run .
```

## Docker

```
docker compose up --build
```

Bygger et minimalt Alpine-image (~10 MB) og starter på port 8080.

## Bygg binær

```
go build -o hvortrygt .
./hvortrygt
```

Alt av statiske filer (HTML, CSS, JS) er innebygd i binæren via `embed.FS`.

## Arkitektur

```
Bruker → Adressesøk (Kartverket) → Velg adresse
                                        ↓
                              Backend fan-out (parallelt):
                              ├── NVE: 8 ArcGIS-spørringer
                              ├── Kartverket: Høyde + Stormflo
                              └── MET: Værvarsler
                                        ↓
                              Risikoscore + Dashboard + Kart
```

- **Backend:** Go stdlib `net/http`, ingen rammeverk
- **Frontend:** Vanilla JS, Leaflet for kart
- **Kart:** Kartverket topografisk (WMTS) med OpenStreetMap som fallback
- **Farelag:** NVE WMS-lag som kan toggles på kartet
- **Cache:** In-memory med TTL (NVE 1t, høyde/stormflo 24t, værvarsler 5min)

## Datakilder

Alle API-er er åpne og gratis. MET krever `User-Agent`-header (satt automatisk).

- [NVE Kartdata](https://www.nve.no/kart/) — Flom, skred, kvikkleire
- [Kartverket Adresser](https://ws.geonorge.no/adresser/v1/) — Geokoding
- [Kartverket Høydedata](https://ws.geonorge.no/hoydedata/v1/) — Terrengdata
- [Kartverket Stormflo](https://stormflo-konsekvens.kartverket.no/) — Konsekvensdata
- [MET MetAlerts](https://api.met.no/weatherapi/metalerts/2.0/) — Farevarsler

## Begrensninger

- Kun veiledende — erstatter ikke profesjonell geoteknisk vurdering
- NVE-data dekker ikke hele landet; områder utenfor kartlagte soner betyr ikke nødvendigvis fravær av fare
- Stormflodata er på kommunenivå, ikke punktnivå
- Cache tømmes ved restart (ingen persistering)
