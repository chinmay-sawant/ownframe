package textrun

// FaceFor exposes faceFor to the external test package, which can import
// internal/page without an import cycle.
var FaceFor = faceFor
