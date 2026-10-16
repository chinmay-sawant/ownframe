#!/usr/bin/env python3
"""Convert a video to 60fps MP4, then make an animated WebP from it.

Pipeline (same directory as the input by default):
  eyes2.mp4 -> eyes2-60fps.mp4 -> eyes2-60fps.webp

Requires an ffmpeg binary: system ffmpeg, or the bundled one from
imageio-ffmpeg (pip install imageio-ffmpeg).

Usage:
  python3 scripts/convert_to_60fps.py
  python3 scripts/convert_to_60fps.py --input assets/screenshots/android/eyes/eyes2.mp4
  python3 scripts/convert_to_60fps.py --fps 60 --crf 20 --webp-quality 80
"""
from __future__ import annotations

import argparse
import shutil
import subprocess
import sys
from pathlib import Path


def resolve_ffmpeg() -> str:
    found = shutil.which("ffmpeg")
    if found:
        return found
    try:
        import imageio_ffmpeg  # type: ignore

        exe = imageio_ffmpeg.get_ffmpeg_exe()
        if Path(exe).is_file():
            return exe
    except Exception:
        pass
    print(
        "ffmpeg not found: install it (apt install ffmpeg) "
        "or pip install imageio-ffmpeg",
        file=sys.stderr,
    )
    sys.exit(2)


def run(cmd: list[str]) -> None:
    print("+ " + " ".join(cmd))
    subprocess.run(cmd, check=True)


def convert_mp4(ffmpeg: str, src: Path, dest: Path, fps: int, crf: int) -> None:
    run(
        [
            ffmpeg,
            "-y",
            "-i",
            str(src),
            "-vf",
            f"fps={fps}",
            "-c:v",
            "libx264",
            "-pix_fmt",
            "yuv420p",
            "-crf",
            str(crf),
            "-preset",
            "medium",
            "-c:a",
            "aac",
            "-movflags",
            "+faststart",
            str(dest),
        ]
    )


def convert_webp(ffmpeg: str, src: Path, dest: Path, fps: int, quality: int) -> None:
    run(
        [
            ffmpeg,
            "-y",
            "-i",
            str(src),
            "-vf",
            f"fps={fps}",
            "-c:v",
            "libwebp",
            "-quality",
            str(quality),
            "-compression_level",
            "6",
            "-loop",
            "0",
            "-an",
            str(dest),
        ]
    )


def format_bytes(n: int) -> str:
    if n >= 1_000_000:
        return f"{n / 1_000_000:.2f} MB"
    if n >= 1_000:
        return f"{n / 1_000:.1f} KB"
    return f"{n} B"


def main() -> int:
    root = Path(__file__).resolve().parents[1]
    default_input = root / "assets" / "screenshots" / "android" / "eyes" / "eyes2.mp4"

    parser = argparse.ArgumentParser(
        description="Convert a video to 60fps MP4, then animated WebP beside it."
    )
    parser.add_argument("--input", type=Path, default=default_input)
    parser.add_argument("--fps", type=int, default=60)
    parser.add_argument("--crf", type=int, default=20)
    parser.add_argument("--webp-quality", type=int, default=80)
    parser.add_argument("--skip-mp4", action="store_true")
    parser.add_argument("--skip-webp", action="store_true")
    args = parser.parse_args()

    src = args.input if args.input.is_absolute() else (root / args.input)
    if not src.is_file():
        print(f"input not found: {src}", file=sys.stderr)
        return 1
    if args.fps <= 0:
        print("--fps must be positive", file=sys.stderr)
        return 1

    mp4_dest = src.parent / f"{src.stem}-{args.fps}fps.mp4"
    webp_dest = src.parent / f"{src.stem}-{args.fps}fps.webp"
    ffmpeg = resolve_ffmpeg()

    if not args.skip_mp4:
        convert_mp4(ffmpeg, src, mp4_dest, args.fps, args.crf)
        print(f"wrote {mp4_dest.name} ({format_bytes(mp4_dest.stat().st_size)})")
    else:
        if not mp4_dest.is_file():
            print(f"--skip-mp4 given but {mp4_dest.name} is missing", file=sys.stderr)
            return 1

    if not args.skip_webp:
        convert_webp(ffmpeg, mp4_dest, webp_dest, args.fps, args.webp_quality)
        print(f"wrote {webp_dest.name} ({format_bytes(webp_dest.stat().st_size)})")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
