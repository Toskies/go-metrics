import random
import unittest

from scripts.load_local_stack import build_request_plan, normalize_base_url


class LoadLocalStackTests(unittest.TestCase):
    def test_normalize_base_url_removes_trailing_slash(self):
        self.assertEqual(normalize_base_url("http://localhost:8080/"), "http://localhost:8080")

    def test_build_request_plan_uses_configured_bounds_and_known_paths(self):
        rng = random.Random(7)

        count, urls = build_request_plan("http://localhost:8080", 10, 40, rng)

        self.assertGreaterEqual(count, 10)
        self.assertLessEqual(count, 40)
        self.assertEqual(len(urls), count)
        self.assertTrue(urls)
        for url in urls:
            self.assertIn(url, {
                "http://localhost:8080/work",
                "http://localhost:8080/checkout",
            })

    def test_build_request_plan_honors_exact_fixed_rate(self):
        rng = random.Random(1)

        count, urls = build_request_plan("http://localhost:8080", 12, 12, rng)

        self.assertEqual(count, 12)
        self.assertEqual(len(urls), 12)


if __name__ == "__main__":
    unittest.main()
