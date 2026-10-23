# Deferred Plans & Architectural Analysis

> **Directory:** `plans/deferred/`  
> **Status:** Active / Open Reference  

---

## Overview

This directory houses long-term architectural evaluations, comparative benchmarks, and capabilities deferred from the immediate 1.0 scope:

1. **[Hello World Framework Benchmark & Resource Comparison](hello-world-framework-comparison.md)**
   An exhaustive technical comparison of a baseline "Hello World" application across **ownframe**, **GPUI** (Zed), **Wails v2/v3**, **Tauri v2**, **Electron**, **Fyne**, **Flutter Desktop**, and **Ruby on Rails**. Compares:
   * **Resident Set Size (RSS)** and memory footprint.
   * **Startup latency (Time to First Frame)**.
   * **Binary / Distribution footprint** (self-contained vs. host WebView dependencies).
   * **Idle CPU pacing and battery efficiency**.
   * **Process architecture & memory boundaries** (in-memory Go struct pointers vs. multi-process JSON/IPC bridges).

2. **Roadmap & Phase Checklists:**
   The active phase-wise checklists for closing key platform gaps (Accessibility, Complex IME, Layout Engine completeness, and Codebase Cohesion) are tracked in **[plans/diffed/](../diffed/README.md)**.
