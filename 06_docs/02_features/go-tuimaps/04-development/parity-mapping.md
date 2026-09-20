# Parity mapping — every first-release row, the work package that owns it, and the test that proves it

| Field | Value |
|---|---|
| Phase | PLAN |
| Date | 2026-09-19 |
| Why this exists | M3's guard is a checked-in test for every counted row. The plan first tested parity last, by one catch-all task; the PLAN red-team found seven of twelve sampled rows with no task at all (PL-BZ-2). This mapping is written **first** (plan task 00.13); each test below is written before the code it checks, inside the work package named; task 14.1 fails if a row is missing here or a named test does not exist. |
| How the owner was chosen | By the upstream source file each row was read from, then corrected by hand where a ruling moved the behaviour (P-05a, P-08, P-36). A work package that finds a row belongs elsewhere moves it here, in the same commit. **Red-team round 2 found that choosing by upstream's source file had mis-owned four rows** — colour and width are resolved by the style at draw time, not by the decoder (P-44, P-45); a failed fetch is the tile pipeline's (P-51); marker ids are the public package's (P-61) — and that six rows had an owner with no task to build them; the plan gained tasks 09.29, 09.30, 12.25 and 12.26. Where a plan task already named a test for a row's behaviour, the task now carries this mapping's name for it. |
| Count | **62 rows** — the rows marked v0.1.0 in [the parity matrix](../02-analysis/parity-matrix.md), which remains the authority for what each row means. |

| Owner | Rows |
|---|---|
| WP-01 project | 4 |
| WP-03 mvt decoder | 5 |
| WP-04 assets | 1 |
| WP-06 tiles | 4 |
| WP-08 style and colour | 6 |
| WP-09 render | 31 |
| WP-12 public package | 7 |
| WP-13 app | 4 |

| Row | Behaviour | Disposition | Owner | Test |
|---|---|---|---|---|
| P-01 | Pixel grid | Match | WP-09 render | `TestParityP01_PixelGrid` |
| P-02 | Dot bitmask | Match | WP-09 render | `TestParityP02_DotBitmask` |
| P-03a | Empty cell | Match | WP-09 render | `TestParityP03a_EmptyCell` |
| P-05a | Colour depth | Extended | WP-08 style and colour | `TestParityP05a_ColourDepth` |
| P-06 | SGR forms | Extended | WP-09 render | `TestParityP06_SgrForms` |
| P-07 | Row self-containment | Extended | WP-09 render | `TestParityP07_RowSelfContainment` |
| P-08 | Cell colour vote | Extended | WP-09 render | `TestParityP08_CellColourVote` |
| P-09 | Forced pixels | Match | WP-09 render | `TestParityP09_ForcedPixels` |
| P-10 | Text cells | Match | WP-09 render | `TestParityP10_TextCells` |
| P-11 | Wide chars | Match | WP-09 render | `TestParityP11_WideChars` |
| P-12 | Text placement | Fix | WP-09 render | `TestParityP12_TextPlacement` |
| P-13 | Thin line | Match | WP-09 render | `TestParityP13_ThinLine` |
| P-14 | Thick line | Match | WP-09 render | `TestParityP14_ThickLine` |
| P-15 | Polygon fill | Match | WP-09 render | `TestParityP15_PolygonFill` |
| P-16 | Degenerate rings | Match | WP-09 render | `TestParityP16_DegenerateRings` |
| P-17 | Tile zoom | Match | WP-01 project | `TestParityP17_TileZoom` |
| P-18 | Tile pixel size | Match | WP-01 project | `TestParityP18_TilePixelSize` |
| P-19 | Visible tiles | Replicate | WP-09 render | `TestParityP19_VisibleTiles` |
| P-20 | Projection | Match | WP-01 project | `TestParityP20_Projection` |
| P-21 | Normalize | Replicate | WP-01 project | `TestParityP21_Normalize` |
| P-22 | Ocean outside world | Fix | WP-09 render | `TestParityP22_OceanOutsideWorld` |
| P-23 | Draw order, z≥2 | Extended | WP-09 render | `TestParityP23_DrawOrderZ2` |
| P-24 | Draw order, z<2 | Extended | WP-09 render | `TestParityP24_DrawOrderBelowZ2` |
| P-25 | Label deferral | Match | WP-09 render | `TestParityP25_LabelDeferral` |
| P-26 | Sort key | Match | WP-03 mvt decoder | `TestParityP26_SortKey` |
| P-27 | Zoom gate | Match | WP-08 style | `TestParityP27_ZoomGate` |
| P-28 | Feature cull | Match | WP-09 render | `TestParityP28_FeatureCull` |
| P-29 | Scaling | Match | WP-09 render | `TestParityP29_Scaling` |
| P-30 | Line clip pad | Fix | WP-09 render | `TestParityP30_LineClipPad` |
| P-31 | Simplify | Match | WP-09 render | `TestParityP31_Simplify` |
| P-32 | Label anchor | Fix | WP-09 render | `TestParityP32_LabelAnchor` |
| P-33 | Label bounds | Fix | WP-09 render | `TestParityP33_LabelBounds` |
| P-34 | Collision | Match | WP-09 render | `TestParityP34_Collision` |
| P-35 | POI glyph | Match | WP-09 render | `TestParityP35_PoiGlyph` |
| P-36 | Label language | Match | WP-03 mvt decoder | `TestParityP36_LabelLanguage` |
| P-37 | Gzip sniff | Match | WP-03 mvt decoder | `TestParityP37_GzipSniff` |
| P-38 | MVT decode | Match | WP-03 mvt decoder | `TestParityP38_MvtDecode` |
| P-39 | Ring grouping | Match | WP-03 mvt decoder | `TestParityP39_RingGrouping` |
| P-41 | Style match | Match | WP-08 style and colour | `TestParityP41_StyleMatch` |
| P-42 | Filter ops | Match | WP-08 style and colour | `TestParityP42_FilterOps` |
| P-43 | Constants and `ref` | Match | WP-08 style and colour | `TestParityP43_ConstantsAndRef` |
| P-44 | Colour pick | Fix | WP-08 style and colour | `TestParityP44_ColourPick` |
| P-45 | Line width | Fix | WP-08 style and colour | `TestParityP45_LineWidth` |
| P-47 | Embedded tiles | Extended | WP-04 assets | `TestParityP47_EmbeddedTiles` |
| P-48 | Disk cache | Fix | WP-06 tiles | `TestParityP48_DiskCache` |
| P-49 | URL modes | Extended | WP-06 tiles | `TestParityP49_UrlModes` |
| P-50 | HTTP | Fix | WP-06 tiles | `TestParityP50_Http` |
| P-51 | Fetch failure | Extended | WP-06 tiles | `TestParityP51_FetchFailure` |
| P-52 | Config defaults | Match | WP-12 public package | `TestParityP52_ConfigDefaults` |
| P-53 | Size from terminal | Fix | WP-12 public package | `TestParityP53_SizeFromTerminal` |
| P-54 | Min zoom | Match | WP-12 public package | `TestParityP54_MinZoom` |
| P-55 | zoom_by / initial zoom | Match | WP-12 public package | `TestParityP55_ZoomByInitialZoom` |
| P-56 | fit_world | Match | WP-12 public package | `TestParityP56_FitWorld` |
| P-57 | Footer | Match | WP-12 public package | `TestParityP57_Footer` |
| P-58 | Marker shapes | Match | WP-09 render | `TestParityP58_MarkerShapes` |
| P-59a | Marker animation | Match | WP-09 render | `TestParityP59a_MarkerAnimation` |
| P-60 | Marker cull and label | Match | WP-09 render | `TestParityP60_MarkerCullAndLabel` |
| P-61 | Marker id | Match | WP-12 public package | `TestParityP61_MarkerId` |
| P-68a | Keys | Match | WP-13 app | `TestParityP68a_Keys` |
| P-69 | Pan step | Match | WP-13 app | `TestParityP69_PanStep` |
| P-71 | Loop | Fix | WP-13 app | `TestParityP71_Loop` |
| P-72a | Chrome | Match | WP-13 app | `TestParityP72a_Chrome` |
