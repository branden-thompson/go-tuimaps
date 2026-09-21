# AI-5 — Weather overlay data sources, formats and input shapes

| Field | Value |
|---|---|
| Phase | DISCOVER — Tier 1 research |
| Date | 2026-09-18 (sizes and counts measured live this date unless a documentation URL is cited) |
| Scope | Which overlay kinds the data contract must accept, from which keyless sources, at what volume (OQ-3, OQ-4, R-3, R-4, RS-2, RS-3) |
| Status | Section 5 is research input for PLAN. Terms are reported as published by each provider; this is not legal advice. Nothing is decided until HUM LEAD rules. |
| Verification | Re-measured by the coordinator against the live NWS feed on 2026-09-18: 410 active alerts, 370 (90%) with null geometry; the 40 polygons present had 5–21 vertices (median 7); geometry-less alerts referenced a median of 1 and a maximum of 58 zones. Consistent with section 1. |

All sizes and counts below were measured live on 2026-09-18 unless a documentation URL is cited. "Keyless" means no key, token or account is needed. Two summary points first:
- Every source in sections 1–4 is keyless. The only terms that restrict use belong to Open-Meteo and RainViewer (section 7).
- I did not fetch terms pages for USGS, FIRMS, SPC, NHC or NOMADS, so their redistribution terms are UNVERIFIED.

## 1. Alert geometry (vector)

| Source | URL | Format / CRS | Measured payload | Cadence | Coverage | Terms |
|---|---|---|---|---|---|---|
| NWS active alerts | https://api.weather.gov/alerts/active | GeoJSON, lon/lat | 2.16 MB; 487 alerts, 442 (91%) with `geometry: null` | past 7 days at `/alerts` (docs below); cadence UNVERIFIED | US | "All of the information presented via the API is intended to be open data, free to use for any purpose." "A User Agent is required to identify your application." (https://www.weather.gov/documentation/services-web-api) |
| NWS zones | `https://api.weather.gov/zones/{forecast\|county\|fire\|marine}/{id}` | GeoJSON; geometry is often a `GeometryCollection`, sometimes `MultiPolygon` | TXZ213: 914 vertices, 134 KB. FLC017: 14,001 vertices, 2.09 MB. GMZ335: 7,732 vertices, 892 KB | static; cadence UNVERIFIED | US | as above |
| SPC outlooks | https://www.spc.noaa.gov/products/outlook/day1otlk_cat.lyr.geojson (`.nolyr` variant; index at https://www.spc.noaa.gov/gis/) | GeoJSON; CRS not stated on the GIS page | 4.9 KB, 2 features, 210 vertices | issue times UNVERIFIED | US | terms UNVERIFIED |
| NHC forecast products | https://www.nhc.noaa.gov/gis/ lists shapefile zip and KMZ only. A keyless ArcGIS MapServer also exists: https://mapservices.weather.noaa.gov/tropical/rest/services/tropical/NHC_tropical_weather/MapServer | MapServer query formats: "JSON, geoJSON, PBF". Layers include "AT1 Forecast Track/Cone/Watch-Warning". Native CRS is GCS_Sphere (lon/lat) | not measured (depends on active storms) | per advisory; cadence UNVERIFIED | Atlantic and Pacific basins | "provided as a convenience to users" (https://www.nhc.noaa.gov/gis/) |
| NIFC WFIGS perimeters | https://services3.arcgis.com/T4QMspbfLg3qTGWY/arcgis/rest/services/WFIGS_Interagency_Perimeters_Current/FeatureServer/0 | GeoJSON, JSON or PBF; EPSG:4326; max 2,000 records | 184 perimeters. 20 features = 5.4 MB (min 125, median 5,117, max 47,068 vertices per feature). With `maxAllowableOffset=0.01&geometryPrecision=3` the same query is 79 KB | cadence UNVERIFIED | US | liability disclaimer only; no reuse restriction stated (https://www.arcgis.com/sharing/rest/content/items/d1c32af3212341869b3c810f1a215824?f=json) |
| USGS earthquakes | https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_day.geojson | GeoJSON points `[lon, lat, depth]` | 201 KB for 282 points; `2.5_week`: 250 KB for 352 points | "Updated every minute" (https://earthquake.usgs.gov/earthquakes/feed/v1.0/geojson.php) | global | terms UNVERIFIED |
| FIRMS fire detections | https://firms.modaps.eosdis.nasa.gov/data/active_fire/noaa-20-viirs-c2/csv/J1_VIIRS_C2_USA_contiguous_and_Hawaii_24h.csv | CSV points | 257 KB (HTTP 200 with no key) | cadence UNVERIFIED | US file | terms UNVERIFIED |

- **Null geometry is the normal case.** 91% of active alerts had no geometry.
  - Each such alert lists `affectedZones` URLs and carries `geocode.UGC` and `geocode.SAME` codes.
  - Its shape comes from dereferencing each zone URL, one call per zone.
  - The 45 alerts that did carry polygons had 5 to 21 vertices (median 7).
- **Zone polygons are large.** They run from 914 to 14,001 vertices in my samples, which far exceeds a 160×96-dot canvas. Either the host or the library must simplify and clip them.
- **Styling fields available on alerts:**
  - `severity`: Minor 339, Moderate 93, Severe 50, Unknown 5 in my sample.
  - `urgency`, `certainty`, `event`, `eventCode`, `messageType`, `onset`, `expires`.
- **SPC ships its own colours.** Each SPC feature has `stroke`, `fill`, `LABEL` and `LABEL2` properties.

## 2. Precipitation and radar (raster)

| Source | URL | Tile scheme | Pixels | Cadence | Coverage | Terms |
|---|---|---|---|---|---|---|
| IEM NEXRAD tiles | `https://mesonet.agron.iastate.edu/cache/tile.py/1.0.0/nexrad-n0q-900913/{z}/{x}/{y}.png` (documented at https://mesonet.agron.iastate.edu/ogc/) | XYZ, Web Mercator | one tile measured: 256×256, 8-bit RGBA, 4.9 KB, pre-coloured | 5 min; `-mXXm` layers cover 5–55 minutes back; `/cache/` sends a 5-minute cache header, `/c/` 14 days | CONUS | "The materials found on this website are in the public domain and may be used freely by anyone for any lawful purpose." "Attributing the Iowa Environmental Mesonet of Iowa State University would be appreciated." (https://mesonet.agron.iastate.edu/disclaimer.php) |
| IEM raw composite | https://mesonet.agron.iastate.edu/docs/nexrad_mosaic/ | single EPSG:4326 PNG, 12200×5400 pixels (~0.005°) | pixel value encodes dBZ at "0.5 dBZ increment"; the exact index-to-dBZ formula is UNVERIFIED | 5 min | CONUS | as above |
| NCEP MRMS WMS | `https://opengeo.ncep.noaa.gov/geoserver/conus/conus_bref_qcd/ows` | WMS; CRS:84, EPSG:3857, EPSG:4326, EPSG:4269; any WIDTH/HEIGHT | pre-coloured PNG; a 256×256 GetMap was 1.7 KB | the TIME dimension showed roughly 2-minute steps | CONUS | capabilities document: Fees "none", AccessConstraints "none" |
| nowCOAST | `https://nowcoast.noaa.gov/geoserver/weather_radar/wms` | WMS, EPSG:3857 offered; layers for CONUS, Alaska, Hawaii, Caribbean, Guam | pre-coloured | cadence UNVERIFIED | US and territories | Fees "NONE", AccessConstraints "NONE" |
| RainViewer | index https://api.rainviewer.com/public/weather-maps.json; tiles `{host}{path}/{size}/{z}/{x}/{y}/{color}/{options}.png` | XYZ, 256 or 512 px; "Maximum zoom level is 7" (https://www.rainviewer.com/api/weather-maps-api.html) | pre-coloured; 13 past frames and 0 nowcast frames in the index | "refreshed every 5 minutes" | global | see section 7 |

- **No keyless service here returns raw values as XYZ tiles.** Every tile or WMS response is pre-coloured.
- **An existing terminal radar app reverse-engineers colour into intensity.** termidar fetches IEM WMS and RainViewer tiles, then derives intensity from RGB with heuristics (https://raw.githubusercontent.com/N-Erickson/termidar/HEAD/internal/radar/client.go).
- **WMS can skip tiling entirely.** Because WMS accepts an arbitrary bounding box and pixel size, a host can request one PNG at exactly the viewport's dot or cell resolution. I did not fetch a PNG at that size.

## 3. Gridded scalar fields

| Source | One call gives a 2-D field? | Measured |
|---|---|---|
| NWS `/gridpoints/{wfo}/{x},{y}` | No. One 2.5 km cell per call. | One cell is 214 KB. A 40×20 grid would be 800 calls and about 171 MB, which is unrealistic. |
| NDFD WMS: https://digital.weather.gov/ndfd/wms and https://nowcoast.noaa.gov/geoserver/ndfd_temperature/wms (also https://mapservices.weather.noaa.gov/raster/rest/services/NDFD/NDFD_temp/MapServer) | Yes, but as pre-coloured images. Raw values via WCS or GeoTIFF are UNVERIFIED. | capabilities documents of 1.8 MB and 72 KB |
| NDFD GRIB2: https://tgftp.nws.noaa.gov/SL.us008001/ST.opnl/DF.gr2/DC.ndfd/AR.conus/VP.001-003/ds.temp.bin | Yes, whole CONUS | 48 MB per element; too heavy for a terminal app |
| **NOMADS GFS grib filter**: https://nomads.ncep.noaa.gov/gribfilter.php?ds=gfs_0p25 | **Yes. It subsets by variable, level and bounding box.** | The box lat 24–50°N, lon 125–66°W (roughly CONUS) with TMP 2 m plus UGRD and VGRD 10 m returned **103 KB**: 3 messages on a 237×105 regular lat/lon grid (template 3.0) using simple packing (5.0). The page asks: "please pause before resubmitting requests". Hard rate limits are UNVERIFIED. |
| NOMADS HRRR filter (`filter_hrrr_2d.pl`) | Yes | A 10°×6° box of TMP 2 m was 78 KB on a 300×231 grid. The grid is Lambert conformal (template 3.30), so the host must reproject. Packing is simple (5.0). |
| HRRR on AWS: https://noaa-hrrr-bdp-pds.s3.amazonaws.com/ | Yes, using the `.idx` file and an HTTP Range request | The full file is 158 MB. TMP 2 m alone is 1.2 MB and UGRD 10 m is 2.4 MB. Licence: "open to the public and can be used as desired"; no AWS account needed (https://registry.opendata.aws/noaa-hrrr-pds/). |
| Open-Meteo: https://open-meteo.com/en/docs | Partly. Comma-separated coordinates return a JSON list. There is no bounding-box parameter. | 100 points: 44 KB in 2.2 s. 500 points: 173 KB. 800 points by GET: HTTP 414. After a cumulative 600 locations within one minute the API answered 429 "Minutely API request limit exceeded". **So each location counts as one call.** A 40×20 grid (800 points) exceeds the 600-per-minute limit on its own. The 10,000-per-day limit allows about 12 such frames a day. |

Pure-Go GRIB2 decoders:

| Library | Grid templates | Packing templates | Licence / activity |
|---|---|---|---|
| github.com/nilsmagnus/grib | UNVERIFIED | complex packing with spatial differencing plus "Data0" (simple); no JPEG2000 or PNG | licence type and last activity UNVERIFIED |
| github.com/amsokol/go-grib2 | UNVERIFIED | simple packing only; "does not support jpeg, png and aec" | MIT, dormant |

NOMADS-filtered output uses simple packing, so both libraries can decode it.

**Which is realistic for a 40×20 temperature grid over a viewport:**
- **US regional and national views:** the NOMADS GFS filter. It is about 100 KB and one call per 1–3 hours. At 0.25° (~28 km) spacing, a 300 km wide viewport has only about 11 native columns, so the host must upsample.
- **Small or global viewports:** Open-Meteo at 200 points or fewer, cached.
- **Inside the host, not the library:** the HRRR filter, when detail matters and the host is willing to reproject.

## 4. Vector fields (wind)

- **Encoding.**
  - GRIB sources carry UGRD and VGRD in m/s.
  - NWS carries `windDirection` as `wmoUnit:degree_(angle)` and `windSpeed` as `km_h-1`.
  - Open-Meteo carries `wind_speed_10m` and `wind_direction_10m` in degrees.
- **Direction convention.** Direction means where the wind comes from: "(0=north,90=east,180=south,270=west) that the wind is coming **from**". `Ugeo = -Spd*sin(Dir)` and `Vgeo = -Spd*cos(Dir)` (https://www.eol.ucar.edu/content/wind-direction-quick-reference).
- **Terminal practice.** wego uses 8 arrows that point the way the wind is going: `{"↓","↙","←","↖","↑","↗","→","↘"}[((deg+22)%360)/45]` (https://github.com/schachmat/wego/blob/master/frontends/ascii-art-table.go). wttr.in's one-line output does the same; `?format=%w` returned "↖7mph".
- **Particle animation.** earth.nullschool animates GFS wind refreshed "every three hours" (https://earth.nullschool.net/about.html).
- **Legibility at one glyph per 4–8 cells (my inference, not cited):**
  - 8-way arrows coloured by speed work.
  - Barbs need more than one cell.
  - Particle or streamline rendering in braille would compete with the basemap lines for the same dots.
  - I recommend a decimated lattice of arrows, with optional animated braille streaks later.

## 5. Input-shape synthesis

| Shape | Contract | Sources that map onto it | Volume per frame | Verdict |
|---|---|---|---|---|
| **A. Features** | Point, LineString, Polygon and Multi* in WGS84 lon/lat, plus a style (stroke, fill or pattern, glyph, z-order), a label and an id | NWS alerts and zones, SPC, NHC, WFIGS, USGS, FIRMS | 10–500 features; raw 10²–10⁵ vertices; at most a few thousand after simplification | **v1-essential.** The library must clip and simplify to dot tolerance, or document that the host must. The zone sizes in section 1 make this unavoidable. |
| **B. Scalar grid** | Regular lon/lat axes (origin, dx, dy, nx, ny), row-major `[]float32`, NaN for missing, a colour ramp with class breaks, and nearest or bilinear resampling to cells | GFS and HRRR via the host, an Open-Meteo lattice, IEM raw dBZ, radar decoded by the host | 800 to 25,000 floats (3–100 KB) | **v1-essential.** It renders as the cell background. |
| **C. Vector grid** | Two B grids (u and v in m/s), plus a helper that converts from speed and "from" direction | GFS and HRRR UGRD/VGRD, Open-Meteo | 2 × B | v1.1. It is cheap once B exists. |
| **D. Georeferenced image** | An `image.Image` plus bounds in EPSG:3857 or lon/lat, in direct-colour mode, with an optional palette-to-value table | WMS GetMap from opengeo, nowCOAST or IEM; stitched XYZ tiles | one PNG of 2–20 KB | v1 if radar is a launch feature. It is the only route to MRMS data. |
| **E. XYZ provider** | `func(z,x,y) → image.Image` | IEM tiles, RainViewer | 4–12 tiles of about 5 KB each | Can wait. The host can stitch tiles into D. |

**Strongest counter-argument.** D alone could replace B and C: the host colours everything and the library only blits. The cost is that the library could no longer adapt ramps to 16-, 256- or true-colour terminals, draw legends, or keep the field's contrast subordinate to the basemap. termidar's RGB heuristics show how lossy colour-to-value recovery is.

**Counter-argument for E.** Tiles align exactly with the basemap pyramid, and the basemap's tile cache could be reused for them.

## 6. Low-resolution cartography prior art

- **linecast** (https://github.com/ashuttl/linecast) "draws streets in braille over water, land, parks, and buildings in solid color" and "animates the recent observations...over a braille map, with US weather warnings drawn on top". It uses IEM, RainViewer and LibreWXR for radar. This is the closest precedent to the technique proposed here.
- **notcurses** (https://notcurses.com/notcurses_visual.3.html): "there are only ever two colors available to us in a given cell"; "NCBLIT_BRAILLE doesn't tend to work out very well for images, but (depending on the font) can be very good for plots".
- **termidar** maps radar onto intensity glyphs `" ·∘○●◉◆◈▰▱█"` with colour, and draws no basemap lines in the same cell (https://raw.githubusercontent.com/N-Erickson/termidar/HEAD/internal/ui/model.go).
- **wttr.in v3 and termrain** avoid text cells for maps altogether. wttr.in offers "PNG… Sixel… IIP" (https://github.com/chubin/wttr.in). termrain uses the Kitty graphics protocol (https://github.com/iorinu/termrain).
- **MapSCII** (https://github.com/rastapasta/mapscii) has braille and block modes and converts RGB "to closest xterm-256 color code".

**Conclusion (my inference).** Half-blocks and shade blocks use up the cell's glyph, so they cannot share a cell with braille. The technique that layers cleanly is field as background colour with basemap as foreground braille, as linecast does. This gives one field sample per cell (80×24 = 1,920 samples), so B and D need only cell resolution. The ramp should have 5–7 muted steps, and the foreground colour should be chosen per cell for contrast. Polygons from shape A can be drawn as a braille outline plus an optional background tint.

## 7. Terms and risk notes

- **Open-Meteo**
  - Licence: data is "under the terms of the CC-BY 4.0 licence", so attribution is required.
  - Use: "The free API is for non-commercial use, rate-limited to 10,000 calls/day". Commercial use includes "websites or apps that have subscriptions or display advertisements" (https://open-meteo.com/en/terms, https://open-meteo.com/en/pricing).
  - Per-location counting makes lattice requests expensive.
  - Whether POST lifts the GET length ceiling is UNVERIFIED.
- **RainViewer**
  - Use and attribution: "The API is free for personal or educational use only." "We kindly ask you to mention the RainViewer API as a source" (https://www.rainviewer.com/api.html).
  - Changes from 1 January 2026 (https://www.rainviewer.com/api/transition-faq.html):
    - nowcast discontinued;
    - "All color schemes except Universal Blue discontinued";
    - zoom 7 maximum;
    - "100 requests/IP/minute".
  - It should not be a default radar source.
- **NWS** requires a User-Agent header. Its rate limit is undisclosed; throttled requests may be retried "typically within 5 seconds".
- **IEM** is public domain and asks for attribution as a courtesy. It is a university-run service with no SLA.
- **NOMADS** asks clients to pause between requests. Its blocking thresholds are UNVERIFIED.
- **Rulings for the project owner:**
  1. Does the library display an attribution line for overlays, and therefore does the contract need an `Attribution` string per overlay?
  2. Is Open-Meteo's non-commercial limit acceptable for Watchpost's distribution model?
  3. Does simplification of 10⁴-vertex zone polygons live in the library or in the host?
  4. Is reprojection of HRRR and NDFD Lambert grids explicitly out of scope, so that the library accepts regular lon/lat grids only?
  5. Is palette decoding of radar images in scope for the library (D's palette-to-value mode) or the host's job?
