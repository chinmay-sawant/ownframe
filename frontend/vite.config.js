import { writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';

export default defineConfig(({ command }) => ({
  base: '/ownframe/',
  plugins: command === 'build' ? [{
    name: 'go-module-boundary',
    closeBundle() {
      const path = fileURLToPath(new URL('../docs/go.mod', import.meta.url));
      writeFileSync(path, 'module github.com/chinmay-sawant/ownframe/docs\n\ngo 1.26.4\n');
    },
  }] : [],
  build: {
    outDir: fileURLToPath(new URL('../docs', import.meta.url)),
    emptyOutDir: true,
  },
}));
