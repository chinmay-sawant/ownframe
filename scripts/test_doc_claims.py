from pathlib import Path
import tempfile
import unittest
from doc_claims import check_doc_pins, check_perf_commands, check_versions, check_current_options


class DocClaimTests(unittest.TestCase):
    def test_module_requirements_must_agree(self):
        source = 'go 1.26.10\nrequire github.com/chinmay-sawant/blinkless v0.0.3\n'
        self.assertEqual(check_versions([source, source]), ('1.26.10', 'v0.0.3'))
        for changed in (source.replace('1.26.10', '1.26.9'), source.replace('v0.0.3', 'v0.0.2'), ''):
            with self.subTest(changed=changed), self.assertRaises(ValueError):
                check_versions([source, changed])

    def test_stale_doc_pins_fail(self):
        check_doc_pins('Use Go 1.26.11 and blinkless v0.1.2.', 'AGENTS.md', '1.26.11', 'v0.1.2')
        with self.assertRaisesRegex(ValueError, 'stale blinkless pin'):
            check_doc_pins('`github.com/chinmay-sawant/blinkless` `v0.0.4-0.2026`', 'AGENTS.md', '1.26.11', 'v0.1.2')
        with self.assertRaisesRegex(ValueError, 'stale Go version'):
            check_doc_pins('Use Go 1.26.10 or newer.', 'CONTRIBUTING.md', '1.26.11', 'v0.1.2')

    def test_snippet_scratch_directory_cleans_up_on_failure(self):
        with tempfile.TemporaryDirectory() as root:
            root_path = Path(root)
            with self.assertRaises(RuntimeError):
                with tempfile.TemporaryDirectory(prefix='.doc-snippets-', dir=root_path) as scratch:
                    (Path(scratch) / 'snippet.go').write_text('package snippet\n')
                    raise RuntimeError('simulated vet failure')
            self.assertEqual(list(root_path.glob('.doc-snippets-*')), [])

    def test_invalid_documented_commands_fail(self):
        flags = {'scripts/perf-idle.sh': {'--example', '--seconds', '--web'}}
        check_perf_commands('sh scripts/perf-idle.sh --example examples/login --seconds 10 --web', flags)
        for command in ('examples/login', '--typo', '--seconds', '--example --seconds 10'):
            with self.subTest(command=command), self.assertRaises(ValueError):
                check_perf_commands('sh scripts/perf-idle.sh ' + command, flags)


class CurrentOptionTests(unittest.TestCase):
    def test_deprecated_serve_options_fail(self):
        for field in ('Debug', 'Perf'):
            with self.subTest(field=field), self.assertRaises(ValueError):
                check_current_options('ServeOptions{' + field + ': true}')
        check_current_options('ServeOptions{InsecureAllowPprof: true}')
