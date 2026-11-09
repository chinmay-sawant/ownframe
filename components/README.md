# Components

Reusable controls live here as named sets. Callers import a set and append its Stylesheet.

```text
components/
├── chadcn/         controls
├── chartjs/        Chart.js
├── shadcnchart/    shadcn charts
├── recharts/       Recharts
├── tremor/         Tremor
├── echarts/        ECharts
├── apexcharts/     ApexCharts
└── nivo/           Nivo
```

`chartjs`, `shadcnchart`, `recharts`, `tremor`, `echarts`, `apexcharts`, and `nivo` are HTML and CSS chart packages. Each set has constructors and a Stylesheet. The UI builder has one sidebar entry per set.

ChadCN is the first control set. See [chadcn/README.md](chadcn/README.md).
