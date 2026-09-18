import json
from pathlib import Path
import re
import unittest


REPO_ROOT = Path(__file__).resolve().parents[1]
RECORD = REPO_ROOT / "acceptance" / "v0.1.0.dev12-pr18-rendered-machine.json"


class RenderedMachineEvidenceRecordTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.record = json.loads(RECORD.read_text(encoding="utf-8"))

    def test_record_is_exact_source_bound(self):
        source = self.record["source"]
        self.assertEqual(source["repository"], "GoreeCloud/goreecloud-search")
        self.assertEqual(source["pullRequest"], 18)
        self.assertEqual(
            source["revision"],
            "7930107d7a572478d852641e74d48e09d0129995",
        )
        self.assertEqual(source["version"], "0.1.0.dev12")
        self.assertRegex(source["revision"], r"^[0-9a-f]{40}$")
        for blob in source["blobs"].values():
            self.assertRegex(blob, r"^[0-9a-f]{40}$")

    def test_glaze_boundary_stays_fail_closed(self):
        glaze = self.record["glazeUi"]
        self.assertEqual(glaze["targetVersion"], "1.5.1")
        self.assertFalse(glaze["consumerAcceptanceEstablished"])
        self.assertFalse(glaze["productionEligible"])

    def test_render_harness_does_not_claim_live_runtime(self):
        renderer = self.record["renderer"]
        self.assertEqual(renderer["engine"], "Chromium")
        self.assertFalse(renderer["liveVps"])
        self.assertFalse(renderer["liveProvider"])
        self.assertIn("synthetic", renderer["fixture"].casefold())

    def test_all_screenshot_hashes_are_sha256(self):
        for state in self.record["states"].values():
            shot = state.get("screenshot")
            if shot is None:
                continue
            self.assertRegex(shot["sha256"], r"^[0-9a-f]{64}$")
            self.assertGreater(shot["bytes"], 0)

    def test_required_machine_states_are_verified(self):
        states = self.record["states"]
        self.assertTrue(states["desktopLight"]["noHorizontalOverflow"])
        self.assertEqual(states["desktopLight"]["formMethod"], "post")
        self.assertEqual(states["desktopLight"]["scripts"], 0)
        self.assertEqual(states["desktopLight"]["remoteResources"], 0)
        self.assertTrue(states["mobileDarkReducedMotion"]["noHorizontalOverflow"])
        constrained = states["reducedTransparencyHighContrast"]
        self.assertEqual(constrained["backdropFilter"], "none")
        self.assertEqual(constrained["boxShadow"], "none")
        self.assertEqual(constrained["borderTopWidth"], "2px")
        self.assertEqual(states["errorState"]["role"], "alert")

    def test_accessibility_tree_is_not_misrepresented_as_at_acceptance(self):
        accessibility = self.record["accessibilityTree"]
        self.assertIn("main", accessibility["landmarks"])
        self.assertIn("search", accessibility["landmarks"])
        self.assertIn("not assistive-technology acceptance", accessibility["note"])

    def test_durable_evidence_is_canonical_search_record(self):
        durable = self.record["durableEvidence"]
        self.assertEqual(
            durable["driveFileId"],
            "1Dk-wqcTIHzjuZNUprlSQGuMtWQx72JC-",
        )
        self.assertEqual(
            durable["title"],
            "GoreeCloud Search — PR 18 Rendered Machine Evidence.docx",
        )

    def test_open_gates_remain_open(self):
        disposition = self.record["disposition"]
        self.assertEqual(disposition["renderedMachineEvidence"], "verified")
        for key, value in disposition.items():
            if key == "renderedMachineEvidence":
                continue
            self.assertEqual(value, "open", key)

    def test_nonclaims_cover_material_boundaries(self):
        text = " ".join(self.record["nonClaims"]).casefold()
        for phrase in (
            "no live brave",
            "no goreecloud-vps-01",
            "no caddy",
            "no owner",
            "no assistive-technology",
            "no search glaze ui consumer acceptance",
            "no deployment",
        ):
            self.assertIn(phrase, text)


if __name__ == "__main__":
    unittest.main()
