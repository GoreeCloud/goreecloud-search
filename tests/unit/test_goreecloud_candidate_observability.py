# SPDX-License-Identifier: AGPL-3.0-or-later

from pathlib import Path


REPOSITORY_ROOT = Path(__file__).resolve().parents[2]
WORKFLOW = REPOSITORY_ROOT / '.github' / 'workflows' / 'goreecloud-stabilization-candidate.yml'
OBSERVABILITY_DOC = REPOSITORY_ROOT / 'docs' / 'goreecloud' / 'CANDIDATE-OBSERVABILITY.md'


def test_candidate_workflow_has_independent_commit_status_and_pr_receipt():
    workflow = WORKFLOW.read_text(encoding='utf-8')

    assert 'statuses: write' in workflow
    assert 'issues: write' in workflow
    assert 'goreecloud/first-stable-candidate' in workflow
    assert 'statuses/${GITHUB_SHA}' in workflow
    assert 'issues/${FIRST_STABLE_PR}/comments' in workflow
    assert workflow.count('if: ${{ always() }}') >= 4

    status_step = workflow.index('- name: Publish first-Stable candidate commit status')
    receipt_step = workflow.index('- name: Publish sanitized first-Stable candidate outcome receipt')
    redundancy_step = workflow.index('- name: Enforce redundant candidate outcome observability')
    assert status_step < receipt_step < redundancy_step


def test_candidate_status_response_validation_has_persistent_environment():
    workflow = WORKFLOW.read_text(encoding='utf-8')

    assert 'export STATUS_STATE="$state"' in workflow
    assert 'export STATUS_DESCRIPTION="$description"' in workflow
    assert 'export RUN_URL="$run_url"' in workflow
    assert "response.get('state') == os.environ['STATUS_STATE']" in workflow
    assert "response.get('target_url') == os.environ['RUN_URL']" in workflow


def test_candidate_outcome_sinks_are_redundant_and_fail_closed_together():
    workflow = WORKFLOW.read_text(encoding='utf-8')

    assert 'id: candidate-status' in workflow
    assert 'id: candidate-receipt' in workflow
    assert workflow.count('continue-on-error: true') >= 2
    assert 'STATUS_SINK_OUTCOME: ${{ steps.candidate-status.outcome }}' in workflow
    assert 'RECEIPT_SINK_OUTCOME: ${{ steps.candidate-receipt.outcome }}' in workflow
    assert 'Both candidate outcome reporting sinks failed.' in workflow
    assert 'At least one candidate outcome reporting sink succeeded.' in workflow


def test_candidate_outcome_reporting_remains_non_authorizing_and_sanitized():
    workflow = WORKFLOW.read_text(encoding='utf-8')
    observability = OBSERVABILITY_DOC.read_text(encoding='utf-8')

    assert '- Production cutover authorized: **no**' in workflow
    assert '- Stable promotion authorized: **no**' in workflow
    assert 'not a substitute for inspecting the release-evidence artifact' in workflow

    assert 'Production cutover' in observability
    assert 'Stable promotion' in observability
    assert 'registry credentials' in observability
    assert 'user search queries' in observability
    assert 'production configuration' in observability
    assert 'at least one outcome sink' in observability
    assert 'both outcome sinks fail' in observability


if __name__ == '__main__':
    test_candidate_workflow_has_independent_commit_status_and_pr_receipt()
    test_candidate_status_response_validation_has_persistent_environment()
    test_candidate_outcome_sinks_are_redundant_and_fail_closed_together()
    test_candidate_outcome_reporting_remains_non_authorizing_and_sanitized()
    print('Candidate observability contract passed.')
