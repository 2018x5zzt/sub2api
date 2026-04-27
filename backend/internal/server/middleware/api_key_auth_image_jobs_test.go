package middleware

import "testing"

func TestAPIKeyAuthSkipsBillingForOpenAIImageJobPollOnly(t *testing.T) {
	if !isOpenAIImageJobPollRequest("GET", "/v1/images/jobs/imgjob_123") {
		t.Fatal("expected image job polling GET to skip billing enforcement")
	}
	if !isOpenAIImageJobPollRequest("GET", "/api-proxy/v1/images/jobs/imgjob_123") {
		t.Fatal("expected image job polling alias GET to skip billing enforcement")
	}
	if isOpenAIImageJobPollRequest("POST", "/v1/images/jobs/generations") {
		t.Fatal("POST image job submit must not skip billing enforcement")
	}
	if isOpenAIImageJobPollRequest("GET", "/v1/images/generations") {
		t.Fatal("regular images endpoint must not skip billing enforcement")
	}
}
