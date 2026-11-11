#!/usr/bin/env python3
"""Require pixel equality between native and Chrome sticky fixture captures."""
import argparse
import json
from pathlib import Path
from PIL import Image, ImageChops


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--native', type=Path, required=True)
    parser.add_argument('--reference', type=Path, required=True)
    parser.add_argument('--out', type=Path, required=True)
    args = parser.parse_args()
    results = []
    for step in (0, 1, 2, 3, 4, 5, 6, 9):
        with Image.open(args.native / f'sticky-{step}.png') as native_file:
            native = native_file.convert('RGBA')
        with Image.open(args.reference / f'chrome-sticky-{step}.png') as reference_file:
            reference = reference_file.convert('RGBA')
        if native.size != (200, 100) or reference.size != native.size:
            raise ValueError(f'step {step}: incompatible viewport captures')
        difference = ImageChops.difference(native, reference)
        if any(channel.getbbox() for channel in difference.split()):
            raise ValueError(f'step {step}: native framebuffer differs from Chrome')
        results.append({'step': step, 'size': list(native.size), 'different_pixels': 0})
    report = {'status': 'passed', 'comparison': 'RGBA pixel equality', 'steps': results,
              'scope': 'Unzoomed solid-color fixture; native zoom has separate pixel and hit assertions'}
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps(report, indent=2) + '\n')
    print(f'PASS: all {len(results)} native sticky frames match Chrome pixel for pixel')


if __name__ == '__main__':
    main()
