# MyGo-inventory dashboard

Native MyGo port of the `builder/inventory` Northwind Goods dashboard:
top bar (brand, nav, search with SKU and PO results, alerts popover,
account popover), page header (period segment, category filter,
`+ New PO` modal dialog), five KPI cells, the stock health table with
working status pills, the bin utilization grid, inbound purchase orders,
value by category, and the product heat map.

Rows of the stock table are the 8 separate `sampleSKUs` products; the
heat map crosses them with aisles A-J, scaling each aisle's average bin
fill by the product's own stock health on the inventory bin scale.
Hovering a bin cell shows its data in the note line above the grid and
in the cell tooltip, the way the builder patches `#bin-detail`.

MyGo `v0.7.4` needs Go `>= 1.27.3` while the repo pins Go `1.26.10`, so this
is an isolated module like `examples/compare/mygo`: it stays out of
`make build`, `make test`, and `make lint` until the repo moves Go.

```sh
cd examples/mygo-inventory && GOWORK=off go run .
cd examples/mygo-inventory && GOWORK=off go test .
cd examples/mygo-inventory && GOWORK=off go run . -bench
```
