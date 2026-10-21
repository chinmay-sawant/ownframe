package main

// rowForButton returns the catalog row with the given button id.
func (d *demo) rowForButton(id string) (catalogRow, bool) {
	for _, row := range d.v.Catalog {
		if row.Button == id {
			return row, true
		}
	}
	return catalogRow{}, false
}

// rowForName returns the catalog row with the given property name.
func (d *demo) rowForName(name string) (catalogRow, bool) {
	for _, row := range d.v.Catalog {
		if row.Name == name {
			return row, true
		}
	}
	return catalogRow{}, false
}
