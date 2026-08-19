# SPDX-License-Identifier: AGPL-3.0-or-later
from __future__ import annotations

import json
import pathlib
import sys
import tempfile

REPO_ROOT = pathlib.Path(__file__).resolve().parents[2]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from goreecloud.candidate_request import CandidateRequestError, validate_request

BASE = "1bb26fba7a4b1658442e1dd872a0ea170abe141f"
CANDIDATE = "2222222222222222222222222222222222222222"


def valid_request() -> dict[str, object]:
    return {
        "schema_version": 1,
        "product": "GoreeCloud Search",
        "request_id": "first-stable-test",
        "requested_at": "2026-08-19T07:35:00-05:00",
        "request": "publish-and-rehearse-final-candidate",
        "reviewed_base_revision": BASE,
        "production_cutover_authorized": False,
        "stable_release_authorized": False,
        "target_host_change_authorized": False,
        "statement": "Candidate publication and isolated rehearsal only.",
    }


def write_request(value: dict[str, object]) -> pathlib.Path:
    handle = tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", suffix=".json", delete=False)
    with handle:
        json.dump(value, handle)
    return pathlib.Path(handle.name)


def expect_rejected(value: dict[str, object], **kwargs: str) -> None:
    path = write_request(value)
    try:
        try:
            validate_request(path, **kwargs)
        except CandidateRequestError:
            return
        raise AssertionError("Unsafe candidate request was accepted")
    finally:
        path.unlink(missing_ok=True)


def main() -> None:
    path = write_request(valid_request())
    try:
        validated = validate_request(
            path,
            expected_reviewed_base=BASE,
            candidate_source=CANDIDATE,
            candidate_parent=BASE,
        )
        assert validated["product"] == "GoreeCloud Search"
        assert validated["production_cutover_authorized"] is False
        assert validated["stable_release_authorized"] is False
        assert validated["target_host_change_authorized"] is False
    finally:
        path.unlink(missing_ok=True)

    wrong_parent = valid_request()
    expect_rejected(
        wrong_parent,
        candidate_source=CANDIDATE,
        candidate_parent="3333333333333333333333333333333333333333",
    )

    wrong_base = valid_request()
    expect_rejected(
        wrong_base,
        expected_reviewed_base="4444444444444444444444444444444444444444",
    )

    production_authorized = valid_request()
    production_authorized["production_cutover_authorized"] = True
    expect_rejected(production_authorized)

    stable_authorized = valid_request()
    stable_authorized["stable_release_authorized"] = True
    expect_rejected(stable_authorized)

    target_authorized = valid_request()
    target_authorized["target_host_change_authorized"] = True
    expect_rejected(target_authorized)

    sensitive = valid_request()
    sensitive["token"] = "do-not-store-this"
    expect_rejected(sensitive)

    extra = valid_request()
    extra["unexpected"] = False
    expect_rejected(extra)

    same_source = valid_request()
    expect_rejected(same_source, candidate_source=BASE)

    print("GoreeCloud candidate request contract passed.")


if __name__ == "__main__":
    main()
