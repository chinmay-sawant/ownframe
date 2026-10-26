# Why ownframe exists

ownframe started as an experiment around a question: **How far can we push Go for native application development?**

I wanted to explore building desktop and Android interfaces with HTML and CSS while keeping application logic and rendering in Go, without bundling a full browser or JavaScript runtime. [Blinkless](https://github.com/chinmay-sawant/blinkless) handles HTML and CSS layout; ownframe connects that output to rendering, input, and Go handlers through Ebitengine.

## Learning by building

The project is a way to test ideas against actual use cases, understand the constraints of a lightweight UI engine, and discover where a Go-first approach works well and where it falls short. Supported HTML and CSS are intentionally a subset, not full browser compatibility.

AI-assisted development is part of the experiment too. Generated code is reviewed and tested rather than treated as automatically correct.

## An open-ended experiment

This is not a claim that existing frameworks are obsolete or that ownframe is the right tool for every app. It is an opportunity to question assumptions, explore trade-offs, and push my understanding of what Go can do. The direction can change as experiments reveal new possibilities and limitations.

For current capabilities and limitations, see the [feature list](features.md) and [platform support](platforms.md).
