package deployment

const (
	DEPLOY_OUTCOME_BUILDING = "building"
	DEPLOY_OUTCOME_SUCCEEDED = "succeeded"
	DEPLOY_OUTCOME_FAILED = "failed"
)

func DeriveDeploymentOutcome(step int32) string {
	if step == DEPLOY_STATUS_BUILD_INSTANCE_STARTED {
		return DEPLOY_OUTCOME_SUCCEEDED
	}
	if step < DEPLOY_STATUS_INITIATED {
		return DEPLOY_OUTCOME_FAILED
	}
	if step > DEPLOY_STATUS_BUILD_IMAGE_PUSHED {
		return DEPLOY_OUTCOME_FAILED
	}

	return DEPLOY_OUTCOME_BUILDING
}
