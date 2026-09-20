# Provider development instructions

This repository is owned by healdropper, not external-org.
Keep the provider generic and independent of all consuming deployments.
Use Terraform Plugin Framework, not Plugin SDK v2.
Work on a dedicated branch and submit a pull request for review.
Use Make targets. Run unit and real disposable acceptance tests for behavior changes.
Prove new API operations with the fixed-document key before adding resources.
Update the canonical document, key migration guidance and tests together.
Never commit API tokens, test state, private signing keys or development overrides.
Keep visibility changes, Registry publication and production consumer adoption as
three separately authorized decisions. Preparing a PR does not authorize merging.
