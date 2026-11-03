// Package chadcn provides the ChadCN component set for OwnFrame.
//
// Components are HTML and CSS built through the ui builder. They return fresh
// mutable trees, keep state in the host application, and take their data
// through options structs. Paint and interactive states live in Stylesheet,
// which an application appends to its theme:
//
//	Theme: theme.Dark + chadcn.Stylesheet
//
// Openable components take an Open field. The same constructor renders the
// closed trigger and the open surface, so no separate open-state type exists.
package chadcn
