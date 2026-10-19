package main

func catalogAction(name string) (string, string) {
	return catalogSampleKey(name), catalogBehavior(name)
}
