package main

import "math"

// vec3 represents a 3D coordinate or direction vector in gravitational radii (rs).
type vec3 [3]float64

func (a vec3) add(b vec3) vec3    { return vec3{a[0] + b[0], a[1] + b[1], a[2] + b[2]} }
func (a vec3) mul(k float64) vec3 { return vec3{a[0] * k, a[1] * k, a[2] * k} }
func (a vec3) dot(b vec3) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func (a vec3) norm() vec3 {
	d := math.Sqrt(a.dot(a))
	if d == 0 {
		return vec3{}
	}
	return a.mul(1 / d)
}
func (a vec3) f32() [3]float32 {
	return [3]float32{float32(a[0]), float32(a[1]), float32(a[2])}
}
func (a vec3) cross(b vec3) vec3 {
	return vec3{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

// focal is the camera's focal length, in heights of the view.
const focal = 1.5

// The camera bounds: distance in Schwarzschild radii (rs), and pitch limits.
const (
	nearDist = 8.0
	restDist = 22.0
	farDist  = 36.0
	maxPitch = 1.4

	orbitSpeed = 0.025 // radians per second
	settleRate = 1.2   // decay damping factor
)

// camera orbits the black hole in spherical coordinates (yaw, pitch, distance)
// with rotational velocity (spinYaw, spinPitch) supporting drag inertia.
type camera struct {
	yaw, pitch, dist   float64
	spinYaw, spinPitch float64
}

func newCamera() camera {
	return camera{
		yaw:     0.6,
		pitch:   0.1,
		dist:    restDist,
		spinYaw: orbitSpeed,
	}
}

// axes returns the camera eye position and the orthonormal right, up, and forward basis vectors.
func (c *camera) axes() (eye, right, up, forward vec3) {
	cp := math.Cos(c.pitch)
	eye = vec3{
		cp * math.Sin(c.yaw),
		math.Sin(c.pitch),
		cp * math.Cos(c.yaw),
	}.mul(c.dist)

	forward = eye.mul(-1).norm()
	right = forward.cross(vec3{0, 1, 0}).norm()
	up = right.cross(forward)
	return
}

// step advances camera motion by dt seconds and damps momentum toward idle orbit.
func (c *camera) step(dt float64, dragging bool, reduceMotion bool) {
	if dragging {
		return
	}
	goal := orbitSpeed
	if reduceMotion {
		goal = 0
	}
	k := 1 - math.Exp(-dt*settleRate)
	c.spinYaw += (goal - c.spinYaw) * k
	c.spinPitch -= c.spinPitch * min(1, 2*k)
	c.yaw += c.spinYaw * dt
	c.pitch = clamp(c.pitch+c.spinPitch*dt, -maxPitch, maxPitch)
}

// spinning reports whether the camera still has active residual angular momentum.
func (c *camera) spinning() bool {
	return math.Abs(c.spinYaw) > 1e-3 || math.Abs(c.spinPitch) > 1e-3
}

func clamp(x, lo, hi float64) float64 {
	return max(lo, min(hi, x))
}
