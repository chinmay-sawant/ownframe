import unittest
from doc_claims import check_perf_commands, check_versions


class DocClaimTests(unittest.TestCase):
    def test_module_requirements_must_agree(self):
        source = 'go 1.26.10\nrequire github.com/chinmay-sawant/blinkless v0.0.3\n'
        self.assertEqual(check_versions([source, source]), ('1.26.10', 'v0.0.3'))
        for changed in (source.replace('1.26.10', '1.26.9'), source.replace('v0.0.3', 'v0.0.2'), ''):
            with self.subTest(changed=changed), self.assertRaises(ValueError):
                check_versions([source, changed])

    def test_invalid_documented_commands_fail(self):
        flags = {'scripts/perf-idle.sh': {'--example', '--seconds', '--web'}}
        check_perf_commands('sh scripts/perf-idle.sh --example examples/login --seconds 10 --web', flags)
        for command in ('examples/login', '--typo', '--seconds', '--example --seconds 10'):
            with self.subTest(command=command), self.assertRaises(ValueError):
                check_perf_commands('sh scripts/perf-idle.sh ' + command, flags)
