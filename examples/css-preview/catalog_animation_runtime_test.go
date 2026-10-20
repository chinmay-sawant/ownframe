package main

import "testing"

func TestCatalogAnimationPropertiesReachOwnFrameFrames(t *testing.T) {
	cases := append(append([]animationRuntimeCase{}, standardAnimationRuntimeCases...), webkitAnimationRuntimeCases...)
	for _, c := range cases {
		t.Run(c.property, func(t *testing.T) {
			before, beforeActive := sampleAnimation(t, c, c.before)
			after, afterActive := sampleAnimation(t, c, c.after)
			if c.compareActive && beforeActive == afterActive {
				t.Fatalf("active = %v for both values", afterActive)
			}
			if !c.compareActive && before == after {
				t.Fatalf("sample = %+v for both values", after)
			}
		})
	}
}
