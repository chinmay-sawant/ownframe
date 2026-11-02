# Components

Reusable controls live here as named sets:

```text
components/
├── chadcn/     the first set
└── <set>/      later sets
```

A set owns its constructors, stylesheet, palettes, tests, and golden files.
Applications import one set, append its stylesheet to the theme, and keep
state in the host.

ChadCN is the first set. See [chadcn/README.md](chadcn/README.md).
