"""Exercise graphics gate session selection without starting a graphics server."""

import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

SCRIPT = Path(__file__).resolve().with_name("check-graphics.sh")
REQUIRED_TESTS = {
    "TestPartialFrameLeavesTheScreen",
    "TestWindowRepaintPixels",
    "TestViewportScrollPixels",
    "TestViewportAnimatedReturnPixels",
    "TestViewportDirectAnimationReturnPixels",
    "TestWindowIdleDeadline",
    "TestGPUVisibleParity",
    "TestDrawPaintsVerticalThumb",
}


class GraphicsGateTests(unittest.TestCase):
    def run_gate(self, session, *, display=":test", action="pass", software=None, status=0):
        with tempfile.TemporaryDirectory() as directory:
            fixture = Path(directory)
            log = fixture / "calls.jsonl"
            go = fixture / "go"
            go.write_text(
                "#!/usr/bin/env python3\n"
                "import json, os, sys\n"
                "name = sys.argv[sys.argv.index('-run') + 1].strip('^$')\n"
                "with open(os.environ['GRAPHICS_TEST_LOG'], 'a') as out:\n"
                "    out.write(json.dumps({'test': name, 'software': "
                "os.environ.get('LIBGL_ALWAYS_SOFTWARE')}) + '\\n')\n"
                "print(json.dumps({'Test': name, 'Action': "
                "os.environ['GRAPHICS_TEST_ACTION']}))\n"
                "sys.exit(int(os.environ['GRAPHICS_TEST_STATUS']))\n"
            )
            go.chmod(0o755)
            runner = fixture / "xvfb-run"
            runner.write_text('#!/usr/bin/env bash\nshift 3\nexec "$@"\n')
            runner.chmod(0o755)
            for name in ("Xvfb", "xauth"):
                tool = fixture / name
                tool.write_text("#!/usr/bin/env bash\nexit 0\n")
                tool.chmod(0o755)
            env = os.environ.copy()
            env.update(
                PATH=str(fixture) + os.pathsep + env["PATH"],
                GRAPHICS_SESSION=session,
                GRAPHICS_TEST_LOG=str(log),
                GRAPHICS_TEST_ACTION=action,
                GRAPHICS_TEST_STATUS=str(status),
                DISPLAY=display,
            )
            env.pop("LIBGL_ALWAYS_SOFTWARE", None)
            if software is not None:
                env["LIBGL_ALWAYS_SOFTWARE"] = software
            result = subprocess.run(
                ["bash", str(SCRIPT)], env=env, text=True,
                stdout=subprocess.PIPE, stderr=subprocess.STDOUT, check=False,
            )
            calls = [json.loads(row) for row in log.read_text().splitlines()] if log.exists() else []
            return result, calls

    def test_desktop_preserves_driver_selection(self):
        for software in (None, "0", "1"):
            with self.subTest(software=software):
                result, calls = self.run_gate("desktop", software=software)
                self.assertEqual(result.returncode, 0, result.stdout)
                self.assertEqual(len(calls), len(REQUIRED_TESTS))
                self.assertTrue(all(call["software"] == software for call in calls))
                self.assertEqual({call["test"] for call in calls}, REQUIRED_TESTS)

    def test_xvfb_selects_software_rendering(self):
        result, calls = self.run_gate("xvfb")
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertEqual({call["test"] for call in calls}, REQUIRED_TESTS)
        self.assertTrue(all(call["software"] == "1" for call in calls))

    def test_missing_display_and_invalid_session_fail(self):
        for session, display, message in (
            ("desktop", "", "requires DISPLAY"),
            ("typo", ":test", "unknown GRAPHICS_SESSION"),
        ):
            with self.subTest(session=session):
                result, calls = self.run_gate(session, display=display)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(message, result.stdout)
                self.assertEqual(calls, [])

    def test_skipped_pixel_test_fails_gate(self):
        result, calls = self.run_gate("desktop", action="skip")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("required test did not pass", result.stdout)
        self.assertEqual(len(calls), 1)

    def test_go_failure_cannot_be_hidden_by_pass_event(self):
        result, calls = self.run_gate("desktop", status=1)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(len(calls), 1)


if __name__ == "__main__":
    unittest.main()
